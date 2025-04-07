package shortener

import (
	"context"
	"encoding/base62"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/jingy/Go-Shortener/internal/models"
	"github.com/jingy/Go-Shortener/internal/storage"
)

const (
	shortCodeLength = 6 // Reduced from 8 to 6 for shorter URLs
)

type Shortener struct {
	baseURL string
	counter *storage.CounterStorage
}

func NewShortener(baseURL string, counter *storage.CounterStorage) *Shortener {
	return &Shortener{
		baseURL: strings.TrimRight(baseURL, "/"),
		counter: counter,
	}
}

// GenerateShortCode generates a unique short code using a counter within the current day bucket
func (s *Shortener) GenerateShortCode(ctx context.Context) (string, error) {
	// Get next counter value
	counter, err := s.counter.GetNextCounter(ctx)
	if err != nil {
		return "", fmt.Errorf("failed to get counter: %w", err)
	}

	// Get current timestamp (seconds since epoch)
	timestamp := time.Now().UTC().Unix()
	
	// Combine timestamp and counter to create a unique value
	// We use the last 6 digits of the timestamp (to keep it manageable)
	// and combine it with the counter value
	timestampLast6 := timestamp % 1000000 // Last 6 digits of timestamp
	combinedValue := timestampLast6*1000 + counter // Combine with counter (assuming counter < 1000)
	
	// Convert to base62 string
	shortCode := base62.Encode(uint64(combinedValue))
	
	// Pad with leading zeros if needed
	if len(shortCode) < shortCodeLength {
		shortCode = strings.Repeat("0", shortCodeLength-len(shortCode)) + shortCode
	}

	return shortCode, nil
}

// ValidateURL checks if the provided URL is valid
func (s *Shortener) ValidateURL(urlStr string) error {
	if urlStr == "" {
		return models.ErrEmptyURL
	}

	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return models.ErrInvalidURL
	}

	if parsedURL.Scheme == "" || parsedURL.Host == "" {
		return models.ErrInvalidURL
	}

	return nil
}

// CreateShortURL creates a new short URL
func (s *Shortener) CreateShortURL(ctx context.Context, originalURL string) (*models.URL, error) {
	if err := s.ValidateURL(originalURL); err != nil {
		return nil, err
	}

	shortCode, err := s.GenerateShortCode(ctx)
	if err != nil {
		return nil, err
	}

	url := models.NewURL(originalURL, shortCode)
	url.ShortURL = s.GetShortURL(shortCode)

	return url, nil
}

// GetShortURL returns the full short URL for a given short code
func (s *Shortener) GetShortURL(shortCode string) string {
	return s.baseURL + "/" + shortCode
} 
} 