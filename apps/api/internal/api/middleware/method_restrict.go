// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// MethodRestriction defines allowed HTTP methods for an endpoint.
type MethodRestriction struct {
	AllowedMethods []string
}

// NewMethodRestriction creates a new method restriction with the specified allowed methods.
func NewMethodRestriction(allowedMethods ...string) *MethodRestriction {
	return &MethodRestriction{
		AllowedMethods: allowedMethods,
	}
}

// Middleware returns a Gin middleware that restricts HTTP methods.
func (m *MethodRestriction) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, allowed := range m.AllowedMethods {
			if c.Request.Method == allowed {
				c.Next()
				return
			}
		}

		// Method not allowed - return 405 with proper error
		c.JSON(http.StatusMethodNotAllowed, gin.H{
			"error":   "method_not_allowed",
			"message": "the requested method is not allowed for this endpoint",
		})
		c.Abort()
	}
}

// GET returns a middleware that only allows GET requests.
func GET() gin.HandlerFunc {
	return NewMethodRestriction(http.MethodGet).Middleware()
}

// POST returns a middleware that only allows POST requests.
func POST() gin.HandlerFunc {
	return NewMethodRestriction(http.MethodPost).Middleware()
}

// GETPOST returns a middleware that only allows GET and POST requests.
func GETPOST() gin.HandlerFunc {
	return NewMethodRestriction(http.MethodGet, http.MethodPost).Middleware()
}

// DisableTrace disables TRACE method at the server level.
func DisableTrace() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodTrace {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error":   "method_not_allowed",
				"message": "TRACE method is not allowed",
			})
			c.Abort()

			return
		}

		c.Next()
	}
}

// DisableConnect disables CONNECT method at the server level.
func DisableConnect() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Method == http.MethodConnect {
			c.JSON(http.StatusMethodNotAllowed, gin.H{
				"error":   "method_not_allowed",
				"message": "CONNECT method is not allowed",
			})
			c.Abort()

			return
		}

		c.Next()
	}
}
