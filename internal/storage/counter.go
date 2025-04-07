package storage

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	counterTableName = "url-counter"
	// Format: YYYY-MM-DD
	dateFormat = "2006-01-02"
)

// getBucketKey returns a string key for the current day bucket
func getBucketKey() string {
	return time.Now().UTC().Format(dateFormat)
}

type CounterStorage struct {
	client *dynamodb.Client
}

func NewCounterStorage(client *dynamodb.Client) *CounterStorage {
	return &CounterStorage{
		client: client,
	}
}

// GetNextCounter retrieves and increments the counter atomically within the current day bucket
func (s *CounterStorage) GetNextCounter(ctx context.Context) (int64, error) {
	bucketKey := getBucketKey()
	
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(counterTableName),
		Key: map[string]types.AttributeValue{
			"BucketKey": &types.AttributeValueMemberS{Value: bucketKey},
		},
		UpdateExpression: aws.String("SET CounterValue = if_not_exists(CounterValue, :zero) + :inc"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":zero": &types.AttributeValueMemberN{Value: "0"},
			":inc":  &types.AttributeValueMemberN{Value: "1"},
		},
		ReturnValues: types.ReturnValueAllNew,
	}

	result, err := s.client.UpdateItem(ctx, input)
	if err != nil {
		return 0, fmt.Errorf("failed to update counter: %w", err)
	}

	newValue, err := strconv.ParseInt(result.Attributes["CounterValue"].(*types.AttributeValueMemberN).Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse counter value: %w", err)
	}

	return newValue, nil
}

// CleanupOldBuckets removes counter entries older than the specified number of days
func (s *CounterStorage) CleanupOldBuckets(ctx context.Context, daysToKeep int) error {
	cutoffDate := time.Now().UTC().AddDate(0, 0, -daysToKeep)
	
	// List all items in the table
	input := &dynamodb.ScanInput{
		TableName: aws.String(counterTableName),
	}
	
	result, err := s.client.Scan(ctx, input)
	if err != nil {
		return fmt.Errorf("failed to scan counter table: %w", err)
	}
	
	// Delete old buckets
	for _, item := range result.Items {
		bucketKey := item["BucketKey"].(*types.AttributeValueMemberS).Value
		bucketDate, err := time.Parse(dateFormat, bucketKey)
		if err != nil {
			continue // Skip items that don't match our date format
		}
		
		if bucketDate.Before(cutoffDate) {
			deleteInput := &dynamodb.DeleteItemInput{
				TableName: aws.String(counterTableName),
				Key: map[string]types.AttributeValue{
					"BucketKey": &types.AttributeValueMemberS{Value: bucketKey},
				},
			}
			
			_, err := s.client.DeleteItem(ctx, deleteInput)
			if err != nil {
				return fmt.Errorf("failed to delete old bucket %s: %w", bucketKey, err)
			}
		}
	}
	
	return nil
} 