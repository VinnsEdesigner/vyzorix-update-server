// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"

	gintimeout "github.com/gin-contrib/timeout"
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

// longLivedPathSuffixes are paths that must never be subject to the request
// timeout: WebSocket upgrades and server-sent event streams are long-lived by
// design, and the timeout wrapper's buffering writer would also break the
// hijacked connection.
var longLivedPathSuffixes = []string{"/ws", "/stream"}

// streamingPathPrefixes are path prefixes served by streaming/file handlers.
// The timeout wrapper buffers the entire response in memory, which would be
// wasteful (or fatal) for static assets and APK/binary downloads, so they bypass
// it as well.
var streamingPathPrefixes = []string{"/assets", "/bin", "/apk", "/api/v1/apk", "/v1/apk"}

// isLongLivedPath reports whether the path is a WebSocket/SSE endpoint that must
// bypass the timeout wrapper.
func isLongLivedPath(path string) bool {
	for _, suffix := range longLivedPathSuffixes {
		if len(path) >= len(suffix) && path[len(path)-len(suffix):] == suffix {
			return true
		}
	}
	return false
}

// isStreamingPath reports whether the path is served by a streaming/file handler
// that must bypass the buffering timeout wrapper.
func isStreamingPath(path string) bool {
	for _, prefix := range streamingPathPrefixes {
		if len(path) >= len(prefix) && path[:len(prefix)] == prefix {
			return true
		}
	}
	return false
}

// GinTimeout returns a Gin middleware that aborts a request after a configured
// deadline, responding with the standard structured error envelope. It wraps
// github.com/gin-contrib/timeout, which runs the rest of the chain under a
// buffered writer and discards the response if the deadline elapses.
//
// Two categories of request deliberately pass through untouched:
//   - paths in config.SkipPaths, and
//   - long-lived paths (WebSocket upgrades, SSE streams), which the buffering
//     writer would break.
func GinTimeout(config TimeoutConfig) gin.HandlerFunc {
	timeout := config.DefaultTimeout
	if timeout <= 0 {
		timeout = DefaultTimeoutConfig().DefaultTimeout
	}

	skip := make(map[string]struct{}, len(config.SkipPaths))
	for _, p := range config.SkipPaths {
		skip[p] = struct{}{}
	}

	timeoutResponse := func(c *gin.Context) {
		// Only write the envelope if the buffered handler produced nothing.
		if !c.Writer.Written() {
			responses.RespondStructured(c, http.StatusGatewayTimeout,
				"the request timed out, please try again")
		}
	}

	inner := gintimeout.New(
		gintimeout.WithTimeout(timeout),
		gintimeout.WithHandler(func(c *gin.Context) { c.Next() }),
		gintimeout.WithResponse(timeoutResponse),
	)

	return func(c *gin.Context) {
		if _, ok := skip[c.Request.URL.Path]; ok {
			c.Next()
			return
		}
		if isLongLivedPath(c.Request.URL.Path) || isStreamingPath(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Per-route override: when a route declares its own timeout, run the
		// library with that deadline instead of the default.
		if rt, ok := config.RouteTimeouts[c.Request.URL.Path]; ok && rt > 0 {
			gintimeout.New(
				gintimeout.WithTimeout(rt),
				gintimeout.WithHandler(func(c *gin.Context) { c.Next() }),
				gintimeout.WithResponse(timeoutResponse),
			)(c)
			return
		}

		inner(c)
	}
}
