use std::collections::HashMap;
use aws_sdk_dynamodb::model::AttributeValue;
use aws_sdk_dynamodb::Client as DynamoDbClient;
use lambda_runtime::{run, service_fn, LambdaEvent};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use geohash::{encode, neighbors, Coord};
use std::error::Error;
use std::sync::Arc;
use tokio::join;

type Result<T> = std::result::Result<T, Box<dyn Error + Send + Sync + 'static>>;

#[derive(Debug, Deserialize, Serialize)]
struct QueryItem {
    lat: String,
    lon: String,
    radius: String,
}

#[derive(Debug, Deserialize, Serialize)]
struct Item {
    owner: String,
    name: String,
    lat: String,
    lon: String,
}

impl From<HashMap<std::string::String, AttributeValue>> for Item {
    fn from(item: HashMap<std::string::String, AttributeValue>) -> Self {
        Item {
            owner: item["owner"].as_s().unwrap().to_string(),
            name: item["name"].as_s().unwrap().to_string(),
            lat: item["lat"].as_s().unwrap().to_string(),
            lon: item["lon"].as_s().unwrap().to_string(),
        }
    }
}

#[tokio::main]
async fn main() -> Result<()> {
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .with_target(false)
        .without_time()
        .init();

    // warmup
    let config = aws_config::load_from_env().await;
    let dynamodb = Arc::new(DynamoDbClient::new(&config));
    dynamodb
        .get_item()
        .table_name("geo")
        .key("pk", AttributeValue::S("nil".to_string()))
        .key("sk", AttributeValue::S("nil".to_string()))
        .send()
        .await?;

    run(service_fn(|event: LambdaEvent<Value>| function_handler(&dynamodb, event))).await
}

async fn function_handler(dynamodb: &Arc<DynamoDbClient>, event: LambdaEvent<Value>) -> Result<Vec<Item>> {

    let query_item: QueryItem = serde_json::from_str(event.payload["body"].as_str().unwrap())?;    
    println!("{:?}", query_item);

    // find the center and all neighboring geohash's
    let coord = Coord {
        x: query_item.lon.parse::<f64>().unwrap(), 
        y: query_item.lat.parse::<f64>().unwrap()
    };
    let gh = encode(coord, 4usize).expect("Invalid geo coordinates");
    let nb = neighbors(gh.as_str()).expect("Invalid geohash string");
    let all = [gh, nb.sw, nb.s, nb.se, nb.w, nb.e, nb.nw, nb.n, nb.ne];

    let mut handles = Vec::new();

    for geohash in all {
        let query_client = Arc::clone(&dynamodb);
        handles.push(tokio::spawn(async move {
            query(query_client, &geohash).await
        }));
    }

    let mut results = Vec::new();
    for handle in handles {
        let query_results = handle.await??;
        results.extend(query_results);
    }

    println!("Results: {:?}", results);

    Ok(results)
}

async fn query(dynamodb: Arc<DynamoDbClient>, geohash: &String) -> Result<Vec<Item>> {
    let response = dynamodb
        .query()
        .table_name("geo")
        .index_name("geo-index")
        .key_condition_expression("gpk = :geohash and begins_with(gsk, :typeAndOwner)")
        .expression_attribute_values(":geohash", AttributeValue::S(geohash.to_string()))
        .expression_attribute_values(":typeAndOwner", AttributeValue::S("RT:rust:".to_string()))
        .send()
        .await?;

    let items = response
        .items
        .unwrap_or_default()
        .into_iter()
        .map(Item::from)
        .collect();

    Ok(items)
}