// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// BodySizeLimit returns a middleware that limits request body size.
// The limit is applied to all incoming JSON and form bodies.
func BodySizeLimit(limit int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.ContentLength > limit {
			c.AbortWithStatusJSON(http.StatusRequestEntityTooLarge, gin.H{
				"error":   "request_too_large",
				"message": "Request body exceeds maximum allowed size",
			})
			return
		}

		// Wrap the body with a limited reader
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, limit)
		c.Next()
	}
}

// DefaultBodySizeLimit is 1MB - reasonable for most API requests.
const DefaultBodySizeLimit int64 = 1 << 20 // 1 MB

// LargeBodySizeLimit is 8MB - for file uploads like APKs.
const LargeBodySizeLimit int64 = 8 << 20 // 8 MB
