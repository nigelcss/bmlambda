use aws_sdk_dynamodb::types::AttributeValue;
use lambda_runtime::{run, service_fn, Error, LambdaEvent};
use serde_dynamo;
use serde::{Deserialize, Serialize};
use serde_json::Value;
use geohash::{encode, neighbors, Coord};

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

#[derive(Debug, Deserialize, Serialize)]
struct Response {
    statusCode: u32,
    body: String,
}

#[tokio::main]
async fn main() -> Result<(), Error> {
    tracing_subscriber::fmt()
        .with_max_level(tracing::Level::INFO)
        .with_target(false)
        .without_time()
        .init();

    // warmup
    let config = aws_config::load_from_env().await;
    let dynamodb = aws_sdk_dynamodb::Client::new(&config);
    dynamodb
        .get_item()
        .table_name("geo")
        .key("pk", AttributeValue::S("nil".to_string()))
        .key("sk", AttributeValue::S("nil".to_string()))
        .send()
        .await?;

    run(service_fn(|event: LambdaEvent<Value>| function_handler(&dynamodb, event))).await
}

async fn function_handler(dynamodb: &aws_sdk_dynamodb::Client, event: LambdaEvent<Value>) -> Result<Response, Error> {

    let query_item: QueryItem = serde_json::from_str(event.payload["body"].as_str().unwrap())?;    
    println!("{:?}", query_item);

    // find the center and all neighboring geohash's
    let coord = Coord {
        x: query_item.lon.parse::<f64>().unwrap(), 
        y: query_item.lat.parse::<f64>().unwrap()
    };
    let gh = encode(coord, 4usize).expect("Invalid geo coordinates");
    let nb = neighbors(gh.as_str()).expect("Invalid geohash string");

    let mut response_items: Vec<Item> = Vec::new();
    for geohash in [gh, nb.sw, nb.s, nb.se, nb.w, nb.e, nb.nw, nb.n, nb.ne] {
        let req = dynamodb
            .query()
            .table_name("geo")
            .index_name("geo-index")
            .key_condition_expression("gpk = :geohash and begins_with(gsk, :typeAndOwner)")
            .expression_attribute_values(":geohash", AttributeValue::S(geohash.to_string()))
            .expression_attribute_values(":typeAndOwner", AttributeValue::S("RT:rust:".to_string()))
            .send()
            .await?;

        if let Some(items) = req.items {
            let matched_items: Vec<Item> = serde_dynamo::from_items(items)?;
            response_items.extend(matched_items);
        }
    } 
    println!("{:?}", response_items);

    Ok(Response { 
        statusCode: 200,
        body: serde_json::to_string(&response_items)?,
    })
}
