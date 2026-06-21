// Package middleware provides HTTP middleware.
package middleware

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
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

			// Return sanitized error response - use centralized responses package.
			responses.RespondWithMessage(c, status, responses.GetErrorCode(status), responses.GetErrorMessage(status))
		}
	}
}