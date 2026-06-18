package client

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a client is not found.
var ErrNotFound = errors.New("client not found")

// ErrSigningKeyNotFound is returned when a signing key is not found.
var ErrSigningKeyNotFound = errors.New("signing key not found")

// Client represents an API client for request signing.
type Client struct {
	ID                string
	OperatorID        string
	Name              string
	Platform          string
	ClientSecretHash  string
	HmacKey           string
	AllowedOrigins    []string
	AllowedPaths      []string
	RateLimit         int
	IsActive          bool
	RequestCount      int64
	LastRequestAt     *int64
	CreatedAt         time.Time
	UpdatedAt         time.Time
}

// SigningKey represents a signing key for a client.
type SigningKey struct {
	ID        string
	ClientID  string
	KeyHash   string
	Version   int
	IssuedAt  int64
	ExpiresAt *int64
	IsActive  bool
}

// IsExpired returns true if the signing key has expired.
func (k *SigningKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().UnixMilli() > *k.ExpiresAt
}
