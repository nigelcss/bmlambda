use aws_sdk_eventbridge::model::PutEventsRequestEntry;
use geohash::{encode, Coord};
use lambda_runtime::{run, service_fn, Error, LambdaEvent};
use serde::{Deserialize, Serialize};
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
    let eventbridge = aws_sdk_eventbridge::Client::new(&config);

    run(service_fn(|event: LambdaEvent<Value>| {
        function_handler(&eventbridge, event)
    }))
    .await
}

async fn function_handler(
    eventbridge: &aws_sdk_eventbridge::Client,
    event: LambdaEvent<Value>,
) -> Result<Response, Error> {
    let event_bus_name = std::env::var("EVENT_BUS_NAME").expect("EVENT_BUS_NAME must be set");

    let mut item: Item = serde_json::from_str(event.payload["body"].as_str().unwrap())?;
    println!("{:?}", item);

    item.pk = Some(format!("RT:{}", item.owner));
    item.sk = Some(item.name.to_string());
    item.gpk = Some(encode(
        Coord {
            x: item.lon.parse::<f64>().unwrap(),
            y: item.lat.parse::<f64>().unwrap(),
        },
        4usize,
    )?);
    item.gsk = Some(format!("RT:{}:{}", item.owner, item.name));

    let params = PutEventsRequestEntry::builder()
        .event_bus_name(event_bus_name)
        .source("rust.event")
        .detail_type("Item")
        .detail(serde_json::to_string(&item)?)
        .build();

    let result = eventbridge.put_events().entries(params).send().await?;

    println!("{:?}", result);

    Ok(Response { status_code: 200 })
}
