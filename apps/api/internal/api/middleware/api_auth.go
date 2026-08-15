// Package middleware provides HTTP middleware.
package middleware

import (
	"github.com/gin-gonic/gin"
)

type Authenticator struct {
	ServerAPIToken    string
	DevelopmentBypass bool
}

func (a Authenticator) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if a.DevelopmentBypass {
			c.Next()
			return
		}

		if a.ServerAPIToken == "" {
			c.JSON(401, map[string]string{"error": "unauthorized", "message": "invalid or missing dashboard token"})
			c.Abort()

			return
		}

		if c.GetHeader("Authorization") == "Bearer "+a.ServerAPIToken || c.GetHeader("X-Vyzorix-Token") == a.ServerAPIToken {
			c.Next()
			return
		}

		c.JSON(401, map[string]string{"error": "unauthorized", "message": "invalid or missing dashboard token"})
		c.Abort()
	}
}
