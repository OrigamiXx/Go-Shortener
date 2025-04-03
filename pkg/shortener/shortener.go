package shortener

import (
	"crypto/rand"
	"encoding/base64"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/jingy/Go-Shortener/internal/models"
)

const (
	shortCodeLength = 8
	alphabet       = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

type Shortener struct {
	baseURL string
}

func NewShortener(baseURL string) *Shortener {
	return &Shortener{
		baseURL: strings.TrimRight(baseURL, "/"),
	}
}

// GenerateShortCode generates a unique short code
func (s *Shortener) GenerateShortCode() string {
	// Use UUID as a base for uniqueness
	id := uuid.New()
	
	// Convert UUID to base64 and take first 8 characters
	encoded := base64.URLEncoding.EncodeToString(id[:])
	return encoded[:shortCodeLength]
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
func (s *Shortener) CreateShortURL(originalURL string) (*models.URL, error) {
	if err := s.ValidateURL(originalURL); err != nil {
		return nil, err
	}

	shortCode := s.GenerateShortCode()
	url := models.NewURL(originalURL, shortCode)
	url.ShortURL = s.GetShortURL(shortCode)

	return url, nil
}

// GetShortURL returns the full short URL for a given short code
func (s *Shortener) GetShortURL(shortCode string) string {
	return s.baseURL + "/" + shortCode
} 