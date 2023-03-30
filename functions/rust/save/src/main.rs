use aws_sdk_dynamodb::model::AttributeValue;
use lambda_runtime::{run, service_fn, Error, LambdaEvent};
use serde::{Deserialize, Serialize};
use serde_json::Value;
use geohash::{encode, Coord};

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

    let item: Item = serde_json::from_str(event.payload["body"].as_str().unwrap())?;
    println!("{:?}", item);

    let pk = format!("RT:{}", item.owner);
    let gpk = encode(Coord {x: item.lon.parse::<f64>().unwrap(), y: item.lat.parse::<f64>().unwrap()}, 4usize)?;
    let gsk = format!("RT:{}:{}", item.owner, item.name);

    dynamodb
        .put_item()
        .table_name("geo")
        .item("pk", AttributeValue::S(pk))
        .item("sk", AttributeValue::S(item.name.to_string()))
        .item("gpk", AttributeValue::S(gpk))
        .item("gsk", AttributeValue::S(gsk))
        .item("owner", AttributeValue::S(item.owner))
        .item("name", AttributeValue::S(item.name))
        .item("lat", AttributeValue::S(item.lat))
        .item("lon", AttributeValue::S(item.lon))
        .send()
        .await?;

    Ok(Response {
        statusCode: 200
    })
}
