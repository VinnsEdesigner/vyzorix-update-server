// Package idempotency provides idempotency key support for API requests.
// This implements Bug 45 fix - enterprise-grade idempotency.
package idempotency

import (
	"context"
	"time"
)

// IdempotencyRecord represents a recorded idempotency key with its response.
type IdempotencyRecord struct {
	ID            string    `json:"id"`              // Primary key (idempotency key)
	Method        string    `json:"method"`          // HTTP method (POST, etc)
	Path          string    `json:"path"`            // Request path
	Hash          string    `json:"hash"`            // Hash of request body
	StatusCode    int       `json:"statusCode"`      // Response status code
	ResponseBody  []byte    `json:"responseBody"`    // Cached response body
	CreatedAt     time.Time `json:"createdAt"`       // When the record was created
	ExpiresAt     time.Time `json:"expiresAt"`       // When the record expires
	ContentType   string    `json:"contentType"`     // Response content type
	ClientIP      string    `json:"clientIp"`        // Client IP for audit
	UserAgent     string    `json:"userAgent"`       // User agent for audit
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
