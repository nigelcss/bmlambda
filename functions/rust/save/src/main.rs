use aws_sdk_dynamodb::types::AttributeValue;
use geohash::{encode, Coord};
use lambda_runtime::{run, service_fn, Error, LambdaEvent};
use serde::{Deserialize, Serialize};
use serde_dynamo::to_item;
use serde_json::Value;

#[derive(Debug, Deserialize, Serialize)]
struct Item {
    pk: Option<String>,
    sk: Option<String>,
    gpk: Option<String>,
    gsk: Option<String>,
    owner: String,
    name: String,
    lat: String,
    lon: String,
}

#[derive(Debug, Deserialize, Serialize)]
struct Response {
    status_code: u32,
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

    run(service_fn(|event: LambdaEvent<Value>| {
        function_handler(&dynamodb, event)
    }))
    .await
}

async fn function_handler(
    dynamodb: &aws_sdk_dynamodb::Client,
    event: LambdaEvent<Value>,
) -> Result<Response, Error> {
    let mut item: Item = serde_json::from_str(event.payload["body"].as_str().unwrap())?;
    println!("{:?}", item);

    item.pk = Some(format!("RT:{}", item.owner));
    item.sk = Some(item.name.clone());
    item.gpk = Some(encode(
        Coord {
            x: item.lon.parse::<f64>().unwrap(),
            y: item.lat.parse::<f64>().unwrap(),
        },
        4usize,
    )?);
    item.gsk = Some(format!("RT:{}:{}", item.owner, item.name));

    let put_item = to_item(item)?;

    dynamodb
        .put_item()
        .table_name("geo")
        .set_item(Some(put_item))
        .send()
        .await?;

    Ok(Response { status_code: 200 })
}
