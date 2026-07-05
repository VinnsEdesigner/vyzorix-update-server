package api_key

import (
	"time"
)

// Scope defines the permission level for an API key.
type Scope string

const (
	ScopeRead  Scope = "read"
	ScopeWrite Scope = "write"
	ScopeAdmin Scope = "admin"
)

// IsValid checks if the scope is valid.
func (s Scope) IsValid() bool {
	switch s {
	case ScopeRead, ScopeWrite, ScopeAdmin:
		return true
	default:
		return false
	}
}

// CanRead returns true if the scope allows read operations.
func (s Scope) CanRead() bool {
	return s == ScopeRead || s == ScopeWrite || s == ScopeAdmin
}

// CanWrite returns true if the scope allows write operations.
func (s Scope) CanWrite() bool {
	return s == ScopeWrite || s == ScopeAdmin
}

// CanDelete returns true if the scope allows delete operations.
func (s Scope) CanDelete() bool {
	return s == ScopeAdmin
}

// APIKey represents a multi-tenant API key.
type APIKey struct {
	ID           string     `json:"id"`
	OperatorID   string     `json:"operator_id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`    // First 8 chars for display: "vxyz_a1b2"
	Scope        Scope      `json:"scope"`         // Key permissions scope
	ExpiresAt    *time.Time `json:"expires_at"`    // nil = never expires
	IsActive     bool       `json:"is_active"`
	RequestCount int64      `json:"request_count"` // Total requests made with this key
	LastRequest  *time.Time `json:"last_request"`  // Last time key was used
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
}

// IsExpired returns true if the key has expired.
func (k *APIKey) IsExpired() bool {
	if k.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*k.ExpiresAt)
}

// IsValid returns true if the key is active and not expired.
func (k *APIKey) IsValid() bool {
	return k.IsActive && !k.IsExpired()
}

// CreateAPIKeyRequest represents a request to create a new API key.
type CreateAPIKeyRequest struct {
	Name         string `json:"name" binding:"required,max=64"`
	Scope        Scope  `json:"scope" binding:"required"`
	ExpiresInDays *int  `json:"expires_in_days"` // nil = no expiration
}

// UpdateAPIKeyRequest represents a request to update an API key.
type UpdateAPIKeyRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=64"`
	Scope *Scope  `json:"scope"`
}

// APIKeyWithFullKey represents an API key response with the full key (only returned on create/rotate).
type APIKeyWithFullKey struct {
	APIKey
	FullKey string `json:"api_key"` // The full key - only shown once!
}

// APIKeyResponse represents an API key in responses (without the full key).
type APIKeyResponse struct {
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	Scope        Scope      `json:"scope"`
	ExpiresAt    *time.Time `json:"expires_at"`
	IsActive     bool       `json:"is_active"`
	RequestCount int64      `json:"request_count"`
	LastRequest  *time.Time `json:"last_request_at"`
	CreatedAt    time.Time  `json:"created_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
}

// ToResponse converts an APIKey to an APIKeyResponse.
func (k *APIKey) ToResponse() APIKeyResponse {
	return APIKeyResponse{
		ID:           k.ID,
		Name:         k.Name,
		KeyPrefix:    k.KeyPrefix,
		Scope:        k.Scope,
		ExpiresAt:    k.ExpiresAt,
		IsActive:     k.IsActive,
		RequestCount: k.RequestCount,
		LastRequest:  k.LastRequest,
		CreatedAt:    k.CreatedAt,
		RevokedAt:    k.RevokedAt,
	}
}

// ListAPIKeysResponse represents a paginated list of API keys.
type ListAPIKeysResponse struct {
	Keys               []APIKeyResponse `json:"keys"`
	Pagination         Pagination       `json:"pagination"`
	MonthlyLimit       int              `json:"monthly_limit"`
	KeysCreatedThisMonth int            `json:"keys_created_this_month"`
}

// Pagination represents pagination info.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}
