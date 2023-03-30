import boto3
from dataclasses import dataclass
from dataclasses_json import dataclass_json
import json
import geohash

@dataclass_json
@dataclass
class Item:
    owner : str
    name : str
    lat : str
    lon : str

@dataclass_json
@dataclass
class WriteItem:
    pk : str
    sk : str
    gpk : str
    gsk : str
    owner : str
    name : str
    lat : str
    lon : str

# warmup while the CPU is boosted
dynamodb = boto3.resource('dynamodb')
geo_table = dynamodb.Table('geo')
try:
    geo_table.get_item(Key={'pk': 'nil', 'sk': 'nil'})
finally:
    print('init done')

def lambda_handler(event, context):
    # get the body as an object tree
    item = Item.from_json(event["body"])
    print(item)

    # add the table and index keys to the body
    write_item = WriteItem(
        'RT:{0}'.format(item.owner),
        item.name,
        geohash.encode(float(item.lat), float(item.lon), 4),
        'RT:{0}:{1}'.format(item.owner, item.name),
        item.owner,
        item.name,
        item.lat,
        item.lon,
    )

    # write to the dynamodb table
    try:
        geo_table.put_item(Item=write_item.to_dict())
        status_code = 200
    except Exception as e:
        print(e)
        status_code = 500

    return {
        "statusCode": status_code,
     }
