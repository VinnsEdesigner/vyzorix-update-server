// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"

	infraauth "github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security"
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
func AuthRevocationMiddleware(revocationList *infraauth.RevocationList) func(c *gin.Context) {
	return func(c *gin.Context) {
		if revocationList == nil {
			c.Next()
			return
		}

		// Get session cookie.
		cookieValue, err := c.Cookie(infraauth.CookieName)
		if err != nil || cookieValue == "" {
			c.Next()
			return
		}

		// Hash the cookie value to check against revocation list.
		tokenHash := infraauth.HashOperatorID(cookieValue)

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
