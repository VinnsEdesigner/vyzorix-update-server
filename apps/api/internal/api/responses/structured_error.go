// Package responses provides standardized API response formatting.
package responses

import (
	"encoding/json"
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/redaction"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
	"github.com/gin-gonic/gin"
)

// ErrorDetail is the structured error response format for all API errors.
type ErrorDetail struct {
	// Code is the machine-readable error code (e.g., "AUTH_INVALID_CREDENTIALS").
	Code string `json:"code"`

	// Message is a human-readable description safe for client display.
	Message string `json:"message"`

	// Details contains additional context (validation errors, rate limit info, etc.).
	Details any `json:"details,omitempty"`

	// TraceID allows correlation with server-side logs.
	TraceID string `json:"trace_id,omitempty"`

	// DocsURL links to documentation for this error.
	DocsURL string `json:"docs_url,omitempty"`
}

// StructuredErrorResponse is the wrapper for structured error responses.
type StructuredErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// RespondWithServerError sends a structured error response using ServerError.
// Logs the internal error details but only sends safe info to the client.
func RespondWithServerError(c *gin.Context, err *errors.ServerError) {
	if err == nil {
		RespondInternalError(c)
		return
	}

	// Ensure we have a trace ID.
	if err.TraceID == "" {
		err.TraceID = GetTraceID(c)
	}
	if err.DocsURL == "" {
		err.DocsURL = tracing.BuildErrorDocsURL(string(err.Code))
	}

	// Get HTTP status from error code.
	status := err.Code.HTTPStatusCode()
	if status == 0 {
		status = http.StatusInternalServerError
	}

	// Log internal error details for debugging (if available).
	if err.Internal != nil {
		logInternalError(c, err)
	}

	// Build client-safe response.
	c.JSON(status, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(err.Code),
			Message: err.Message,
			Details: err.Details,
			TraceID: err.TraceID,
			DocsURL: err.DocsURL,
		},
	})
}

// RespondWithErrorCode sends a structured error response using an error code.
// This is a convenience method when you don't have a full ServerError.
func RespondWithErrorCode(c *gin.Context, code errors.ErrorCode, message string) {
	traceID := GetTraceID(c)

	c.JSON(code.HTTPStatusCode(), StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(code),
			Message: message,
			TraceID: traceID,
			DocsURL: tracing.BuildErrorDocsURL(string(code)),
		},
	})
}

// logInternalError logs error details internally without exposing them to the.
// / client. Sensitive values (the wrapped error, internal context) are redacted,.
// but structural fields (path, method, actor_id, error_code, trace_id) are.
// preserved so the log line is actually useful for debugging.
func logInternalError(c *gin.Context, err *errors.ServerError) {
	if err.Internal == nil {
		return
	}

	args := []any{
		slog.String("trace_id", err.TraceID),
		slog.String("error_code", string(err.Code)),
		slog.String("path", c.Request.URL.Path),
		slog.String("method", c.Request.Method),
	}
	if err.Internal.ActorID != "" {
		args = append(args, slog.String("actor_id", err.Internal.ActorID))
	}
	if err.Internal.Err != nil {
		args = append(args, slog.String("original_error",
			redaction.DefaultRedactor.Redact(err.Internal.Err.Error())))
	}
	if err.Internal.Context != nil {
		args = append(args, slog.Any("context",
			redaction.DefaultRedactor.RedactStruct(err.Internal.Context)))
	}

	category := err.Code.Category()
	switch category {
	case "Security", "Internal":
		slog.Error("security_internal_error", args...)
	default:
		slog.Warn("api_error", args...)
	}
}

// GetTraceID extracts trace ID from the request context.
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(tracing.ContextKeyTraceID); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return tracing.GenerateTraceID()
}

// RespondValidationError sends a validation error response.
func RespondValidationError(c *gin.Context, validationErrors []errors.ValidationDetail) {
	c.JSON(http.StatusBadRequest, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(errors.CodeValidationFailed),
			Message: "One or more fields have validation errors",
			Details: validationErrors,
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeValidationFailed)),
		},
	})
}

// RespondRateLimitError sends a rate limit error response with retry info.
func RespondRateLimitError(c *gin.Context, retryAfterSeconds int) {
	c.JSON(http.StatusTooManyRequests, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(errors.CodeRateLimitExceeded),
			Message: "Too many requests. Please slow down.",
			Details: map[string]int{
				"retry_after_seconds": retryAfterSeconds,
			},
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeRateLimitExceeded)),
		},
	})
}

// RespondSecurityError sends a security-related error response.
func RespondSecurityError(c *gin.Context, code errors.ErrorCode, threatType string) {
	c.JSON(http.StatusForbidden, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(code),
			Message: "Security policy violation",
			Details: map[string]string{
				"threat_type": threatType,
			},
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(code)),
		},
	})
}

// RespondAuthError sends an authentication error response.
func RespondAuthError(c *gin.Context, code errors.ErrorCode, message string) {
	c.JSON(http.StatusUnauthorized, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(code),
			Message: message,
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(code)),
		},
	})
}

// RespondForbiddenError sends a forbidden error response.
func RespondForbiddenError(c *gin.Context, message string) {
	if message == "" {
		message = "You don't have permission to perform this action"
	}

	c.JSON(http.StatusForbidden, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(errors.CodeAuthzInsufficientPermissions),
			Message: message,
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeAuthzInsufficientPermissions)),
		},
	})
}

// RespondNotFoundError sends a not found error response.
func RespondNotFoundError(c *gin.Context, resourceType string) {
	message := resourceType + " not found"
	if resourceType == "" {
		message = "Resource not found"
	}

	c.JSON(http.StatusNotFound, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(errors.CodeResourceNotFound),
			Message: message,
			Details: map[string]string{
				"resource_type": resourceType,
			},
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeResourceNotFound)),
		},
	})
}

// RespondConflictError sends a conflict error response.
func RespondConflictError(c *gin.Context, message string) {
	if message == "" {
		message = "Resource already exists"
	}

	c.JSON(http.StatusConflict, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(errors.CodeResourceAlreadyExists),
			Message: message,
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeResourceAlreadyExists)),
		},
	})
}

// RespondInternalError sends a generic internal server error.
// This is the fallback for unhandled errors.
func RespondInternalError(c *gin.Context) {
	traceID := GetTraceID(c)

	// Log the actual error for debugging.
	slog.Error("unhandled_internal_error",
		slog.String("trace_id", traceID),
		slog.String("path", c.Request.URL.Path),
		slog.String("method", c.Request.Method),
	)

	c.JSON(http.StatusInternalServerError, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(errors.CodeInternalServerError),
			Message: "An unexpected error occurred. Please try again later.",
			TraceID: traceID,
			DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeInternalServerError)),
		},
	})
}

// statusCodeToErrorCode maps an HTTP status to the canonical error code. It.
// mirrors the middleware's statusToErrorCode so handlers that emit a structured.
// response directly (rather than recording a gin error) produce the same code.
// the error middleware would.
func statusCodeToErrorCode(status int) errors.ErrorCode {
	switch {
	case status == http.StatusUnauthorized:
		return errors.CodeAuthTokenInvalid
	case status == http.StatusForbidden:
		return errors.CodeAuthzInsufficientPermissions
	case status == http.StatusNotFound:
		return errors.CodeResourceNotFound
	case status == http.StatusRequestTimeout:
		return errors.CodeInternalTimeout
	case status == http.StatusConflict:
		return errors.CodeResourceConflict
	case status == http.StatusTooManyRequests:
		return errors.CodeRateLimitExceeded
	case status == http.StatusGatewayTimeout:
		return errors.CodeInternalTimeout
	case status >= http.StatusInternalServerError:
		return errors.CodeInternalServerError
	default:
		return errors.CodeValidationFailed
	}
}

// RespondStructured is a drop-in for legacy `c.JSON(status, gin.H{...})` calls,
// so every error response shares the structured envelope with a trace id and
// docs link.
func RespondStructured(c *gin.Context, status int, message string) {
	code := StatusToErrorCode(status)
	c.JSON(status, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(code),
			Message: message,
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(code)),
		},
	})
}

// StatusToErrorCode is the exported form of the status→code mapping, for callers
// (such as the apierrors adapter) that need the canonical code for a bare HTTP
// status.
func StatusToErrorCode(status int) errors.ErrorCode {
	return statusCodeToErrorCode(status)
}

// WithTrace takes an already-encoded structured error body and returns a copy
// with the trace id and docs URL filled in. It is used by value-typed responses
// that build the envelope before a request context exists.
//
// It returns ok=false when body is not a structured error envelope, in which
// case the caller should write body unchanged.
func WithTrace(body []byte, traceID string) ([]byte, bool) {
	var env StructuredErrorResponse
	if err := json.Unmarshal(body, &env); err != nil || env.Error.Code == "" {
		return nil, false
	}
	if env.Error.TraceID == "" {
		env.Error.TraceID = traceID
	}
	if env.Error.DocsURL == "" {
		env.Error.DocsURL = tracing.BuildErrorDocsURL(env.Error.Code)
	}
	out, err := json.Marshal(env)
	if err != nil {
		return nil, false
	}
	return out, true
}

// RespondStructuredAbort is RespondStructured plus c.Abort(), the drop-in.
// replacement for legacy `c.AbortWithStatusJSON(status, gin.H{"error","message"})`.
// middleware calls. Aborting is essential in middleware so the chain stops.
// after an error response is written.
func RespondStructuredAbort(c *gin.Context, status int, message string) {
	RespondStructured(c, status, message)
	c.Abort()
}

// RespondStructuredDetails is RespondStructured with an explicit error code and
// optional structured details (e.g. field-level validation errors). Use it when
// the code is known more precisely than the status alone implies, or when the
// client needs machine-readable detail.
func RespondStructuredDetails(c *gin.Context, status int, code errors.ErrorCode, message string, details any) {
	if code == "" {
		code = statusCodeToErrorCode(status)
	}
	c.JSON(status, StructuredErrorResponse{
		Error: ErrorDetail{
			Code:    string(code),
			Message: message,
			Details: details,
			TraceID: GetTraceID(c),
			DocsURL: tracing.BuildErrorDocsURL(string(code)),
		},
	})
}

// RespondStructuredAbortDetails is RespondStructuredDetails plus c.Abort().
func RespondStructuredAbortDetails(c *gin.Context, status int, code errors.ErrorCode, message string, details any) {
	RespondStructuredDetails(c, status, code, message, details)
	c.Abort()
}
