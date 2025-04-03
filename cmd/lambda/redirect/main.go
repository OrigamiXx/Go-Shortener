package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jingy/Go-Shortener/internal/models"
	"github.com/jingy/Go-Shortener/internal/storage"
)

var urlStorage *storage.DynamoDBStorage

func init() {
	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	// Initialize DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)
	urlStorage = storage.NewDynamoDBStorage(dynamoClient)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Extract short code from path
	shortCode := request.PathParameters["shortCode"]
	if shortCode == "" {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Missing short code"}`,
		}, nil
	}

	// Get URL from storage
	url, err := urlStorage.Get(ctx, shortCode)
	if err != nil {
		if err == models.ErrURLNotFound {
			return events.APIGatewayProxyResponse{
				StatusCode: 404,
				Body:       `{"error": "URL not found"}`,
			}, nil
		}
		if err == models.ErrURLExpired {
			return events.APIGatewayProxyResponse{
				StatusCode: 410,
				Body:       `{"error": "URL has expired"}`,
			}, nil
		}
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to retrieve URL"}`,
		}, nil
	}

	// Return redirect response
	return events.APIGatewayProxyResponse{
		StatusCode: 302,
		Headers: map[string]string{
			"Location": url.OriginalURL,
		},
		Body: "",
	}, nil
}

func main() {
	lambda.Start(handleRequest)
} 