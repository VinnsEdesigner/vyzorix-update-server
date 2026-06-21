// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	security "github.com/VinnsEdesigner/vyzorix/apps/api/internal/auth"
)

// RevocationConfig holds session revocation configuration.
type RevocationConfig struct {
	Enabled bool
}

// DefaultRevocationConfig returns the default revocation configuration.
func DefaultRevocationConfig() RevocationConfig {
	return RevocationConfig{
		Enabled: true, // Enabled by default for security
	}
}

// LoadRevocationConfig loads revocation configuration from environment variables.
func LoadRevocationConfig() RevocationConfig {
	return RevocationConfig{
		Enabled: getEnvBool("SESSION_REVOCATION_ENABLED", true), // Default to enabled
	}
}

// AuthRevocationMiddleware returns a middleware that checks if a session has been revoked.
func AuthRevocationMiddleware(revocationList *security.RevocationList) func(c *gin.Context) {
	return func(c *gin.Context) {
		if revocationList == nil {
			c.Next()
			return
		}

		// Get session cookie.
		cookieValue, err := c.Cookie(security.CookieName)
		if err != nil || cookieValue == "" {
			c.Next()
			return
		}

		// Hash the cookie value to check against revocation list.
		tokenHash := security.HashOperatorID(cookieValue)

		if revocationList.IsRevoked(tokenHash) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "session_revoked",
				"message": "Session has been revoked, please login again",
			})
			return
		}

		c.Next()
	}
}