package storage

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

// MockDynamoDBClient is a mock implementation of the DynamoDB client
type MockDynamoDBClient struct {
	UpdateItemFunc func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
	ScanFunc       func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
	DeleteItemFunc func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	return m.UpdateItemFunc(ctx, params, optFns...)
}

func (m *MockDynamoDBClient) Scan(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
	return m.ScanFunc(ctx, params, optFns...)
}

func (m *MockDynamoDBClient) DeleteItem(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
	return m.DeleteItemFunc(ctx, params, optFns...)
}

func TestCounterStorage_GetNextCounter(t *testing.T) {
	// Get current bucket key
	currentBucket := time.Now().UTC().Format("2006-01-02")
	
	tests := []struct {
		name           string
		mockUpdateItem func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error)
		expectedValue  int64
		expectError    bool
	}{
		{
			name: "successful increment",
			mockUpdateItem: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
				// Verify the bucket key is correct
				if params.Key["BucketKey"].(*types.AttributeValueMemberS).Value != currentBucket {
					t.Errorf("Expected bucket key %s, got %s", currentBucket, params.Key["BucketKey"].(*types.AttributeValueMemberS).Value)
				}
				
				return &dynamodb.UpdateItemOutput{
					Attributes: map[string]types.AttributeValue{
						"CounterValue": &types.AttributeValueMemberN{Value: "42"},
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
						"CounterValue": &types.AttributeValueMemberN{Value: "1"},
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
	currentBucket := time.Now().UTC().Format("2006-01-02")
	counters := make(map[string]int64)
	
	mockClient := &MockDynamoDBClient{
		UpdateItemFunc: func(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
			bucketKey := params.Key["BucketKey"].(*types.AttributeValueMemberS).Value
			
			// Verify we're using the current bucket
			if bucketKey != currentBucket {
				t.Errorf("Expected bucket key %s, got %s", currentBucket, bucketKey)
			}
			
			// Simulate atomic increment
			counters[bucketKey]++
			return &dynamodb.UpdateItemOutput{
				Attributes: map[string]types.AttributeValue{
					"CounterValue": &types.AttributeValueMemberN{Value: aws.String(counters[bucketKey])},
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

func TestCounterStorage_CleanupOldBuckets(t *testing.T) {
	tests := []struct {
		name           string
		daysToKeep     int
		mockScan       func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error)
		mockDeleteItem func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error)
		expectError    bool
	}{
		{
			name:       "successful cleanup",
			daysToKeep: 7,
			mockScan: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				// Create some test items with different dates
				now := time.Now().UTC()
				oldDate := now.AddDate(0, 0, -10).Format("2006-01-02")
				recentDate := now.AddDate(0, 0, -5).Format("2006-01-02")
				today := now.Format("2006-01-02")
				
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{
							"BucketKey": &types.AttributeValueMemberS{Value: oldDate},
						},
						{
							"BucketKey": &types.AttributeValueMemberS{Value: recentDate},
						},
						{
							"BucketKey": &types.AttributeValueMemberS{Value: today},
						},
					},
				}, nil
			},
			mockDeleteItem: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				// Verify we're only deleting old buckets
				bucketKey := params.Key["BucketKey"].(*types.AttributeValueMemberS).Value
				bucketDate, _ := time.Parse("2006-01-02", bucketKey)
				cutoffDate := time.Now().UTC().AddDate(0, 0, -7)
				
				if !bucketDate.Before(cutoffDate) {
					t.Errorf("Attempting to delete recent bucket: %s", bucketKey)
				}
				
				return &dynamodb.DeleteItemOutput{}, nil
			},
			expectError: false,
		},
		{
			name:       "scan error",
			daysToKeep: 7,
			mockScan: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				return nil, aws.NewError("InternalServerError")
			},
			mockDeleteItem: nil,
			expectError:    true,
		},
		{
			name:       "delete error",
			daysToKeep: 7,
			mockScan: func(ctx context.Context, params *dynamodb.ScanInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ScanOutput, error) {
				now := time.Now().UTC()
				oldDate := now.AddDate(0, 0, -10).Format("2006-01-02")
				
				return &dynamodb.ScanOutput{
					Items: []map[string]types.AttributeValue{
						{
							"BucketKey": &types.AttributeValueMemberS{Value: oldDate},
						},
					},
				}, nil
			},
			mockDeleteItem: func(ctx context.Context, params *dynamodb.DeleteItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.DeleteItemOutput, error) {
				return nil, aws.NewError("InternalServerError")
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock client
			mockClient := &MockDynamoDBClient{
				ScanFunc:       tt.mockScan,
				DeleteItemFunc: tt.mockDeleteItem,
			}

			// Create counter storage with mock client
			counterStorage := NewCounterStorage(mockClient)

			// Call CleanupOldBuckets
			err := counterStorage.CleanupOldBuckets(context.Background(), tt.daysToKeep)

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("CleanupOldBuckets() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
} 