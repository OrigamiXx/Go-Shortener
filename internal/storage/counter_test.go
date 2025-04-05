package storage

import (
	"context"
	"testing"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// MockDynamoDBClient is a mock implementation of the DynamoDB client
type MockDynamoDBClient struct {
	UpdateItemFunc func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return m.UpdateItemFunc(ctx, params, optFns...)
}

func TestCounterStorage_GetNextCounter(t *testing.T) {
	tests := []struct {
		name           string
		mockUpdateItem func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
		expectedValue  int64
		expectError    bool
	}{
		{
			name: "successful increment",
			mockUpdateItem: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
				return &dynamodb.UpdateItemOutput{
					Attributes: map[string]types.AttributeValue{
						"CurrentValue": &types.AttributeValueMemberN{Value: "42"},
					},
				}, nil
			},
			expectedValue: 42,
			expectError:   false,
		},
		{
			name: "first increment",
			mockUpdateItem: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
				return &dynamodb.UpdateItemOutput{
					Attributes: map[string]types.AttributeValue{
						"CurrentValue": &types.AttributeValueMemberN{Value: "1"},
					},
				}, nil
			},
			expectedValue: 1,
			expectError:   false,
		},
		{
			name: "error from dynamodb",
			mockUpdateItem: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
				return nil, aws.NewError("InternalServerError")
			},
			expectedValue: 0,
			expectError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &MockDynamoDBClient{
				UpdateItemFunc: tt.mockUpdateItem,
			}

			// Create counter storage with mock client
			counterStorage := NewCounterStorage(mockClient)

			// Call GetNextCounter
			value, err := counterStorage.GetNextCounter(context.Background())

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("GetNextCounter() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Check value if no error expected
			if !tt.expectError && value != tt.expectedValue {
				t.Errorf("GetNextCounter() value = %v, expected %v", value, tt.expectedValue)
			}
		})
	}
}

func TestCounterStorage_GetNextCounter_Concurrent(t *testing.T) {
	// This test verifies that the counter is thread-safe by simulating concurrent access
	
	// Create a counter that returns sequential values
	currentValue := int64(0)
	mockClient := &MockDynamoDBClient{
		UpdateItemFunc: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			// Simulate atomic increment
			currentValue++
			return &dynamodb.UpdateItemOutput{
				Attributes: map[string]types.AttributeValue{
					"CurrentValue": &types.AttributeValueMemberN{Value: aws.String(currentValue)},
				},
			}, nil
		},
	}

	// Create counter storage with mock client
	counterStorage := NewCounterStorage(mockClient)

	// Number of concurrent goroutines
	numGoroutines := 10
	// Number of increments per goroutine
	incrementsPerGoroutine := 100

	// Channel to collect results
	results := make(chan int64, numGoroutines*incrementsPerGoroutine)

	// Launch goroutines
	for i := 0; i < numGoroutines; i++ {
		go func() {
			for j := 0; j < incrementsPerGoroutine; j++ {
				value, err := counterStorage.GetNextCounter(context.Background())
				if err != nil {
					t.Errorf("GetNextCounter() error = %v", err)
					return
				}
				results <- value
			}
		}()
	}

	// Collect results
	values := make(map[int64]bool)
	for i := 0; i < numGoroutines*incrementsPerGoroutine; i++ {
		value := <-results
		values[value] = true
	}

	// Check that we got the expected number of unique values
	expectedCount := numGoroutines * incrementsPerGoroutine
	if len(values) != expectedCount {
		t.Errorf("Expected %d unique values, got %d", expectedCount, len(values))
	}

	// Check that all values are in the expected range
	for value := range values {
		if value < 1 || value > int64(expectedCount) {
			t.Errorf("Value %d is outside the expected range [1, %d]", value, expectedCount)
		}
	}
} 