package domain

import (
	"context"
)

// APIKeyRepository defines the interface for API key data access.
type APIKeyRepository interface {
	// Create creates a new API key.
	Create(ctx context.Context, key *APIKey, keyHash string) error

	// GetByID retrieves an API key by its ID.
	GetByID(ctx context.Context, id string) (*APIKey, error)

	// GetByKey retrieves an API key by the full key (for validation).
	GetByKey(ctx context.Context, key string) (*APIKey, string, error)

	// GetByPrefix retrieves an API key by its prefix (for lookup).
	GetByPrefix(ctx context.Context, prefix string) (*APIKey, string, error)

	// ListByOperator lists all API keys for an operator with pagination.
	ListByOperator(ctx context.Context, operatorID string, page, limit int) ([]*APIKey, int, error)

	// ListAll lists all API keys with pagination (for super admin).
	ListAll(ctx context.Context, page, limit int, operatorID *string) ([]*APIKey, int, error)

	// Update updates an API key.
	Update(ctx context.Context, key *APIKey) error

	// Revoke revokes an API key.
	Revoke(ctx context.Context, id string) error

	// Delete soft-deletes an API key (marks as inactive).
	Delete(ctx context.Context, id string) error

	// IncrementUsage increments the request count for an API key.
	IncrementUsage(ctx context.Context, id string) error

	// CountByOperatorThisMonth counts how many keys an operator has created this month.
	CountByOperatorThisMonth(ctx context.Context, operatorID string) (int, error)

	// GetMonthlyUsage gets usage stats for an operator this month.
	GetMonthlyUsage(ctx context.Context, operatorID string) (*MonthlyUsage, error)
}

// MonthlyUsage represents monthly API key usage statistics.
type MonthlyUsage struct {
	KeysCreated int   `json:"keys_created"`
	TotalReqs  int64 `json:"total_requests"`
}
