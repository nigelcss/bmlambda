import {
  DynamoDBClient,
  QueryCommand,
  GetItemCommand,
} from "@aws-sdk/client-dynamodb";
import { unmarshall } from "@aws-sdk/util-dynamodb";
import ngeohash from "ngeohash";

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

  const items = [];
  for (let igh in matches) {
    const geohash = matches[igh];

    const command = new QueryCommand({
      ExpressionAttributeValues: {
        ":gpk": { S: geohash },
        ":gsk": { S: 'RT:node' },
      },
      KeyConditionExpression: 'gpk = :gpk and begins_with(gsk, :gsk)',
      TableName: 'geo',
      IndexName: 'geo-index',
    });

    const response = await dynamodb.send(command)

    for (let item in response.Items) {
      items.push(unmarshall(response.Items[item]));
    }
  }

  try {
    return {
      statusCode: 200,
      body: items,
    };
  } catch (err) {
    console.log(err);
    return err;
  }
};
