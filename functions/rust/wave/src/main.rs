use lambda_runtime::{run, service_fn, Error, LambdaEvent};
use serde::{Deserialize, Serialize};
use serde_json::Value;

#[derive(Debug, Deserialize, Serialize)]
struct Item {
    lat: String,
    lon: String,
    radius: String
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

    run(service_fn(|event: LambdaEvent<Value>| function_handler(event))).await
}

async fn function_handler(event: LambdaEvent<Value>) -> Result<Response, Error> {

    let item: Item = serde_json::from_str(event.payload["body"].as_str().unwrap())?;    
    println!("{:?}", item);

    Ok(Response { 
        statusCode: 200,
        body: serde_json::to_string(&item)?,
    })
}
