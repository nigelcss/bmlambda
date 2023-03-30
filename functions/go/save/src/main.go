package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
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

type Item struct {
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Lat   string `json:"lat"`
	Lon   string `json:"lon"`
}

type WriteItem struct {
	Pk    string `json:"pk"`
	Sk    string `json:"sk"`
	Gpk   string `json:"gpk"`
	Gsk   string `json:"gsk"`
	Owner string `json:"owner"`
	Name  string `json:"name"`
	Lat   string `json:"lat"`
	Lon   string `json:"lon"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var item Item
	json.Unmarshal([]byte(event.Body), &item)
	log.Print(item)

	lat, _ := strconv.ParseFloat(item.Lat, 64)
	lon, _ := strconv.ParseFloat(item.Lon, 64)

	writeItem := WriteItem{
		Pk:    fmt.Sprintf("RT:%s", item.Owner),
		Sk:    item.Name,
		Gpk:   geohash.EncodeWithPrecision(lat, lon, 4),
		Gsk:   fmt.Sprintf("RT:%s:%s", item.Owner, item.Name),
		Owner: item.Owner,
		Name:  item.Name,
		Lat:   item.Lat,
		Lon:   item.Lon,
	}

	av, err := dynamodbattribute.MarshalMap(writeItem)
	if err != nil {
		log.Fatalf("Got error marshalling item: %s", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: tableName,
	}

	_, err2 := ddb.PutItem(input)
	if err2 != nil {
		log.Fatalf("Got error calling PutItem: %s", err2)
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
	}, nil
}

func main() {
	lambda.Start(handler)
}
