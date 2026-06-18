// Package middleware provides HTTP middleware.
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ErrorHandler returns a middleware that ensures all error responses follow.
// a consistent, secure format without leaking sensitive information.
func ErrorHandler(logger *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		c.Next()

		// Handle any errors that occurred during the request.
		if len(c.Errors) > 0 {
			err := c.Errors.Last()

			// Log the error with details (for debugging).
			logger.Warn("request error",
				slog.String("path", c.Request.URL.Path),
				slog.String("method", c.Request.Method),
				slog.Int("status", c.Writer.Status()),
				slog.String("error", err.Error()),
			)

			// Don't override a response that's already been written.
			if c.Writer.Written() {
				return
			}

			// Determine status code from context or error.
			status := c.Writer.Status()
			if status == 200 {
				// If we ended with 200 but have errors, something went wrong.
				status = http.StatusInternalServerError
			}

			// Return sanitized error response - no internal details exposed.
			c.JSON(status, gin.H{
				"error":   getErrorCode(status),
				"message": getErrorMessage(status),
			})
		}
	}
}

// getErrorCode returns a safe error code string based on HTTP status.
// We use generic error codes that don't expose implementation details.
func getErrorCode(status int) string {
	switch {
	case status >= 500:
		return "internal_server_error"
	case status == http.StatusUnauthorized:
		return "unauthorized"
	case status == http.StatusForbidden:
		return "forbidden"
	case status == http.StatusNotFound:
		return "not_found"
	case status == http.StatusBadRequest:
		return "bad_request"
	case status == http.StatusTooManyRequests:
		return "rate_limit_exceeded"
	case status >= 400:
		return "client_error"
	default:
		return "error"
	}
}

// getErrorMessage returns a safe, user-friendly error message.
// We NEVER expose internal details, paths, or implementation specifics.
func getErrorMessage(status int) string {
	switch {
	case status >= 500:
		return "An unexpected error occurred"
	case status == http.StatusUnauthorized:
		return "Authentication required"
	case status == http.StatusForbidden:
		return "Access denied"
	case status == http.StatusNotFound:
		return "Resource not found"
	case status == http.StatusBadRequest:
		return "Invalid request"
	case status == http.StatusTooManyRequests:
		return "Too many requests, please try again later"
	case status >= 400:
		return "Request could not be completed"
	default:
		return "An error occurred"
	}
}

// RespondWithError is a helper to respond with a sanitized error.
// Use this in handlers instead of returning raw error details.
func RespondWithError(c *gin.Context, status int, err error) {
	// Log the actual error internally.
	if err != nil {
		c.Set("error", err.Error())
	}

	// But return only sanitized info to client.
	c.JSON(status, gin.H{
		"error":   getErrorCode(status),
		"message": getErrorMessage(status),
	})
}

// RespondWithErrorMessage is a helper to respond with a sanitized error.
// and a custom message (that doesn't reveal internals).
func RespondWithErrorMessage(c *gin.Context, status int, message string) {
	c.JSON(status, gin.H{
		"error":   getErrorCode(status),
		"message": message,
	})
}