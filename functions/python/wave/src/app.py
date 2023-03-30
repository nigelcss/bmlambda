import boto3
from dataclasses import dataclass
from dataclasses_json import dataclass_json
import json

@dataclass_json
@dataclass
class Item:
    lat : str
    lon : str
    radius : str

def lambda_handler(event, context):
    # get the body as an object tree
    item = Item.from_json(event["body"])    
    print(item)

    return {
        "statusCode": 200,
        "body": item.to_json()
     }
