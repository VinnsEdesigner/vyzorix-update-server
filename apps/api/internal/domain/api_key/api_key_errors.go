package api_key

import "errors"

// API key errors.
var (
	ErrAPIKeyNotFound          = errors.New("api key not found")
	ErrAPIKeyExpired           = errors.New("api key has expired")
	ErrAPIKeyRevoked          = errors.New("api key has been revoked")
	ErrAPIKeyInactive          = errors.New("api key is inactive")
	ErrInvalidScope           = errors.New("invalid scope")
	ErrMonthlyLimitExceeded   = errors.New("monthly key creation limit exceeded")
	ErrKeyNameConflict        = errors.New("key with this name already exists")
	ErrInsufficientScope      = errors.New("insufficient scope for this operation")
	ErrRateLimitExceeded      = errors.New("rate limit exceeded")
	ErrAPIKeyRequired         = errors.New("api key required")
	ErrInvalidAPIKey          = errors.New("invalid api key")
	ErrKeyNameTooLong         = errors.New("key name exceeds maximum length of 64 characters")
	ErrInvalidExpiryDays      = errors.New("invalid expiry days")
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
