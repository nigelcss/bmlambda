package main

import (
	"context"
	"encoding/json"
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/mmcloughlin/geohash"
)

var ddb *dynamodb.Client
var tableName *string

func init() {
	tableName = aws.String("geo")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-southeast-2"),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ddb = dynamodb.NewFromConfig(cfg)

	// warm-up the connection
	ddb.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: tableName,
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "nil"},
			"sk": &types.AttributeValueMemberS{Value: "nil"},
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

	var items []Item
	for _, geohash := range matches {

		input := &dynamodb.QueryInput{
			TableName:              aws.String("geo"),
			IndexName:              aws.String("geo-index"),
			KeyConditionExpression: aws.String("gpk = :gpk and begins_with(gsk, :gsk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":gpk": &types.AttributeValueMemberS{Value: geohash},
				":gsk": &types.AttributeValueMemberS{Value: "RT:go:"},
			},
		}

		response, err2 := ddb.Query(ctx, input)
		if err2 != nil {
			log.Fatalf("Got error calling Query: %s", err2)
		}

		var responseItems []Item
		err = attributevalue.UnmarshalListOfMaps(response.Items, &responseItems)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500}, err
		}

		items = append(items[:], responseItems[:]...)
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
