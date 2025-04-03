package models

import (
	"time"
)

// URL represents a shortened URL entry in the database
type URL struct {
	ShortCode   string    `json:"shortCode" dynamodbav:"ShortCode"`
	OriginalURL string    `json:"originalUrl" dynamodbav:"OriginalURL"`
	CreatedAt   time.Time `json:"createdAt" dynamodbav:"CreatedAt"`
	ExpiresAt   time.Time `json:"expiresAt,omitempty" dynamodbav:"ExpiresAt,omitempty"`
}

// CreateURLRequest represents the request body for creating a new short URL
type CreateURLRequest struct {
	URL      string     `json:"url"`
	ExpiresAt *time.Time `json:"expiresAt,omitempty"`
}

// CreateURLResponse represents the response for creating a new short URL
type CreateURLResponse struct {
	ShortCode string `json:"shortCode"`
	ShortURL  string `json:"shortUrl"`
}

// Validate checks if the URL is valid
func (r *CreateURLRequest) Validate() error {
	if r.URL == "" {
		return ErrEmptyURL
	}
	return nil
}

// NewURL creates a new URL instance
func NewURL(originalURL, shortCode string) *URL {
	return &URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   time.Now().UTC(),
	}
} 