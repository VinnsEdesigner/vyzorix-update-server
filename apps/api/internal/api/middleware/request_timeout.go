// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
	"github.com/gin-gonic/gin"
)

// TimeoutConfig holds configuration for the timeout middleware.
type TimeoutConfig struct {
	RouteTimeouts  map[string]time.Duration
	SkipPaths      []string
	DefaultTimeout time.Duration
}

// DefaultTimeoutConfig returns the default timeout configuration.
func DefaultTimeoutConfig() TimeoutConfig {
	return TimeoutConfig{
		DefaultTimeout: 30 * time.Second,
		RouteTimeouts:  make(map[string]time.Duration),
		SkipPaths:      []string{},
	}
}

// GinTimeout returns a Gin middleware that applies request timeouts.
// This is the enterprise-grade solution for Bug 49 - timeouts at middleware level.
func GinTimeout(config TimeoutConfig) func(c *gin.Context) {
	return func(c *gin.Context) {
		// Check if this path should skip timeout.
		for _, path := range config.SkipPaths {
			if c.Request.URL.Path == path {
				c.Next()
				return
			}
		}

		// Determine timeout for this request.
		timeout := config.DefaultTimeout
		if routeTimeout, ok := config.RouteTimeouts[c.Request.URL.Path]; ok {
			timeout = routeTimeout
		}

		// Create context with timeout.
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace request context with timeout context.
		c.Request = c.Request.WithContext(ctx)

		// Process request in goroutine to allow timeout detection.
		done := make(chan struct{})
		go func() {
			c.Next()
			close(done)
		}()

		// Wait for either completion or timeout.
		select {
		case <-done:
			// Request completed normally.
		case <-ctx.Done():
			// Timeout occurred.
			c.Abort()
			if !c.Writer.Written() {
				responses.RespondStructured(c, http.StatusGatewayTimeout,

					fmt.Sprintf("request timed out after %v", timeout),
				)
			}
		}
	}
}

// TimeoutMiddleware returns an http.Handler middleware for request timeouts.
// This version works with standard http.Handler (non-Gin).
func TimeoutMiddleware(defaultTimeout time.Duration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx, cancel := context.WithTimeout(r.Context(), defaultTimeout)
			defer cancel()

			r = r.WithContext(ctx)

			done := make(chan struct{})
			go func() {
				next.ServeHTTP(w, r)
				close(done)
			}()

			select {
			case <-done:
				// Request completed.
			case <-ctx.Done():
				// Emit the structured error envelope so this net/http-level
				// timeout matches the gin structured-error shape. The docs URL.
				// is built dynamically via the tracing builder (not hardcoded).
				body, _ := json.Marshal(responses.StructuredErrorResponse{
					Error: responses.ErrorDetail{
						Code:    string(errors.CodeInternalTimeout),
						Message: "The request timed out. Please try again.",
						DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeInternalTimeout)),
					},
				})
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusGatewayTimeout)
				_, _ = w.Write(body)
			}
		})
	}
}
