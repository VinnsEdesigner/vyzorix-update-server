// Package middleware provides HTTP middleware.
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/tracing"
	"github.com/gin-gonic/gin"
)

// GinPanicRecovery returns a Gin middleware that recovers from panics, logs the.
// stack trace internally (never sent to the client), and returns the structured.
// error envelope with a trace_id so the panic response correlates with logs.
// It must be registered after the Tracing middleware so a trace_id exists.
func GinPanicRecovery(logger *slog.Logger) func(c *gin.Context) {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				traceID := GetTraceID(c)

				args := []any{
					slog.String("trace_id", traceID),
					slog.Any("error", err),
					slog.String("path", c.Request.URL.Path),
					slog.String("method", c.Request.Method),
					slog.String("remote_addr", c.Request.RemoteAddr),
					slog.String("user_agent", c.Request.UserAgent()),
				}
				if userID := c.GetString("operator_id"); userID != "" {
					args = append(args, slog.String("actor_id", userID))
				}
				if orgID := c.GetString("org_id"); orgID != "" {
					args = append(args, slog.String("org_id", orgID))
				}

				if logger != nil {
					// Keep the stack trace in logs only; never expose to the client.
					stackStr := string(debug.Stack())
					lines := strings.Split(stackStr, "\n")
					var relevant []string
					skip := true
					for _, line := range lines {
						if skip {
							skip = false
							continue
						}
						if strings.Contains(line, "runtime/") {
							continue
						}
						relevant = append(relevant, line)
						if len(relevant) >= 20 {
							break
						}
					}
					args = append(args, slog.String("stack", strings.Join(relevant, "\n")))
					logger.Error("panic recovered", args...)
				}

				c.AbortWithStatusJSON(http.StatusInternalServerError, responses.StructuredErrorResponse{
					Error: responses.ErrorDetail{
						Code:    string(errors.CodeInternalServerError),
						Message: "An unexpected error occurred. Please try again later.",
						TraceID: traceID,
						DocsURL: tracing.BuildErrorDocsURL(string(errors.CodeInternalServerError)),
					},
				})
			}
		}()
		c.Next()
	}
}
