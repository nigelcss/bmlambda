import { DynamoDBClient, PutItemCommand, GetItemCommand } from "@aws-sdk/client-dynamodb"
import { marshall } from "@aws-sdk/util-dynamodb"
import ngeohash from 'ngeohash'

const dynamodb = new DynamoDBClient({ region: 'ap-southeast-2' })
const warmupCmd = new GetItemCommand({
    TableName: 'geo',
    Key: { pk: {S: 'nil'}, sk: {S: 'nil'} }    
})
await dynamodb.send(warmupCmd)

export const lambdaHandler = async (event, context) => {
    const item = JSON.parse(event.body)
    console.log(item)

    item.pk = `RT:${item.owner}`
    item.sk = item.name
    item.gpk = ngeohash.encode(parseFloat(item.lat), parseFloat(item.lon), 4)
    item.gsk = `RT:${item.owner}:${item.name}`

    const dynamo_item = marshall(item)

    const command = new PutItemCommand({
        TableName: 'geo',
        Item: dynamo_item
    })

    await dynamodb.send(command)

    try {
        return {
            'statusCode': 200,
        }
    } catch (err) {
        console.log(err)
        return err
    }
};
