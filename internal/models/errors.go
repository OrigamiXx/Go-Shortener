package models

import "errors"

var (
	ErrEmptyURL          = errors.New("empty URL provided")
	ErrInvalidURL        = errors.New("invalid URL format")
	ErrURLNotFound       = errors.New("URL not found")
	ErrURLExpired        = errors.New("URL has expired")
	ErrDuplicateShortCode = errors.New("duplicate short code")
) 