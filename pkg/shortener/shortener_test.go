package shortener

import (
	"context"
	"testing"

	"github.com/jingy/Go-Shortener/internal/models"
	"github.com/jingy/Go-Shortener/internal/storage"
)

// MockCounterStorage is a mock implementation of the CounterStorage
type MockCounterStorage struct {
	GetNextCounterFunc func(ctx context.Context) (int64, error)
}

func (m *MockCounterStorage) GetNextCounter(ctx context.Context) (int64, error) {
	return m.GetNextCounterFunc(ctx)
}

func TestShortener_GenerateShortCode(t *testing.T) {
	tests := []struct {
		name           string
		mockCounter    func(ctx context.Context) (int64, error)
		expectedLength int
		expectError    bool
	}{
		{
			name: "successful generation",
			mockCounter: func(ctx context.Context) (int64, error) {
				return 42, nil
			},
			expectedLength: 6,
			expectError:    false,
		},
		{
			name: "counter error",
			mockCounter: func(ctx context.Context) (int64, error) {
				return 0, models.ErrURLNotFound
			},
			expectedLength: 0,
			expectError:    true,
		},
		{
			name: "large counter value",
			mockCounter: func(ctx context.Context) (int64, error) {
				return 1000000, nil
			},
			expectedLength: 6,
			expectError:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock counter
			mockCounter := &MockCounterStorage{
				GetNextCounterFunc: tt.mockCounter,
			}

			// Create shortener with mock counter
			shortener := NewShortener("https://example.com", mockCounter)

			// Call GenerateShortCode
			shortCode, err := shortener.GenerateShortCode(context.Background())

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("GenerateShortCode() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Check length if no error expected
			if !tt.expectError && len(shortCode) != tt.expectedLength {
				t.Errorf("GenerateShortCode() length = %v, expected %v", len(shortCode), tt.expectedLength)
			}
		})
	}
}

func TestShortener_ValidateURL(t *testing.T) {
	tests := []struct {
		name        string
		url         string
		expectError bool
	}{
		{
			name:        "valid URL",
			url:         "https://example.com",
			expectError: false,
		},
		{
			name:        "valid URL with path",
			url:         "https://example.com/path/to/resource",
			expectError: false,
		},
		{
			name:        "valid URL with query parameters",
			url:         "https://example.com?param=value",
			expectError: false,
		},
		{
			name:        "empty URL",
			url:         "",
			expectError: true,
		},
		{
			name:        "invalid URL",
			url:         "not-a-url",
			expectError: true,
		},
		{
			name:        "URL without scheme",
			url:         "example.com",
			expectError: true,
		},
		{
			name:        "URL without host",
			url:         "https://",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create shortener
			shortener := NewShortener("https://example.com", nil)

			// Call ValidateURL
			err := shortener.ValidateURL(tt.url)

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("ValidateURL() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}

func TestShortener_CreateShortURL(t *testing.T) {
	tests := []struct {
		name           string
		url            string
		mockCounter    func(ctx context.Context) (int64, error)
		expectError    bool
		expectedPrefix string
	}{
		{
			name: "successful creation",
			url:  "https://example.com",
			mockCounter: func(ctx context.Context) (int64, error) {
				return 42, nil
			},
			expectError:    false,
			expectedPrefix: "https://example.com/",
		},
		{
			name: "invalid URL",
			url:  "not-a-url",
			mockCounter: func(ctx context.Context) (int64, error) {
				return 42, nil
			},
			expectError:    true,
			expectedPrefix: "",
		},
		{
			name: "counter error",
			url:  "https://example.com",
			mockCounter: func(ctx context.Context) (int64, error) {
				return 0, models.ErrURLNotFound
			},
			expectError:    true,
			expectedPrefix: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock counter
			mockCounter := &MockCounterStorage{
				GetNextCounterFunc: tt.mockCounter,
			}

			// Create shortener with mock counter
			shortener := NewShortener("https://example.com", mockCounter)

			// Call CreateShortURL
			url, err := shortener.CreateShortURL(context.Background(), tt.url)

			// Check error
			if (err != nil) != tt.expectError {
				t.Errorf("CreateShortURL() error = %v, expectError %v", err, tt.expectError)
				return
			}

			// Check URL if no error expected
			if !tt.expectError {
				if url == nil {
					t.Errorf("CreateShortURL() returned nil URL")
					return
				}

				if url.OriginalURL != tt.url {
					t.Errorf("CreateShortURL() OriginalURL = %v, expected %v", url.OriginalURL, tt.url)
				}

				if len(url.ShortCode) != 6 {
					t.Errorf("CreateShortURL() ShortCode length = %v, expected 6", len(url.ShortCode))
				}

				if url.ShortURL[:len(tt.expectedPrefix)] != tt.expectedPrefix {
					t.Errorf("CreateShortURL() ShortURL = %v, expected prefix %v", url.ShortURL, tt.expectedPrefix)
				}
			}
		})
	}
}

func TestShortener_GetShortURL(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		shortCode    string
		expectedURL  string
	}{
		{
			name:        "simple URL",
			baseURL:     "https://example.com",
			shortCode:   "abc123",
			expectedURL: "https://example.com/abc123",
		},
		{
			name:        "base URL with trailing slash",
			baseURL:     "https://example.com/",
			shortCode:   "abc123",
			expectedURL: "https://example.com/abc123",
		},
		{
			name:        "base URL with path",
			baseURL:     "https://example.com/api",
			shortCode:   "abc123",
			expectedURL: "https://example.com/api/abc123",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create shortener
			shortener := NewShortener(tt.baseURL, nil)

			// Call GetShortURL
			shortURL := shortener.GetShortURL(tt.shortCode)

			// Check URL
			if shortURL != tt.expectedURL {
				t.Errorf("GetShortURL() = %v, expected %v", shortURL, tt.expectedURL)
			}
		})
	}
} 