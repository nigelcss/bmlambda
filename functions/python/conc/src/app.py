import json
import boto3
from boto3.dynamodb.conditions import Key
from dataclasses import dataclass
from dataclasses_json import dataclass_json
import json
import geohash
import concurrent.futures
import threading


@dataclass_json
@dataclass
class QueryItem:
    lat: str
    lon: str
    radius: str


@dataclass_json
@dataclass
class Item:
    owner: str
    name: str
    lat: str
    lon: str


# warmup while the CPU is boosted
MAX_THREADS = 9
executor = concurrent.futures.ThreadPoolExecutor(max_workers=MAX_THREADS)
dynamodb = boto3.resource("dynamodb")
table = dynamodb.Table("geo")
try:
    table.get_item(Key={'pk': 'nil', 'sk': 'nil'})
finally:
    print('init done')

def worker(geohash):
    response = table.query(
        IndexName="geo-index",
        KeyConditionExpression=(
            Key('gpk').eq(geohash) & Key('gsk').begins_with('RT:python')
        ),
    )
    return response['Items']


def perform_concurrent_queries(matches):
    futures = [executor.submit(worker, geohash) for geohash in matches]
    results = [future.result() for future in concurrent.futures.as_completed(futures)]
    return results


def lambda_handler(event, context):
    # get the body as an object tree
    query_item = QueryItem.from_json(event["body"])
    print(query_item)

    # find the center and all neighboring geohash's
    gh = geohash.encode(float(query_item.lat), float(query_item.lon), 4)
    matches = geohash.expand(gh)

    results = perform_concurrent_queries(matches)

    items = [Item.from_dict(item) for sublist in results for item in sublist]

    print(items)

    json_dict = [item.to_dict() for item in items]

    return {
        "statusCode": 200,
        "body": json.dumps(json_dict)
    }
