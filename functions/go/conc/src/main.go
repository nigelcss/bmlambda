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

var tableName *string
var jobs = make(chan QueryJob)
var results = make(chan []Item, 9)

func init() {
	tableName = aws.String("geo")
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithRegion("ap-southeast-2"),
	)
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	ddb := dynamodb.NewFromConfig(cfg)

	// warm-up the connection
	ddb.GetItem(context.TODO(), &dynamodb.GetItemInput{
		TableName: tableName,
		Key: map[string]types.AttributeValue{
			"pk": &types.AttributeValueMemberS{Value: "nil"},
			"sk": &types.AttributeValueMemberS{Value: "nil"},
		},
	})

	for i := 0; i < 9; i++ {
		go worker(ddb)
	}
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

type QueryJob struct {
	Geohash string
	Ctx     context.Context
	Result  chan []Item
}

func worker(ddb *dynamodb.Client) {

	for job := range jobs {
		input := &dynamodb.QueryInput{
			TableName:              aws.String("geo"),
			IndexName:              aws.String("geo-index"),
			KeyConditionExpression: aws.String("gpk = :gpk and begins_with(gsk, :gsk)"),
			ExpressionAttributeValues: map[string]types.AttributeValue{
				":gpk": &types.AttributeValueMemberS{Value: job.Geohash},
				":gsk": &types.AttributeValueMemberS{Value: "RT:go:"},
			},
		}

		response, err := ddb.Query(job.Ctx, input)
		if err != nil {
			log.Printf("Got error calling Query: %s", err)
			job.Result <- nil
			continue
		}

		var items []Item
		err = attributevalue.UnmarshalListOfMaps(response.Items, &items)
		if err != nil {
			log.Printf("Error unmarshalling items for geohash %s: %s", job.Geohash, err)
			job.Result <- nil
			continue
		}

		job.Result <- items
	}
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

	for _, geohash := range matches {
		job := QueryJob{
			Ctx:     ctx,
			Geohash: geohash,
			Result:  results,
		}
		jobs <- job
	}

	mergedResults := []Item{}
	for range matches {
		items := <-results
		if items != nil {
			mergedResults = append(mergedResults, items...)
		}
	}

	jsonResults, err := json.Marshal(mergedResults)
	if err != nil {
		log.Printf("Error marshalling results: %s", err)
		return events.APIGatewayProxyResponse{StatusCode: 500}, err
	}

	log.Print(mergedResults)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(jsonResults),
	}, nil
}

func main() {
	lambda.Start(handler)
}
