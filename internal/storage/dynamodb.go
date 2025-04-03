package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jingy/Go-Shortener/internal/models"
)

const (
	tableName = "url-shortener"
)

type DynamoDBStorage struct {
	client *dynamodb.Client
}

func NewDynamoDBStorage(client *dynamodb.Client) *DynamoDBStorage {
	return &DynamoDBStorage{
		client: client,
	}
}

func (s *DynamoDBStorage) Create(ctx context.Context, url *models.URL) error {
	av, err := attributevalue.MarshalMap(url)
	if err != nil {
		return fmt.Errorf("failed to marshal URL: %w", err)
	}

	input := &dynamodb.PutItemInput{
		Item:      av,
		TableName: aws.String(tableName),
		ConditionExpression: aws.String("attribute_not_exists(ShortCode)"),
	}

	_, err = s.client.PutItem(ctx, input)
	if err != nil {
		var condErr *types.ConditionalCheckFailedException
		if errors.As(err, &condErr) {
			return models.ErrDuplicateShortCode
		}
		return fmt.Errorf("failed to put item: %w", err)
	}

	return nil
}

func (s *DynamoDBStorage) Get(ctx context.Context, shortCode string) (*models.URL, error) {
	input := &dynamodb.GetItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ShortCode": &types.AttributeValueMemberS{Value: shortCode},
		},
	}

	result, err := s.client.GetItem(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("failed to get item: %w", err)
	}

	if result.Item == nil {
		return nil, models.ErrURLNotFound
	}

	var url models.URL
	err = attributevalue.UnmarshalMap(result.Item, &url)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal URL: %w", err)
	}

	if !url.ExpiresAt.IsZero() && time.Now().After(url.ExpiresAt) {
		return nil, models.ErrURLExpired
	}

	return &url, nil
}

func (s *DynamoDBStorage) Delete(ctx context.Context, shortCode string) error {
	input := &dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key: map[string]types.AttributeValue{
			"ShortCode": &types.AttributeValueMemberS{Value: shortCode},
		},
	}

	_, err := s.client.DeleteItem(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	return nil
} 