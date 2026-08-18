<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
// Package errors provides the unified error system for Vyzorix.
package errors

import (
	"fmt"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
)

// ServerError is the standard error type returned by all Vyzorix API operations.
// It provides structured, client-safe error information with trace context.
type ServerError struct {
	// Code is a machine-readable error identifier (e.g., "AUTH_INVALID_CREDENTIALS").
	// Clients should use this for programmatic error handling, not the Message.
	Code ErrorCode `json:"code"`

	// Message is a human-readable description safe for client display.
	// It should never contain sensitive information (tokens, passwords, internal paths).
	Message string `json:"message"`

	// Details contains additional context-specific data.
	// For validation errors, this contains []ValidationDetail.
	// For rate limit errors, this may contain RetryAfter duration.
	Details any `json:"details,omitempty"`

	// TraceID is the unique identifier for request tracing.
	// Use NewServerError or WrapError to automatically generate.
	TraceID string `json:"trace_id"`

	// Timestamp is when the error occurred (UTC).
	Timestamp time.Time `json:"timestamp"`

	// DocsURL is an optional link to documentation for this error code.
	// Format: "https://docs.vyzorix.com/errors/{ERROR_CODE}"
	DocsURL string `json:"docs_url,omitempty"`

	// Internal is internal error context that should NEVER be exposed to clients.
	// This field is omitted from JSON serialization via struct tags.
	Internal *InternalError `json:"-"`
}

// InternalError contains implementation details for logging/debugging.
// These fields are NEVER sent to clients.
type InternalError struct {
	// Original error from the underlying operation
	Err error

	// Request path where error occurred
	Path string

	// Method of the request
	Method string

	// User/System ID that triggered the error (safe to log)
	ActorID string

	// Additional internal context
	Context map[string]any
}

// Error implements the error interface for ServerError.
func (e *ServerError) Error() string {
	if e.TraceID != "" {
		return fmt.Sprintf("[%s] %s: %s", e.TraceID, e.Code, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying error if this is a wrapped error.
func (e *ServerError) Unwrap() error {
	if e.Internal != nil {
		return e.Internal.Err
	}
	return nil
}

// WithInternal adds internal error context to the server error.
// This is only for logging purposes - the internal data is never serialized.
func (e *ServerError) WithInternal(err error, path, method, actorID string) *ServerError {
	e.Internal = &InternalError{
		Err:     err,
		Path:    path,
		Method:  method,
		ActorID: actorID,
	}
	return e
}

// WithInternalContext adds key-value context to internal error data.
func (e *ServerError) WithInternalContext(key string, value any) *ServerError {
	if e.Internal == nil {
		e.Internal = &InternalError{}
	}
	if e.Internal.Context == nil {
		e.Internal.Context = make(map[string]any)
	}
	e.Internal.Context[key] = value
	return e
}

// NewServerError creates a new ServerError with current timestamp and optional trace ID.
func NewServerError(code ErrorCode, message string) *ServerError {
	return &ServerError{
		Code:      code,
		Message:   message,
		TraceID:   tracing.GenerateTraceID(),
		Timestamp: WithTimestamp(),
		DocsURL:   tracing.BuildErrorDocsURL(string(code)),
	}
}

// NewServerErrorFromStatus creates a ServerError whose code is derived from an
// HTTP status code, mirroring the middleware's status→code mapping. Use this
// when the status is not statically known (e.g. carried by a service error).
func NewServerErrorFromStatus(status int, message string) *ServerError {
	return NewServerError(statusToErrorCode(status), message)
}

// statusToErrorCode maps an HTTP status to the canonical ErrorCode. It is the
// single source of truth shared by the response helpers and the runtime-status
// ServerError constructor.
func statusToErrorCode(status int) ErrorCode {
	switch {
	case status == 401:
		return CodeAuthTokenInvalid
	case status == 403:
		return CodeAuthzInsufficientPermissions
	case status == 404:
		return CodeResourceNotFound
	case status == 409:
		return CodeResourceConflict
	case status == 429:
		return CodeRateLimitExceeded
	case status == 504:
		return CodeInternalTimeout
	case status >= 500:
		return CodeInternalServerError
	default:
		return CodeValidationFailed
	}
}

// NewServerErrorWithTrace creates a ServerError with a specific trace ID.
func NewServerErrorWithTrace(code ErrorCode, message, traceID string) *ServerError {
	return &ServerError{
		Code:      code,
		Message:   message,
		TraceID:   traceID,
		Timestamp: WithTimestamp(),
		DocsURL:   tracing.BuildErrorDocsURL(string(code)),
	}
}

// WrapError wraps an existing error into a ServerError.
func WrapError(code ErrorCode, message string, err error) *ServerError {
	return &ServerError{
		Code:      code,
		Message:   message,
		TraceID:   tracing.GenerateTraceID(),
		Timestamp: WithTimestamp(),
		Internal: &InternalError{
			Err: err,
		},
		DocsURL: tracing.BuildErrorDocsURL(string(code)),
	}
}

// WrapErrorWithTrace wraps an error with a specific trace ID.
func WrapErrorWithTrace(code ErrorCode, message, traceID string, err error) *ServerError {
	return &ServerError{
		Code:      code,
		Message:   message,
		TraceID:   traceID,
		Timestamp: WithTimestamp(),
		Internal: &InternalError{
			Err: err,
		},
		DocsURL: tracing.BuildErrorDocsURL(string(code)),
	}
}

// WithDetails sets the details field and returns the ServerError.
func (e *ServerError) WithDetails(details any) *ServerError {
	e.Details = details
	return e
}

// WithTraceID sets a specific trace ID and returns the ServerError.
func (e *ServerError) WithTraceID(traceID string) *ServerError {
	e.TraceID = traceID
	return e
}

// Common error factory functions for authentication errors

// ErrInvalidCredentials creates an invalid credentials error.
func ErrInvalidCredentials() *ServerError {
	return NewServerError(CodeAuthInvalidCredentials, "Invalid email or password")
}

// ErrAccountLocked creates an account locked error with remaining lockout time.
func ErrAccountLocked(remainingMinutes int) *ServerError {
	return NewServerError(CodeAuthAccountLocked,
		fmt.Sprintf("Account is temporarily locked. Try again in %d minutes.", remainingMinutes)).
		WithDetails(map[string]int{"retry_after_minutes": remainingMinutes})
}

// ErrMFARequired creates an MFA required error.
func ErrMFARequired() *ServerError {
	return NewServerError(CodeAuthMFARequired, "Multi-factor authentication is required")
}

// ErrSessionExpired creates a session expired error.
func ErrSessionExpired() *ServerError {
	return NewServerError(CodeAuthSessionExpired, "Your session has expired. Please log in again.")
}

// Common factory functions for authorization errors

// ErrForbidden creates a forbidden error with a generic message.
func ErrForbidden() *ServerError {
	return NewServerError(CodeAuthzInsufficientPermissions,
		"You don't have permission to perform this action")
}

// ErrForbiddenWithReason creates a forbidden error with a specific reason.
func ErrForbiddenWithReason(reason string) *ServerError {
	return NewServerError(CodeAuthzInsufficientPermissions, reason)
}

// ErrOrgMembershipRequired creates an organization membership required error.
func ErrOrgMembershipRequired() *ServerError {
	return NewServerError(CodeAuthzOrgMembershipRequired,
		"You must be a member of this organization to perform this action")
}

// Common factory functions for resource errors

// ErrNotFound creates a resource not found error.
func ErrNotFound(resourceType, resourceID string) *ServerError {
	return NewServerError(CodeResourceNotFound,
		fmt.Sprintf("%s not found", resourceType)).
		WithDetails(map[string]string{
			"resource_type": resourceType,
			"resource_id":   resourceID,
		})
}

// ErrAlreadyExists creates an already exists error.
func ErrAlreadyExists(resourceType string) *ServerError {
	return NewServerError(CodeResourceAlreadyExists,
		fmt.Sprintf("A %s with this identifier already exists", resourceType))
}

// Common factory functions for validation errors

// ErrValidationFailed creates a general validation error.
func ErrValidationFailed(message string) *ServerError {
	return NewServerError(CodeValidationFailed, message)
}

// ErrRequiredField creates a required field validation error.
func ErrRequiredField(fieldName string) *ServerError {
	return NewServerError(CodeValidationRequiredField,
		fmt.Sprintf("%s is required", fieldName)).
		WithDetails([]ValidationDetail{
			NewValidationDetail(fieldName, "this field is required"),
		})
}

// ErrInvalidFormat creates an invalid format validation error.
func ErrInvalidFormat(fieldName, expectedFormat string) *ServerError {
	return NewServerError(CodeValidationInvalidFormat,
		fmt.Sprintf("%s has an invalid format. Expected: %s", fieldName, expectedFormat)).
		WithDetails([]ValidationDetail{
			NewValidationDetail(fieldName, "invalid format: "+expectedFormat),
		})
}

// ErrValidationErrors creates a validation error with multiple field errors.
func ErrValidationErrors(errors []ValidationDetail) *ServerError {
	return NewServerError(CodeValidationFailed, "One or more fields have validation errors").
		WithDetails(errors)
}

// Common factory functions for rate limit errors

// ErrRateLimitExceeded creates a rate limit exceeded error.
func ErrRateLimitExceeded(retryAfterSeconds int) *ServerError {
	return NewServerError(CodeRateLimitExceeded, "Too many requests. Please slow down.").
		WithDetails(map[string]int{"retry_after_seconds": retryAfterSeconds})
}

// Common factory functions for security errors

// ErrThreatDetected creates a security threat detected error.
func ErrThreatDetected(threatType string) *ServerError {
	return NewServerError(CodeSecurityThreatDetected,
		"Suspicious activity detected and blocked").
		WithDetails(map[string]string{"threat_type": threatType})
}

// ErrRiskConfirmationRequired creates a risk confirmation required error.
func ErrRiskConfirmationRequired(operation string, riskLevel string) *ServerError {
	return NewServerError(CodeSecurityRiskUnconfirmed,
		fmt.Sprintf("This action (%s) requires confirmation due to its risk level", operation)).
		WithDetails(map[string]string{
			"operation":  operation,
			"risk_level": riskLevel,
		})
}

// Common factory functions for internal errors

// ErrInternal creates an internal server error.
func ErrInternal(message string) *ServerError {
	return NewServerError(CodeInternalServerError, message)
}

// ErrInternalWithError wraps an underlying error in an internal error.
func ErrInternalWithError(message string, err error) *ServerError {
	return WrapError(CodeInternalServerError, message, err)
}

// Common factory functions for device errors

// ErrDeviceNotOnline creates a device not online error.
func ErrDeviceNotOnline(deviceID string) *ServerError {
	return NewServerError(CodeDeviceNotOnline,
		"Device is not currently online").
		WithDetails(map[string]string{"device_id": deviceID})
}

// ErrDeviceCommandTimeout creates a device command timeout error.
func ErrDeviceCommandTimeout(deviceID string) *ServerError {
	return NewServerError(CodeDeviceCommandTimeout,
		"Device did not respond to command in time").
		WithDetails(map[string]string{"device_id": deviceID})
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
