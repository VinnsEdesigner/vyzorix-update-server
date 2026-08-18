// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
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
		responses.

			// Method not allowed - return 405 with proper error.
			RespondStructured(c, http.StatusMethodNotAllowed,

				"the requested method is not allowed for this endpoint",
			)
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
			responses.RespondStructured(c, http.StatusMethodNotAllowed,

				"TRACE method is not allowed",
			)
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
			responses.RespondStructured(c, http.StatusMethodNotAllowed,

				"CONNECT method is not allowed",
			)
			c.Abort()

			return
		}

		c.Next()
	}
}
