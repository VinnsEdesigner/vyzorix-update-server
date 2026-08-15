// Package idempotency provides idempotency key support for API requests.

package idempotency

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound indicates an idempotency record was not found.
var ErrNotFound = errors.New("idempotency record not found")

// ErrConflict indicates a conflict when creating an idempotency record.
var ErrConflict = errors.New("idempotency record already exists")

// IdempotencyRecord represents a recorded idempotency key with its response.
type IdempotencyRecord struct {
	ID             string
	Method         string
	Path           string
	Hash           string
	ContentType    string
	ClientIP       string
	UserAgent      string
	OrganizationID string
	CreatedAt      time.Time
	ExpiresAt      time.Time
	ResponseBody   []byte
	StatusCode     int
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
	// Delete removes an idempotency record by key.
	Delete(ctx context.Context, key string) error
	// DeleteExpired removes all expired records (cleanup job).
	DeleteExpired(ctx context.Context) (int64, error)
	// Count returns the total number of records.
	Count(ctx context.Context) (int64, error)
	// CountByPath returns the number of records for a given path prefix.
	CountByPath(ctx context.Context, pathPrefix string) (int64, error)
	// GetByPath retrieves records by path prefix with pagination.
	GetByPath(ctx context.Context, pathPrefix string, limit, offset int) ([]*IdempotencyRecord, error)
}
