package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Item struct {
	Lat    string `json:"lat"`
	Lon    string `json:"lon"`
	Radius string `json:"radius"`
}

func handler(ctx context.Context, event events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	var item Item
	json.Unmarshal([]byte(event.Body), &item)
	log.Print(item)

	body, _ := json.Marshal(item)

	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       string(body),
	}, nil
}

func main() {
	lambda.Start(handler)
}
