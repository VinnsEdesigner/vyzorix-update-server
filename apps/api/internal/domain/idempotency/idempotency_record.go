// Package idempotency provides idempotency key support for API requests.

package idempotency

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound indicates an idempotency record was not found.
var ErrNotFound = errors.New("idempotency record not found")

// IdempotencyRecord represents a recorded idempotency key with its response.
type IdempotencyRecord struct {
	CreatedAt    time.Time `json:"createdAt"`
	ExpiresAt    time.Time `json:"expiresAt"`
	ID           string    `json:"id"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Hash         string    `json:"hash"`
	ContentType  string    `json:"contentType"`
	ClientIP     string    `json:"clientIp"`
	UserAgent    string    `json:"userAgent"`
	ResponseBody []byte    `json:"responseBody"`
	StatusCode   int       `json:"statusCode"`
}

// IsExpired returns true if the record has expired.
func (r *IdempotencyRecord) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

// TTLRemaining returns the remaining time-to-live.
func (r *IdempotencyRecord) TTLRemaining() time.Duration {
	return time.Until(r.ExpiresAt)
}

// Repository defines the interface for idempotency record storage.
type Repository interface {
	// Get retrieves an idempotency record by key.
	Get(ctx context.Context, key string) (*IdempotencyRecord, error)
	// Create stores a new idempotency record.
	Create(ctx context.Context, record *IdempotencyRecord) error
	// Delete removes an expired idempotency record.
	Delete(ctx context.Context, key string) error
	// DeleteExpired removes all expired records (cleanup job).
	DeleteExpired(ctx context.Context) (int64, error)
}
