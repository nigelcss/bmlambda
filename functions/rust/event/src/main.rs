use aws_sdk_eventbridge::model::PutEventsRequestEntry;
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
struct PutItem {
    pk: String,
    sk: String,
    gpk: String,
    gsk: String,
    owner: String,
    name: String,
    lat: String,
    lon: String,
}

impl PutItem {
    pub fn from(pk: String, sk: String, gpk: String, gsk: String, item: Item) -> Self {
        PutItem {
            pk: pk,
            sk: sk,
            gpk: gpk,
            gsk: gsk,
            owner: item.owner,
            name: item.name,
            lat: item.lat,
            lon: item.lon,        
        }
    }
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
    let eventbridge = aws_sdk_eventbridge::Client::new(&config);

    run(service_fn(|event: LambdaEvent<Value>| function_handler(&eventbridge, event))).await
}

async fn function_handler(eventbridge: &aws_sdk_eventbridge::Client, event: LambdaEvent<Value>) -> Result<Response, Error> {

    let event_bus_name = std::env::var("EVENT_BUS_NAME").expect("EVENT_BUS_NAME must be set");

    let item: Item = serde_json::from_str(event.payload["body"].as_str().unwrap())?;
    println!("{:?}", item);

    let put_item = PutItem::from(
        format!("RT:{}", item.owner),
        item.name.to_string(),
        encode(Coord {x: item.lon.parse::<f64>().unwrap(), y: item.lat.parse::<f64>().unwrap()}, 4usize)?,
        format!("RT:{}:{}", item.owner, item.name),
        item
    );

    let params = PutEventsRequestEntry::builder()
        .event_bus_name(event_bus_name)
        .source("rust.event")
        .detail_type("Item")
        .detail(serde_json::to_string(&put_item)?)
        .build();

    let result = eventbridge.put_events().entries(params).send().await?;

    println!("{:?}", result);

    Ok(Response {
        statusCode: 200
    })
}
