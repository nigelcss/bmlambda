import json
from aiodynamo.client import Client
from aiodynamo.credentials import Credentials
from aiodynamo.http.aiohttp import AIOHTTP
from aiodynamo.expressions import (HashKey, RangeKey)
from aiohttp import ClientSession
from dataclasses import dataclass
from dataclasses_json import dataclass_json
import json
import geohash
import concurrent.futures
import asyncio


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
loop = asyncio.get_event_loop()
session = ClientSession()
client = Client(AIOHTTP(session), Credentials.auto(), "ap-southeast-2")
table = client.table("geo")
#dynamodb = boto3.resource("dynamodb")
#table = dynamodb.Table("geo")
#try:
    #table.get_item(Key={'pk': 'nil', 'sk': 'nil'})
#finally:
    #print('init done')

async def worker(geohash):
    response = table.query(
        index="geo-index",
        key_condition=(
            HashKey('gpk', geohash) & RangeKey('gsk').begins_with('RT:python')
        ),
    )

    results = []
    async for item in response:
        results.append(item)
    return results


def lambda_handler(event, context):
    # get the body as an object tree
    query_item = QueryItem.from_json(event["body"])
    print(query_item)

    # find the center and all neighboring geohash's
    gh = geohash.encode(float(query_item.lat), float(query_item.lon), 4)
    matches = geohash.expand(gh)

    results = loop.run_until_complete(
        asyncio.gather(*(worker(geohash) for geohash in matches))
    )

    items = [Item.from_dict(item) for sublist in results for item in sublist]

    print(items)

    json_dict = [item.to_dict() for item in items]

    return {
        "statusCode": 200,
        "body": json.dumps(json_dict)
    }
