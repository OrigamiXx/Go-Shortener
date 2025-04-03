package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/jingy/Go-Shortener/internal/models"
	"github.com/jingy/Go-Shortener/internal/storage"
	"github.com/jingy/Go-Shortener/pkg/shortener"
)

var (
	shortenerService *shortener.Shortener
	urlStorage      *storage.DynamoDBStorage
)

func init() {
	// Initialize AWS config
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		panic(fmt.Sprintf("unable to load SDK config: %v", err))
	}

	// Initialize DynamoDB client
	dynamoClient := dynamodb.NewFromConfig(cfg)
	urlStorage = storage.NewDynamoDBStorage(dynamoClient)

	// Initialize shortener service
	baseURL := os.Getenv("BASE_URL")
	if baseURL == "" {
		baseURL = "https://your-domain.com" // Replace with your actual domain
	}
	shortenerService = shortener.NewShortener(baseURL)
}

func handleRequest(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	// Parse request body
	var req models.CreateURLRequest
	if err := json.Unmarshal([]byte(request.Body), &req); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       `{"error": "Invalid request body"}`,
		}, nil
	}

	// Create short URL
	url, err := shortenerService.CreateShortURL(req.URL)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 400,
			Body:       fmt.Sprintf(`{"error": "%v"}`, err),
		}, nil
	}

	// Store in DynamoDB
	if err := urlStorage.Create(ctx, url); err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to create short URL"}`,
		}, nil
	}

	// Prepare response
	response := models.CreateURLResponse{
		ShortCode: url.ShortCode,
		ShortURL:  url.ShortURL,
	}

	responseBody, err := json.Marshal(response)
	if err != nil {
		return events.APIGatewayProxyResponse{
			StatusCode: 500,
			Body:       `{"error": "Failed to generate response"}`,
		}, nil
	}

	return events.APIGatewayProxyResponse{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(responseBody),
	}, nil
}

func main() {
	lambda.Start(handleRequest)
} 