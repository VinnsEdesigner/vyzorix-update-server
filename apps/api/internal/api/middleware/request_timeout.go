// Package middleware provides HTTP middleware.
package middleware

import (
	"context"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"

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

// timeoutSkipPaths are always excluded: liveness/readiness probes must never be
// delayed.
var timeoutSkipPaths = []string{"/health", "/healthz", "/metrics"}

// GinTimeout returns a Gin middleware that applies a request deadline.
//
// The deadline is carried on the request context; the handler runs on this
// goroutine. It deliberately does NOT run the handler in a goroutine and race it
// against a timer: gin's Context and ResponseWriter are not safe for concurrent
// use, so a parent writing a 504 while the handler still touches the same Context
// is a data race. Handlers that respect context cancellation abort early; the
// hard cutoff for a handler that ignores it is the http.Server WriteTimeout.
//
// Paths in config.SkipPaths and the built-in health/metrics paths bypass the
// deadline entirely. Per-route overrides are read from TimeoutConfig.RouteTimeouts.
func GinTimeout(config TimeoutConfig) gin.HandlerFunc {
	timeout := config.DefaultTimeout
	if timeout <= 0 {
		timeout = DefaultTimeoutConfig().DefaultTimeout
	}

	skip := make(map[string]struct{}, len(config.SkipPaths)+len(timeoutSkipPaths))
	for _, p := range config.SkipPaths {
		skip[p] = struct{}{}
	}
	for _, p := range timeoutSkipPaths {
		skip[p] = struct{}{}
	}

	return func(c *gin.Context) {
		if _, ok := skip[c.Request.URL.Path]; ok {
			c.Next()
			return
		}

		deadline := timeout
		if rt, ok := config.RouteTimeouts[c.Request.URL.Path]; ok && rt > 0 {
			deadline = rt
		}

		// The deadline is enforced with a context, not by racing the handler in a
		// goroutine: gin's Context (and its ResponseWriter) is not safe for
		// concurrent use, so having the parent write a 504 while the handler
		// goroutine still touches the same Context is a data race. Setting a
		// context deadline lets a well-behaved handler observe cancellation, and
		// the handler still runs to completion on this goroutine, so nothing is
		// written concurrently.
		ctx, cancel := context.WithTimeout(c.Request.Context(), deadline)
		defer cancel()

		c.Request = c.Request.WithContext(ctx)

		// A handler that returns after the deadline has passed is reported as a
		// timeout, without touching the Context from another goroutine.
		start := time.Now()
		c.Next()

		if ctx.Err() == context.DeadlineExceeded && !c.Writer.Written() {
			responses.RespondStructured(c, http.StatusGatewayTimeout,
				"the request timed out, please try again")
			return
		}
		_ = start
	}
}
