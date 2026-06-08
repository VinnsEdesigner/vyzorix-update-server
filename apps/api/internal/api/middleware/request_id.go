package middleware

import (
	"crypto/rand"
	"encoding/hex"

	"github.com/gin-gonic/gin"
)

// RequestIDHeader is the header name for request ID.
const RequestIDHeader = "X-Request-ID"

// RequestIDKey is the context key for request ID.
const RequestIDKey = "request_id"

// RequestIDMiddleware adds a unique request ID to each request.
// The ID is generated as a random hex string and added to:
// - Response header X-Request-ID
// - Gin context for use in handlers
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get request ID from header or generate new one
		requestID := c.GetHeader(RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		// Set in context
		c.Set(RequestIDKey, requestID)

		// Set in response header
		c.Header(RequestIDHeader, requestID)

		c.Next()
	}
}

// generateRequestID creates a random request ID.
func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b) //nolint:errcheck
	return hex.EncodeToString(b)
}

// GetRequestID retrieves the request ID from the Gin context.
func GetRequestID(c *gin.Context) string {
	if id, exists := c.Get(RequestIDKey); exists {
		if s, ok := id.(string); ok {
			return s
		}
	}
	return ""
}
