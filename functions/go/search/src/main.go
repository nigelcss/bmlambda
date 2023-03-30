package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/aws/aws-sdk-go/service/dynamodb/expression"
	"github.com/mmcloughlin/geohash"
)

var ddb *dynamodb.DynamoDB
var tableName *string

func init() {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	}))
	ddb = dynamodb.New(sess)
	tableName = aws.String("geo")

	// warm-up the connection
	ddb.GetItem(&dynamodb.GetItemInput{
		TableName: tableName,
		Key: map[string]*dynamodb.AttributeValue{
			"pk": {S: aws.String("nil")},
			"sk": {S: aws.String("nil")},
		},
	})
}

type QueryItem struct {
	Lat    string `json:"lat"`
	Lon    string `json:"lon"`
	Radius string `json:"radius"`
}

type Item struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Lat   string `json:"lat"`
	Lon   string `json:"lon"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var queryItem QueryItem
	err := json.Unmarshal([]byte(event.Body), &queryItem)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 400}, err
	}
	log.Print(queryItem)

	lat, _ := strconv.ParseFloat(queryItem.Lat, 64)
	lon, _ := strconv.ParseFloat(queryItem.Lon, 64)
	gh := geohash.EncodeWithPrecision(lat, lon, 4)
	nb := geohash.Neighbors(gh)
	matches := append(nb, gh)

	var responseItems []map[string]*dynamodb.AttributeValue
	for _, geohash := range matches {

		keyCond := expression.Key("gpk").Equal(expression.Value(geohash)).And(expression.Key("gsk").BeginsWith("RT:go:"))
		expr, err := expression.NewBuilder().WithKeyCondition(keyCond).Build()
		if err != nil {
			log.Println("Error building expression:", err)
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}

		input := &dynamodb.QueryInput{
			TableName:                 aws.String("geo"),
			IndexName:                 aws.String("geo-index"),
			KeyConditionExpression:    expr.KeyCondition(),
			ExpressionAttributeNames:  expr.Names(),
			ExpressionAttributeValues: expr.Values(),
		}

		response, err2 := ddb.Query(input)
		if err2 != nil {
			log.Fatalf("Got error calling Query: %s", err2)
		}

		responseItems = append(responseItems[:], response.Items[:]...)
	}

	var items []Item
	err = dynamodbattribute.UnmarshalListOfMaps(responseItems, &items)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	body, err := json.Marshal(items)
	if err != nil {
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	log.Print(items)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
