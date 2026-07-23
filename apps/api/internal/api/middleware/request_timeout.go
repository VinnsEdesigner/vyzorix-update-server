// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

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
				c.JSON(http.StatusGatewayTimeout, gin.H{
					"error":   "timeout",
					"code":    "REQUEST_TIMEOUT",
					"message": fmt.Sprintf("request timed out after %v", timeout),
				})
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
				http.Error(w, `{"error":"timeout","code":"REQUEST_TIMEOUT","message":"request timed out"}`,
					http.StatusGatewayTimeout)
			}
		})
	}
}
