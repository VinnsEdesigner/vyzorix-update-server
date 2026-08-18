// Package errors provides the unified error system for Vyzorix.
package errors

import (
	"net/http"
	"time"
)

// ErrorCode represents a stable, machine-readable error identifier.
// These codes are used in API responses and should never change once published.
type ErrorCode string

// Authentication error codes (AUTH_*)
//
//nolint:gosec // G101 false positive: these are error-code identifiers, not hardcoded credentials.
const (
	// AUTH_INVALID_CREDENTIALS - Email or password is incorrect
	CodeAuthInvalidCredentials ErrorCode = "AUTH_INVALID_CREDENTIALS"

	// AUTH_ACCOUNT_LOCKED - Account is temporarily locked due to failed attempts
	CodeAuthAccountLocked ErrorCode = "AUTH_ACCOUNT_LOCKED"

	// AUTH_MFA_REQUIRED - MFA verification is required to complete the action
	CodeAuthMFARequired ErrorCode = "AUTH_MFA_REQUIRED"

	// AUTH_MFA_INVALID - MFA code is invalid or expired
	CodeAuthMFAInvalid ErrorCode = "AUTH_MFA_INVALID"

	// AUTH_SESSION_EXPIRED - Session has expired, please re-authenticate
	CodeAuthSessionExpired ErrorCode = "AUTH_SESSION_EXPIRED"

	// AUTH_SESSION_REVOKED - Session was explicitly revoked
	CodeAuthSessionRevoked ErrorCode = "AUTH_SESSION_REVOKED"

	// AUTH_TOKEN_INVALID - JWT or session token is malformed or invalid
	CodeAuthTokenInvalid ErrorCode = "AUTH_TOKEN_INVALID"

	// AUTH_TOKEN_EXPIRED - JWT has passed its expiration time
	CodeAuthTokenExpired ErrorCode = "AUTH_TOKEN_EXPIRED"

	// AUTH_ORIGIN_FORBIDDEN - Request origin is not allowed
	CodeAuthOriginForbidden ErrorCode = "AUTH_ORIGIN_FORBIDDEN"

	// AUTH_CSRF_INVALID - CSRF token is missing or invalid
	CodeAuthCSRFInvalid ErrorCode = "AUTH_CSRF_INVALID"
)

// Authorization error codes (AUTHZ_*)
const (
	// AUTHZ_INSUFFICIENT_PERMISSIONS - Actor lacks required role/permission
	CodeAuthzInsufficientPermissions ErrorCode = "AUTHZ_INSUFFICIENT_PERMISSIONS"

	// AUTHZ_ORG_MEMBERSHIP_REQUIRED - User must be a member of the organization
	CodeAuthzOrgMembershipRequired ErrorCode = "AUTHZ_ORG_MEMBERSHIP_REQUIRED"

	// AUTHZ_ORG_OWNERSHIP_REQUIRED - Action requires organization ownership
	CodeAuthzOrgOwnershipRequired ErrorCode = "AUTHZ_ORG_OWNERSHIP_REQUIRED"

	// AUTHZ_SCOPE_FORBIDDEN - API key lacks required scope for this operation
	CodeAuthzScopeForbidden ErrorCode = "AUTHZ_SCOPE_FORBIDDEN"

	// AUTHZ_DEVICE_ACCESS_DENIED - No access to this device
	CodeAuthzDeviceAccessDenied ErrorCode = "AUTHZ_DEVICE_ACCESS_DENIED"
)

// Resource error codes (RESOURCE_*)
const (
	// RESOURCE_NOT_FOUND - The requested resource does not exist
	CodeResourceNotFound ErrorCode = "RESOURCE_NOT_FOUND"

	// RESOURCE_ALREADY_EXISTS - A resource with these identifiers already exists
	CodeResourceAlreadyExists ErrorCode = "RESOURCE_ALREADY_EXISTS"

	// RESOURCE_CONFLICT - State conflict prevents the operation
	CodeResourceConflict ErrorCode = "RESOURCE_CONFLICT"

	// RESOURCE_DELETED - Resource was previously deleted
	CodeResourceDeleted ErrorCode = "RESOURCE_DELETED"

	// RESOURCE_LIMIT_EXCEEDED - Quota or limit for this resource type is reached
	CodeResourceLimitExceeded ErrorCode = "RESOURCE_LIMIT_EXCEEDED"
)

// Validation error codes (VALIDATION_*)
const (
	// VALIDATION_FAILED - General validation failure
	CodeValidationFailed ErrorCode = "VALIDATION_FAILED"

	// VALIDATION_REQUIRED_FIELD - A required field is missing or empty
	CodeValidationRequiredField ErrorCode = "VALIDATION_REQUIRED_FIELD"

	// VALIDATION_INVALID_FORMAT - Field value doesn't match expected format
	CodeValidationInvalidFormat ErrorCode = "VALIDATION_INVALID_FORMAT"

	// VALIDATION_INVALID_EMAIL - Email address is malformed
	CodeValidationInvalidEmail ErrorCode = "VALIDATION_INVALID_EMAIL"

	// VALIDATION_INVALID_PASSWORD - Password doesn't meet policy requirements
	CodeValidationInvalidPassword ErrorCode = "VALIDATION_INVALID_PASSWORD"

	// VALIDATION_PASSWORD_BREACHED - Password found in known data breaches
	CodeValidationPasswordBreached ErrorCode = "VALIDATION_PASSWORD_BREACHED"

	// VALIDATION_TOO_LONG - Field value exceeds maximum length
	CodeValidationTooLong ErrorCode = "VALIDATION_TOO_LONG"

	// VALIDATION_TOO_SHORT - Field value is below minimum length
	CodeValidationTooShort ErrorCode = "VALIDATION_TOO_SHORT"

	// VALIDATION_INVALID_ENUM - Value is not one of the allowed options
	CodeValidationInvalidEnum ErrorCode = "VALIDATION_INVALID_ENUM"
)

// Rate limiting error codes (RATE_LIMIT_*)
const (
	// RATE_LIMIT_EXCEEDED - Too many requests, slow down
	CodeRateLimitExceeded ErrorCode = "RATE_LIMIT_EXCEEDED"

	// RATE_LIMIT_RETRY_AFTER - Client should retry after specified time
	CodeRateLimitRetryAfter ErrorCode = "RATE_LIMIT_RETRY_AFTER"
)

// Security error codes (SECURITY_*)
const (
	// SECURITY_THREAT_DETECTED - Suspicious activity blocked
	CodeSecurityThreatDetected ErrorCode = "SECURITY_THREAT_DETECTED"

	// SECURITY_RISK_UNCONFIRMED - High-risk operation requires confirmation
	CodeSecurityRiskUnconfirmed ErrorCode = "SECURITY_RISK_UNCONFIRMED"

	// SECURITY_IP_BLOCKED - IP address is blocked
	CodeSecurityIPBlocked ErrorCode = "SECURITY_IP_BLOCKED"

	// SECURITY_DEVICE_FINGERPRINT_BLOCKED - Device fingerprint blocked
	CodeSecurityDeviceFingerprintBlocked ErrorCode = "SECURITY_DEVICE_FINGERPRINT_BLOCKED"

	// SECURITY_REPLAY_DETECTED - Request replay attack detected
	CodeSecurityReplayDetected ErrorCode = "SECURITY_REPLAY_DETECTED"

	// SECURITY_SIGNATURE_INVALID - Request signature verification failed
	CodeSecuritySignatureInvalid ErrorCode = "SECURITY_SIGNATURE_INVALID"

	// SECURITY_TURNSTILE_FAILED - Turnstile/CAPTCHA verification failed
	CodeSecurityTurnstileFailed ErrorCode = "SECURITY_TURNSTILE_FAILED"
)

// Command/device error codes (DEVICE_*)
const (
	// DEVICE_NOT_ONLINE - Device is not connected to receive commands
	CodeDeviceNotOnline ErrorCode = "DEVICE_NOT_ONLINE"

	// DEVICE_COMMAND_PENDING - A command is already pending on this device
	CodeDeviceCommandPending ErrorCode = "DEVICE_COMMAND_PENDING"

	// DEVICE_COMMAND_TIMEOUT - Device didn't respond to command in time
	CodeDeviceCommandTimeout ErrorCode = "DEVICE_COMMAND_TIMEOUT"

	// DEVICE_COMMAND_FAILED - Command execution failed on device
	CodeDeviceCommandFailed ErrorCode = "DEVICE_COMMAND_FAILED"

	// DEVICE_NOT_REGISTERED - Device ID not found in registry
	CodeDeviceNotRegistered ErrorCode = "DEVICE_NOT_REGISTERED"

	// DEVICE_ALREADY_REGISTERED - Device already registered with this ID
	CodeDeviceAlreadyRegistered ErrorCode = "DEVICE_ALREADY_REGISTERED"
)

// Organization error codes (ORG_*)
const (
	// ORG_NOT_FOUND - Organization ID not found
	CodeOrgNotFound ErrorCode = "ORG_NOT_FOUND"

	// ORG_MEMBER_LIMIT - Organization has reached member limit
	CodeOrgMemberLimit ErrorCode = "ORG_MEMBER_LIMIT"

	// ORG_INVITATION_EXPIRED - Invitation has expired
	CodeOrgInvitationExpired ErrorCode = "ORG_INVITATION_EXPIRED"

	// ORG_INVITATION_INVALID - Invitation code is invalid
	CodeOrgInvitationInvalid ErrorCode = "ORG_INVITATION_INVALID"
)

// Update system error codes (UPDATE_*)
const (
	// UPDATE_NOT_FOUND - Update version not found
	CodeUpdateNotFound ErrorCode = "UPDATE_NOT_FOUND"

	// UPDATE_IN_PROGRESS - Another update is currently in progress
	CodeUpdateInProgress ErrorCode = "UPDATE_IN_PROGRESS"

	// UPDATE_DEVICE_INCOMPATIBLE - Device is not compatible with this update
	CodeUpdateDeviceIncompatible ErrorCode = "UPDATE_DEVICE_INCOMPATIBLE"
)

// Internal error codes (INTERNAL_*)
const (
	// INTERNAL_SERVER_ERROR - Unexpected server-side error
	CodeInternalServerError ErrorCode = "INTERNAL_SERVER_ERROR"

	// INTERNAL_DATABASE_ERROR - Database operation failed
	CodeInternalDatabaseError ErrorCode = "INTERNAL_DATABASE_ERROR"

	// INTERNAL_EXTERNAL_SERVICE_ERROR - External service call failed
	CodeInternalExternalServiceError ErrorCode = "INTERNAL_EXTERNAL_SERVICE_ERROR"

	// INTERNAL_TIMEOUT - Operation timed out
	CodeInternalTimeout ErrorCode = "INTERNAL_TIMEOUT"
)

// CodeCategory returns the category prefix for an error code.
func (c ErrorCode) Category() string {
	switch {
	case c.startswith("AUTH_"):
		return "Authentication"
	case c.startswith("AUTHZ_"):
		return "Authorization"
	case c.startswith("RESOURCE_"):
		return "Resource"
	case c.startswith("VALIDATION_"):
		return "Validation"
	case c.startswith("RATE_LIMIT_"):
		return "Rate Limiting"
	case c.startswith("SECURITY_"):
		return "Security"
	case c.startswith("DEVICE_"):
		return "Device"
	case c.startswith("ORG_"):
		return "Organization"
	case c.startswith("UPDATE_"):
		return "Update"
	case c.startswith("INTERNAL_"):
		return "Internal"
	default:
		return "Unknown"
	}
}

// HTTPStatusCode returns the default HTTP status code for an error code.
// Mapping is per-code (not per-category) so that e.g. RESOURCE_ALREADY_EXISTS
// resolves to 409 Conflict rather than 404. Codes without a specific case fall
// back to their category default.
func (c ErrorCode) HTTPStatusCode() int {
	switch c {
	// Authentication → 401 (CSRF/origin violations are authn-context, 401)
	case CodeAuthInvalidCredentials, CodeAuthAccountLocked, CodeAuthMFARequired,
		CodeAuthMFAInvalid, CodeAuthSessionExpired, CodeAuthSessionRevoked,
		CodeAuthTokenInvalid, CodeAuthTokenExpired, CodeAuthOriginForbidden,
		CodeAuthCSRFInvalid:
		return http.StatusUnauthorized

	// Authorization → 403
	case CodeAuthzInsufficientPermissions, CodeAuthzOrgMembershipRequired,
		CodeAuthzOrgOwnershipRequired, CodeAuthzScopeForbidden,
		CodeAuthzDeviceAccessDenied:
		return http.StatusForbidden

	// Resource — per-code, not blanket 404
	case CodeResourceNotFound, CodeResourceDeleted:
		return http.StatusNotFound
	case CodeResourceAlreadyExists, CodeResourceConflict:
		return http.StatusConflict
	case CodeResourceLimitExceeded:
		return http.StatusTooManyRequests

	// Validation → 400
	case CodeValidationFailed, CodeValidationRequiredField,
		CodeValidationInvalidFormat, CodeValidationInvalidEmail,
		CodeValidationInvalidPassword, CodeValidationPasswordBreached,
		CodeValidationTooLong, CodeValidationTooShort, CodeValidationInvalidEnum:
		return http.StatusBadRequest

	// Rate limiting → 429
	case CodeRateLimitExceeded, CodeRateLimitRetryAfter:
		return http.StatusTooManyRequests

	// Security → 403 (threats/blocks), confirmation-required → 449 Retry With
	case CodeSecurityThreatDetected, CodeSecurityIPBlocked,
		CodeSecurityDeviceFingerprintBlocked, CodeSecurityReplayDetected,
		CodeSecuritySignatureInvalid, CodeSecurityTurnstileFailed:
		return http.StatusForbidden
	case CodeSecurityRiskUnconfirmed:
		return 449 // "Retry With": client must retry after confirming the risky op.

	// Device — mostly 400; timeouts are gateway errors
	case CodeDeviceNotOnline, CodeDeviceCommandPending,
		CodeDeviceCommandFailed, CodeDeviceNotRegistered,
		CodeDeviceAlreadyRegistered:
		return http.StatusBadRequest
	case CodeDeviceCommandTimeout:
		return http.StatusGatewayTimeout

	// Organization — per-code
	case CodeOrgNotFound:
		return http.StatusNotFound
	case CodeOrgMemberLimit, CodeOrgInvitationExpired, CodeOrgInvitationInvalid:
		return http.StatusBadRequest

	// Update — per-code
	case CodeUpdateNotFound:
		return http.StatusNotFound
	case CodeUpdateInProgress, CodeUpdateDeviceIncompatible:
		return http.StatusConflict

	// Internal → 500; timeouts are gateway errors
	case CodeInternalServerError, CodeInternalDatabaseError,
		CodeInternalExternalServiceError:
		return http.StatusInternalServerError
	case CodeInternalTimeout:
		return http.StatusGatewayTimeout
	}

	// Defensive fallback for any future code not yet listed here.
	switch {
	case c.startswith("VALIDATION_"):
		return http.StatusBadRequest
	case c.startswith("AUTHZ_"):
		return http.StatusForbidden
	case c.startswith("AUTH_"):
		return http.StatusUnauthorized
	case c.startswith("RATE_LIMIT_"):
		return http.StatusTooManyRequests
	case c.startswith("SECURITY_"):
		return http.StatusForbidden
	case c.startswith("RESOURCE_"):
		return http.StatusNotFound
	case c.startswith("DEVICE_"), c.startswith("ORG_"), c.startswith("UPDATE_"):
		return http.StatusBadRequest
	case c.startswith("INTERNAL_"):
		return http.StatusInternalServerError
	default:
		return http.StatusInternalServerError
	}
}

// IsRetryable returns true if the error code indicates a retryable condition.
func (c ErrorCode) IsRetryable() bool {
	switch c {
	case CodeRateLimitExceeded, CodeRateLimitRetryAfter,
		CodeInternalServerError, CodeInternalTimeout,
		CodeInternalExternalServiceError, CodeDeviceCommandTimeout:
		return true
	default:
		return false
	}
}

// Helper to check string prefix
func (c ErrorCode) startswith(prefix string) bool {
	s := string(c)
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

// ValidationDetail contains field-level validation error information.
type ValidationDetail struct {
	Field   string `json:"field"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// NewValidationDetail creates a new validation detail.
func NewValidationDetail(field, message string) ValidationDetail {
	return ValidationDetail{
		Field:   field,
		Message: message,
	}
}

// NewValidationDetailWithCode creates a validation detail with an error code.
func NewValidationDetailWithCode(field, message, code string) ValidationDetail {
	return ValidationDetail{
		Field:   field,
		Message: message,
		Code:    code,
	}
}

// WithTimestamp returns the current time as a time.Time.
func WithTimestamp() time.Time {
	return time.Now().UTC()
}
