import asyncio
import json
import aioboto3
from boto3.dynamodb.conditions import Key
from dataclasses import dataclass
from dataclasses_json import dataclass_json
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


async def query(table, geohash):
    response = await table.query(
        IndexName="geo-index",
        KeyConditionExpression=(
            Key('gpk').eq(geohash) & Key('gsk').begins_with('RT:python')
        ),
    )

    return [Item.from_dict(item) for item in response['Items']]


async def perform_concurrent_queries(matches):
    async with aioboto3.Session().resource("dynamodb") as ddb:
        table = await ddb.Table('geo')
        tasks = [query(table, geohash) for geohash in matches]
        results = await asyncio.gather(*tasks)
        return results


def lambda_handler(event, context):
    # get the body as an object tree
    query_item = QueryItem.from_json(event["body"])
    print(query_item)

    # find the center and all neighboring geohash's
    gh = geohash.encode(float(query_item.lat), float(query_item.lon), 4)
    matches = geohash.expand(gh)

    loop = asyncio.get_event_loop()
    results = loop.run_until_complete(perform_concurrent_queries(matches))

    items = [item.to_dict() for sublist in results for item in sublist]

    print(items)

    return {
        "statusCode": 200,
        "body": json.dumps(items)
    }
