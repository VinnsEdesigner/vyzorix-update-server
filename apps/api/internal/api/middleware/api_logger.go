// Package middleware provides HTTP middleware.
package middleware

import (
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger(log *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Get trace ID from context (set by Tracing middleware). This is the.
		// single correlation ID and matches the trace_id in error/panic logs.
		traceID := GetTraceID(c)

		c.Next()

		log.Info("http_request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
			"duration_ms", time.Since(start).Milliseconds(),
			"remote", c.ClientIP(),
			"trace_id", traceID,
		)
	}
}
