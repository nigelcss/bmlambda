import {
  DynamoDBClient,
  QueryCommand,
  GetItemCommand,
} from "@aws-sdk/client-dynamodb";
import { unmarshall } from "@aws-sdk/util-dynamodb";
import ngeohash from "ngeohash";

class Item {
  constructor({ owner, name, lat, lon }) {
    this.owner = owner;
    this.name = name;
    this.lat = lat;
    this.lon = lon;
  }
}

const dynamodb = new DynamoDBClient({ region: "ap-southeast-2" });
const warmupCmd = new GetItemCommand({
  TableName: "geo",
  Key: { pk: { S: "nil" }, sk: { S: "nil" } },
});
await dynamodb.send(warmupCmd);

export const lambdaHandler = async (event, context) => {
  const query_item = JSON.parse(event.body);
  console.log(query_item);

  const gh = ngeohash.encode(
    parseFloat(query_item.lat),
    parseFloat(query_item.lon),
    4
  );
  const matches = ngeohash.neighbors(gh);
  matches.push(gh);

  console.log(`matches: ${matches}`);

  const promises = matches.map((geohash) => query(geohash));

  try {
    const results = await Promise.all(promises);
    const mergedResults = [].concat(...results);
    return {
      statusCode: 200,
      body: JSON.stringify(mergedResults),
    };
  } catch (error) {
    console.error(error);
    return {
      statusCode: 500,
      body: "Error performing parallel DynamoDB queries",
    };
  }
};

async function query(geohash) {
  const command = new QueryCommand({
    ExpressionAttributeValues: {
      ":gpk": { S: geohash },
      ":gsk": { S: "RT:node" },
    },
    KeyConditionExpression: "gpk = :gpk and begins_with(gsk, :gsk)",
    TableName: "geo",
    IndexName: "geo-index",
  });

  try {
    const data = await dynamodb.send(command);
    return data.Items.map((item) => new Item(item));
  } catch (error) {
    console.error(`Error querying for geohash ${geohash}:`, error);
    throw error;
  }
}
