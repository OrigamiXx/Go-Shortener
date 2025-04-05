package storage

import (
	"context"
	"fmt"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

const (
	counterTableName = "url-counter"
	counterKey      = "url_counter"
)

type CounterStorage struct {
	client *dynamodb.Client
}

func NewCounterStorage(client *dynamodb.Client) *CounterStorage {
	return &CounterStorage{
		client: client,
	}
}

// GetNextCounter retrieves and increments the counter atomically
func (s *CounterStorage) GetNextCounter(ctx context.Context) (int64, error) {
	input := &dynamodb.UpdateItemInput{
		TableName: aws.String(counterTableName),
		Key: map[string]types.AttributeValue{
			"CounterName": &types.AttributeValueMemberS{Value: counterKey},
		},
		UpdateExpression: aws.String("SET CurrentValue = if_not_exists(CurrentValue, :zero) + :inc"),
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

	newValue, err := strconv.ParseInt(result.Attributes["CurrentValue"].(*types.AttributeValueMemberN).Value, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("failed to parse counter value: %w", err)
	}

	return newValue, nil
} 