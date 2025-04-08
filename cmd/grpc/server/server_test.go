package main

import (
	"context"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/jingy/Go-Shortener/internal/storage"
	"github.com/jingy/Go-Shortener/pkg/shortener"
	pb "github.com/jingy/Go-Shortener/proto"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
	"net"
)

const bufSize = 1024 * 1024

// MockDynamoDBClient is a mock implementation of the DynamoDB client
type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) GetItem(ctx context.Context, params *dynamodb.GetItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.GetItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.GetItemOutput), args.Error(1)
}

func (m *MockDynamoDBClient) UpdateItem(ctx context.Context, params *dynamodb.UpdateItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.UpdateItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.UpdateItemOutput), args.Error(1)
}

// setupTestServer creates a test server with mocked dependencies
func setupTestServer(t *testing.T) (*grpc.Server, pb.URLShortenerClient, *bufconn.Listener) {
	// Create mock DynamoDB client
	mockDynamoClient := new(MockDynamoDBClient)
	
	// Initialize storage with mock client
	urlStorage := storage.NewURLStorage(mockDynamoClient)
	counterStorage := storage.NewCounterStorage(mockDynamoClient)
	
	// Initialize shortener
	urlShortener := shortener.NewURLShortener(urlStorage, counterStorage)
	
	// Create a buffer listener
	lis := bufconn.Listen(bufSize)
	
	// Create gRPC server
	s := grpc.NewServer()
	pb.RegisterURLShortenerServer(s, &server{
		shortener: urlShortener,
		storage:   urlStorage,
	})
	
	// Start server in a goroutine
	go func() {
		if err := s.Serve(lis); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("Failed to serve: %v", err)
		}
	}()
	
	// Create a client connection
	conn, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return lis.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		t.Fatalf("Failed to dial bufnet: %v", err)
	}
	
	// Create client
	client := pb.NewURLShortenerClient(conn)
	
	return s, client, lis
}

// teardownTestServer cleans up the test server
func teardownTestServer(s *grpc.Server, lis *bufconn.Listener) {
	s.Stop()
	lis.Close()
}

func TestCreateShortURL(t *testing.T) {
	// Setup
	s, client, lis := setupTestServer(t)
	defer teardownTestServer(s, lis)
	
	// Test cases
	tests := []struct {
		name           string
		url            string
		expirationSecs int64
		expectError    bool
	}{
		{
			name:           "valid URL",
			url:            "https://example.com",
			expirationSecs: 3600,
			expectError:    false,
		},
		{
			name:           "invalid URL",
			url:            "not-a-url",
			expirationSecs: 3600,
			expectError:    true,
		},
		{
			name:           "empty URL",
			url:            "",
			expirationSecs: 3600,
			expectError:    true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := &pb.CreateShortURLRequest{
				Url:              tt.url,
				ExpirationSeconds: tt.expirationSecs,
			}
			
			// Call the service
			resp, err := client.CreateShortURL(context.Background(), req)
			
			// Check error
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			// Check response
			assert.NoError(t, err)
			assert.NotEmpty(t, resp.ShortCode)
			assert.NotEmpty(t, resp.ShortUrl)
			assert.NotZero(t, resp.CreatedAt)
			assert.NotZero(t, resp.ExpiresAt)
			
			// Verify short URL format
			assert.Contains(t, resp.ShortUrl, resp.ShortCode)
		})
	}
}

func TestGetOriginalURL(t *testing.T) {
	// Setup
	s, client, lis := setupTestServer(t)
	defer teardownTestServer(s, lis)
	
	// Test cases
	tests := []struct {
		name        string
		shortCode   string
		expectError bool
	}{
		{
			name:        "valid short code",
			shortCode:   "abc123",
			expectError: false,
		},
		{
			name:        "invalid short code",
			shortCode:   "nonexistent",
			expectError: true,
		},
		{
			name:        "empty short code",
			shortCode:   "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := &pb.GetOriginalURLRequest{
				ShortCode: tt.shortCode,
			}
			
			// Call the service
			resp, err := client.GetOriginalURL(context.Background(), req)
			
			// Check error
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			// Check response
			assert.NoError(t, err)
			assert.NotEmpty(t, resp.OriginalUrl)
			assert.NotZero(t, resp.CreatedAt)
			assert.NotZero(t, resp.ExpiresAt)
		})
	}
}

func TestGetURLStats(t *testing.T) {
	// Setup
	s, client, lis := setupTestServer(t)
	defer teardownTestServer(s, lis)
	
	// Test cases
	tests := []struct {
		name        string
		shortCode   string
		expectError bool
	}{
		{
			name:        "valid short code",
			shortCode:   "abc123",
			expectError: false,
		},
		{
			name:        "invalid short code",
			shortCode:   "nonexistent",
			expectError: true,
		},
		{
			name:        "empty short code",
			shortCode:   "",
			expectError: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create request
			req := &pb.GetURLStatsRequest{
				ShortCode: tt.shortCode,
			}
			
			// Call the service
			resp, err := client.GetURLStats(context.Background(), req)
			
			// Check error
			if tt.expectError {
				assert.Error(t, err)
				return
			}
			
			// Check response
			assert.NoError(t, err)
			assert.Equal(t, tt.shortCode, resp.ShortCode)
			assert.NotZero(t, resp.CreatedAt)
			assert.NotZero(t, resp.ExpiresAt)
			assert.NotNil(t, resp.ClicksByCountry)
			assert.NotNil(t, resp.ClicksByHour)
		})
	}
}

// TestServerIntegration tests the integration of all three endpoints
func TestServerIntegration(t *testing.T) {
	// Setup
	s, client, lis := setupTestServer(t)
	defer teardownTestServer(s, lis)
	
	// Step 1: Create a short URL
	createReq := &pb.CreateShortURLRequest{
		Url:              "https://example.com",
		ExpirationSeconds: 3600,
	}
	
	createResp, err := client.CreateShortURL(context.Background(), createReq)
	assert.NoError(t, err)
	assert.NotEmpty(t, createResp.ShortCode)
	
	// Step 2: Get the original URL
	getReq := &pb.GetOriginalURLRequest{
		ShortCode: createResp.ShortCode,
	}
	
	getResp, err := client.GetOriginalURL(context.Background(), getReq)
	assert.NoError(t, err)
	assert.Equal(t, createReq.Url, getResp.OriginalUrl)
	
	// Step 3: Get URL stats
	statsReq := &pb.GetURLStatsRequest{
		ShortCode: createResp.ShortCode,
	}
	
	statsResp, err := client.GetURLStats(context.Background(), statsReq)
	assert.NoError(t, err)
	assert.Equal(t, createResp.ShortCode, statsResp.ShortCode)
	assert.Equal(t, createResp.CreatedAt, statsResp.CreatedAt)
	assert.Equal(t, createResp.ExpiresAt, statsResp.ExpiresAt)
} 