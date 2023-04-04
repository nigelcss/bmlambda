package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
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

func (item WriteItem) ToAttributeValueMap() (map[string]types.AttributeValue, error) {
	return map[string]types.AttributeValue{
		"pk":    &types.AttributeValueMemberS{Value: item.Pk},
		"sk":    &types.AttributeValueMemberS{Value: item.Sk},
		"gpk":   &types.AttributeValueMemberS{Value: item.Gpk},
		"gsk":   &types.AttributeValueMemberS{Value: item.Gsk},
		"owner": &types.AttributeValueMemberS{Value: item.Owner},
		"name":  &types.AttributeValueMemberS{Value: item.Name},
		"lat":   &types.AttributeValueMemberS{Value: item.Lat},
		"lon":   &types.AttributeValueMemberS{Value: item.Lon},
	}, nil
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
	log.Println(writeItem)

	av, err := writeItem.ToAttributeValueMap()
	if err != nil {
		log.Fatalf("Got error marshalling item: %s", err)
	}

	log.Println(av)

	input := &dynamodb.PutItemInput{
		TableName: tableName,
		Item:      av,
	}

	_, err2 := ddb.PutItem(ctx, input)
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
