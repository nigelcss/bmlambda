from dataclasses import dataclass
from dataclasses_json import dataclass_json
import boto3
from boto3.dynamodb.conditions import Key
import json
import geohash


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
dynamodb = boto3.resource('dynamodb')
geo_table = dynamodb.Table('geo')
try:
    geo_table.get_item(Key={'pk': 'nil', 'sk': 'nil'})
finally:
    print('init done')


def query(geohash):
    response = geo_table.query(
        IndexName='geo-index',
        KeyConditionExpression=(
            Key('gpk').eq(geohash) & Key('gsk').begins_with('RT:python')
        ),
    )
    return [Item.from_dict(item) for item in response['Items']]


def lambda_handler(event, context):
    # get the body as an object tree
    query_item = QueryItem.from_json(event["body"])
    print(query_item)

    # find the center and all neighboring geohash's
    gh = geohash.encode(float(query_item.lat), float(query_item.lon), 4)
    matches = geohash.expand(gh)

    # load any items from dynamodb with a matching geohash
    results = [query(geohash) for geohash in matches]

    items = [item.to_dict() for sublist in results for item in sublist]

    print(items)

    return {
        "statusCode": 200,
        "body": json.dumps(items)
    }
