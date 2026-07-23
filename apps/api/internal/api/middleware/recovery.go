// Package middleware provides HTTP middleware.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/gin-gonic/gin"
)

// PanicRecovery returns a middleware that recovers from panics and returns a.
// sanitized 500 error response without exposing stack traces.
func PanicRecovery(logger *slog.Logger) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					// Get stack trace for logging (never sent to client).
					stack := debug.Stack()
					stackStr := string(stack)

					// Log the panic with full details for debugging.
					logger.Error("panic recovered",
						slog.Any("error", err),
						slog.String("path", r.URL.Path),
						slog.String("method", r.Method),
						slog.String("remote_addr", r.RemoteAddr),
						slog.String("user_agent", r.UserAgent()),
						slog.String("stack", stackStr),
					)

					// Return generic error to client - never expose internal details.
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					_, _ = w.Write([]byte(`{"error":"internal_server_error","message":"An unexpected error occurred"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

// GinPanicRecovery returns a Gin middleware that recovers from panics and.
// returns a sanitized 500 error response.
func GinPanicRecovery(logger *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Get stack trace for logging (never sent to client).
				stack := debug.Stack()
				stackStr := string(stack)

				// Extract just the relevant stack frames (skip this function).
				lines := strings.Split(stackStr, "\n")

				var relevantLines []string

				skipNext := true // Skip the first line (panic line).
				for _, line := range lines {
					if skipNext {
						skipNext = false
						continue
					}
					// Skip internal runtime frames.
					if strings.Contains(line, "runtime/") {
						continue
					}

					relevantLines = append(relevantLines, line)
					// Only keep first 20 relevant lines.
					if len(relevantLines) >= 20 {
						break
					}
				}

				// Log the panic with full details for debugging (only if logger is not nil).
				if logger != nil {
					logger.Error("panic recovered",
						slog.Any("error", err),
						slog.String("path", c.Request.URL.Path),
						slog.String("method", c.Request.Method),
						slog.String("remote_addr", c.Request.RemoteAddr),
						slog.String("user_agent", c.Request.UserAgent()),
						slog.String("stack", strings.Join(relevantLines, "\n")),
					)
				}

				// Return generic error to client - never expose internal details.
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error":   "internal_server_error",
					"message": "An unexpected error occurred",
				})
			}
		}()
		c.Next()
	}
}
