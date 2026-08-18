// Package middleware provides HTTP middleware.
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/redaction"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a middleware that ensures all error responses follow.
// the structured format without leaking sensitive information. It only fires.
// when a handler recorded a gin error (c.Error) and did not already write a.
// body; handlers that emit their own structured response keep full control.
func ErrorHandler(logger *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		// Don't override a response that's already been written.
		if c.Writer.Written() {
			return
		}

		ginErr := c.Errors.Last()
		traceID := GetTraceID(c)

		// If the recorded error is a structured ValidationError, render a 400.
		// with field-level details rather than the generic status→code path.
		if details, ok := errors.ValidationDetailsOf(ginErr.Err); ok {
			logStructuredError(logger, c, ginErr, traceID, http.StatusBadRequest, errors.CodeValidationFailed)
			c.JSON(http.StatusBadRequest, responses.StructuredErrorResponse{
				Error: responses.ErrorDetail{
					Code:    string(errors.CodeValidationFailed),
					Message: "One or more fields have validation errors",
					Details: details,
					TraceID: traceID,
					DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeValidationFailed)),
				},
			})
			return
		}

		// If the recorded error is a *errors.ServerError, render it directly from.
		// the error's own code/message/details/docs-url, using the code's canonical.
		// HTTP status. This is the path migrated handlers take: they record a.
		// ServerError via c.Error and return, and this middleware is the single.
		// render point.
		if se := errors.AsServerError(ginErr.Err); se != nil {
			status := se.Code.HTTPStatusCode()
			if se.TraceID == "" {
				se.TraceID = traceID
			}
			logStructuredError(logger, c, ginErr, traceID, status, se.Code)
			c.JSON(status, responses.StructuredErrorResponse{
				Error: responses.ErrorDetail{
					Code:    string(se.Code),
					Message: se.Message,
					Details: se.Details,
					TraceID: se.TraceID,
					DocsURL: se.DocsURL,
				},
			})
			return
		}

		// Determine status code. If a handler set an error status, honor it;
		// otherwise default to 500 (a recorded error with a 200 is a bug).
		status := c.Writer.Status()
		if status < http.StatusBadRequest {
			status = http.StatusInternalServerError
		}

		code := statusToErrorCode(status)
		safeMessage := getSafeMessageForStatus(status)

		logStructuredError(logger, c, ginErr, traceID, status, code)

		c.JSON(status, responses.StructuredErrorResponse{
			Error: responses.ErrorDetail{
				Code:    string(code),
				Message: safeMessage,
				TraceID: traceID,
				DocsURL: tracing.BuildErrorDocsURL(string(code)),
			},
		})
	}
}

// logStructuredError logs an error with its structural fields intact (trace_id,
// path, method, status, actor_id, org_id) and redacts only the value of the.
// error message, which may carry sensitive data. Routing this through.
// SanitizeForLog previously dropped path/method/error_message entirely, making.
// the log line useless for debugging.
func logStructuredError(logger *slog.Logger, c *gin.Context, ginErr *gin.Error, traceID string, status int, code errors.ErrorCode) {
	args := []any{
		slog.String("trace_id", traceID),
		slog.String("error_code", string(code)),
		slog.String("path", c.Request.URL.Path),
		slog.String("method", c.Request.Method),
		slog.Int("status", status),
	}
	if userID := c.GetString("operator_id"); userID != "" {
		args = append(args, slog.String("actor_id", userID))
	}
	if orgID := c.GetString("org_id"); orgID != "" {
		args = append(args, slog.String("org_id", orgID))
	}
	if ginErr != nil {
		args = append(args, slog.String("error_type", "gin.Error"),
			slog.String("error_message", redaction.DefaultRedactor.Redact(ginErr.Error())))
	}

	if status >= http.StatusInternalServerError {
		logger.Error("request_error", args...)
	} else {
		logger.Warn("request_error", args...)
	}
}

// statusToErrorCode maps HTTP status codes to a canonical error code. This is.
// the lossy fallback path used when a handler only set a status / recorded a.
// plain error. Handlers that build a real ServerError should call.
// responses.RespondWithServerError directly to preserve the specific code.
func statusToErrorCode(status int) errors.ErrorCode {
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
		// 400-range without a specific match → validation/client error.
		return errors.CodeValidationFailed
	}
}

// getSafeMessageForStatus returns a safe error message based on HTTP status.
// These messages are intentionally generic to avoid leaking internal details.
func getSafeMessageForStatus(status int) string {
	switch {
	case status == http.StatusUnauthorized:
		return "Authentication required"
	case status == http.StatusForbidden:
		return "You don't have permission to perform this action"
	case status == http.StatusNotFound:
		return "Resource not found"
	case status == http.StatusConflict:
		return "Resource already exists"
	case status == http.StatusTooManyRequests:
		return "Too many requests. Please slow down."
	case status == http.StatusRequestTimeout, status == http.StatusGatewayTimeout:
		return "The request timed out. Please try again."
	case status >= http.StatusInternalServerError:
		return "An unexpected error occurred. Please try again later."
	default:
		return "Request could not be completed"
	}
}
