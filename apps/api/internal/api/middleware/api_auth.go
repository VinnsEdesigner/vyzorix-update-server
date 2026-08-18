// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
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
			responses.RespondStructured(c, http.StatusUnauthorized, "invalid or missing dashboard token")
			c.Abort()
			return
		}

		if c.GetHeader("Authorization") == "Bearer "+a.ServerAPIToken || c.GetHeader("X-Vyzorix-Token") == a.ServerAPIToken {
			c.Next()
			return
		}

		responses.RespondStructured(c, http.StatusUnauthorized, "invalid or missing dashboard token")
		c.Abort()
	}
}
