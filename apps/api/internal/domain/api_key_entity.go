package domain

import (
	"errors"
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
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	ExpiresAt     *time.Time `json:"expires_at"`
	LastRequest   *time.Time `json:"last_request"`
	RevokedAt     *time.Time `json:"revoked_at"`
	ID            string     `json:"id"`
	OperatorID    string     `json:"operator_id"`
	Name          string     `json:"name"`
	KeyPrefix     string     `json:"key_prefix"`
	SigningSecret string     `json:"-"`
	Scope         Scope      `json:"scope"`
	RequestCount  int64      `json:"request_count"`
	IsActive      bool       `json:"is_active"`
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
	ExpiresInDays *int   `json:"expires_in_days"`
	Name          string `json:"name" binding:"required,max=64"`
	Scope         Scope  `json:"scope" binding:"required"`
}

// UpdateAPIKeyRequest represents a request to update an API key.
type UpdateAPIKeyRequest struct {
	Name  *string `json:"name" binding:"omitempty,max=64"`
	Scope *Scope  `json:"scope"`
}

// APIKeyWithFullKey represents an API key response with the full key (only returned on create/rotate).
type APIKeyWithFullKey struct {
	FullKey    string `json:"api_key"`
	SigningKey string `json:"signing_key,omitempty"`
	APIKey
}

// APIKeyResponse represents an API key in responses (without the full key).
type APIKeyResponse struct {
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
	ExpiresAt    *time.Time `json:"expires_at"`
	LastRequest  *time.Time `json:"last_request_at"`
	RevokedAt    *time.Time `json:"revoked_at"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	KeyPrefix    string     `json:"key_prefix"`
	Scope        Scope      `json:"scope"`
	RequestCount int64      `json:"request_count"`
	IsActive     bool       `json:"is_active"`
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
		UpdatedAt:    k.UpdatedAt,
		RevokedAt:    k.RevokedAt,
	}
}

// ListAPIKeysResponse represents a paginated list of API keys.
type ListAPIKeysResponse struct {
	Keys                 []APIKeyResponse `json:"keys"`
	Pagination           Pagination       `json:"pagination"`
	MonthlyLimit         int              `json:"monthly_limit"`
	KeysCreatedThisMonth int              `json:"keys_created_this_month"`
}

// Pagination represents pagination info.
type Pagination struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// ListAllAPIKeysResponse represents a paginated list of all API keys (super admin).
type ListAllAPIKeysResponse struct {
	Keys       []AdminAPIKeyResponse `json:"keys"`
	Pagination Pagination            `json:"pagination"`
}

// AdminAPIKeyResponse extends APIKeyResponse with operator identity, which the.
// super-admin "list all keys" view requires. APIKeyResponse omits operator_id.
// because operator-scoped responses imply the operator; the admin view does not.
type AdminAPIKeyResponse struct {
	OperatorID   string `json:"operator_id"`
	OperatorName string `json:"operator_name"`
	APIKeyResponse
}

// GlobalAPIKeyStats represents global API key statistics (super admin).
// request_count is cumulative (lifetime) per key — the schema does not store a.
// per-request log, so time-windowed "today"/"this month" totals are not tracked;
// total_requests is the sum of every key's lifetime request_count.
type GlobalAPIKeyStats struct {
	RequestsByScope map[string]int64  `json:"requests_by_scope"`
	TopOperators    []TopOperatorStat `json:"top_operators"`
	TotalKeys       int               `json:"total_keys"`
	ActiveKeys      int               `json:"active_keys"`
	RevokedKeys     int               `json:"revoked_keys"`
	TotalOperators  int               `json:"total_operators"`
	MaxPerMonth     int               `json:"max_per_month"`
	TotalRequests   int64             `json:"total_requests"`
}

// TopOperatorStat is a per-operator cumulative request aggregate returned in.
// GlobalAPIKeyStats.TopOperators.
type TopOperatorStat struct {
	OperatorID     string `json:"operator_id"`
	OperatorName   string `json:"operator_name"`
	TotalRequests  int64  `json:"total_requests"`
	ActiveKeyCount int    `json:"active_key_count"`
}

// API key errors.
var (
	ErrAPIKeyNotFound       = errors.New("api key not found")
	ErrAPIKeyExpired        = errors.New("api key has expired")
	ErrAPIKeyRevoked        = errors.New("api key has been revoked")
	ErrAPIKeyInactive       = errors.New("api key is inactive")
	ErrInvalidScope         = errors.New("invalid scope")
	ErrMonthlyLimitExceeded = errors.New("monthly key creation limit exceeded")
	ErrKeyNameConflict      = errors.New("key with this name already exists")
	ErrInsufficientScope    = errors.New("insufficient scope for this operation")
	ErrRateLimitExceeded    = errors.New("rate limit exceeded")
	ErrAPIKeyRequired       = errors.New("api key required")
	ErrInvalidAPIKey        = errors.New("invalid api key")
	ErrKeyNameTooLong       = errors.New("key name exceeds maximum length of 64 characters")
	ErrInvalidExpiryDays    = errors.New("invalid expiry days")
)

// ErrorCode returns the error code for an error.
func ErrorCode(err error) string {
	switch {
	case errors.Is(err, ErrAPIKeyNotFound):
		return "key_not_found"
	case errors.Is(err, ErrAPIKeyExpired):
		return "expired_api_key"
	case errors.Is(err, ErrAPIKeyRevoked):
		return "revoked_api_key"
	case errors.Is(err, ErrAPIKeyInactive):
		return "inactive_api_key"
	case errors.Is(err, ErrInvalidScope):
		return "invalid_scope"
	case errors.Is(err, ErrMonthlyLimitExceeded):
		return "monthly_limit_exceeded"
	case errors.Is(err, ErrKeyNameConflict):
		return "key_name_conflict"
	case errors.Is(err, ErrInsufficientScope):
		return "insufficient_scope"
	case errors.Is(err, ErrRateLimitExceeded):
		return "rate_limit_exceeded"
	case errors.Is(err, ErrAPIKeyRequired):
		return "api_key_required"
	case errors.Is(err, ErrInvalidAPIKey):
		return "invalid_api_key"
	case errors.Is(err, ErrKeyNameTooLong):
		return "validation_error"
	case errors.Is(err, ErrInvalidExpiryDays):
		return "validation_error"
	default:
		return "internal_error"
	}
}

// HTTPStatusCode returns the appropriate HTTP status code for an error.
func HTTPStatusCode(err error) int {
	switch {
	case errors.Is(err, ErrAPIKeyNotFound):
		return 404
	case errors.Is(err, ErrAPIKeyExpired), errors.Is(err, ErrAPIKeyRevoked),
		errors.Is(err, ErrAPIKeyInactive), errors.Is(err, ErrInvalidAPIKey):
		return 401
	case errors.Is(err, ErrInsufficientScope):
		return 403
	case errors.Is(err, ErrMonthlyLimitExceeded), errors.Is(err, ErrKeyNameConflict):
		return 403
	case errors.Is(err, ErrRateLimitExceeded):
		return 429
	case errors.Is(err, ErrAPIKeyRequired), errors.Is(err, ErrInvalidScope),
		errors.Is(err, ErrKeyNameTooLong), errors.Is(err, ErrInvalidExpiryDays):
		return 400
	default:
		return 500
	}
}
