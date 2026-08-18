// Package middleware provides HTTP middleware for the Vyzorix API.
package middleware

import (
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
	"github.com/gin-gonic/gin"
)

// RequestIDHeader is the legacy inbound alias header accepted by Tracing in.
// addition to X-Trace-ID. It is no longer emitted on responses; X-Trace-ID is.
const RequestIDHeader = "X-Request-ID"

// Tracing returns a middleware that adds a trace ID to every request. The ID is.
// the single correlation identifier for the request: it is read from the.
// X-Trace-ID header (falling back to X-Request-ID for clients that send the.
// latter), generated if absent, stored in the gin context under.
// ContextKeyTraceID, and echoed back on the X-Trace-ID response header so.
// clients can correlate a response with server-side logs.
func Tracing() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Extract or generate trace ID. Accept X-Request-ID as a legacy alias.
		clientTraceID := c.GetHeader(tracing.TraceIDHeader)
		if clientTraceID == "" {
			clientTraceID = c.GetHeader(RequestIDHeader)
		}
		traceID := tracing.ExtractOrGenerate(clientTraceID)

		// Record request start time.
		start := time.Now()

		// Store in context for use by handlers.
		c.Set(tracing.ContextKeyTraceID, traceID)
		c.Set(tracing.ContextKeyRequestStart, start)

		// Add to response headers so client can correlate.
		c.Header(tracing.TraceIDHeader, traceID)

		// Process request.
		c.Next()

		// Add timing header.
		c.Header("X-Response-Time", time.Since(start).String())
	}
}

// GetTraceID extracts the trace ID from the gin context.
func GetTraceID(c *gin.Context) string {
	if traceID, exists := c.Get(tracing.ContextKeyTraceID); exists {
		if id, ok := traceID.(string); ok {
			return id
		}
	}
	return ""
}

// GetRequestStart extracts the request start time from the gin context.
func GetRequestStart(c *gin.Context) time.Time {
	if start, exists := c.Get(tracing.ContextKeyRequestStart); exists {
		if t, ok := start.(time.Time); ok {
			return t
		}
	}
	return time.Now()
}

// GetRequestDuration calculates how long the request has been in progress.
func GetRequestDuration(c *gin.Context) time.Duration {
	return time.Since(GetRequestStart(c))
}
