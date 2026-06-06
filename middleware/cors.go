package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type CORS struct {
	AllowedOrigins []string
	MaxAge         string // Preflight cache duration
}

func (co CORS) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")

		// Check if origin is allowed
		if co.allowed(origin) {
			c.Writer.Header().Set("Vary", "Origin")
			// Only set Allow-Origin header for allowed origins
			// Never use wildcard "*" when credentials are involved
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			// Allow credentials for authenticated requests
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		// Always set these headers for CORS support
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", strings.Join([]string{
			"Authorization",
			"Content-Type",
			"X-Request-ID",
			"X-Vyzorix-Nonce",
			"X-Vyzorix-Timestamp",
			"X-Vyzorix-Signature",
			"X-Vyzorix-Token",
		}, ", "))

		// Cache preflight response for 1 hour by default
		if co.MaxAge != "" {
			c.Writer.Header().Set("Access-Control-Max-Age", co.MaxAge)
		} else {
			c.Writer.Header().Set("Access-Control-Max-Age", "3600")
		}

		// Handle preflight requests
		if c.Request.Method == "OPTIONS" {
			c.Writer.WriteHeader(204)
			return
		}
		c.Next()
	}
}

func (co CORS) allowed(origin string) bool {
	if origin == "" {
		return false // Reject requests without Origin header for security
	}
	for _, v := range co.AllowedOrigins {
		if v == "*" {
			return true // Wildcard only allowed if explicitly configured
		}
		if strings.EqualFold(v, origin) {
			return true
		}
	}
	return false
}

func CORSHandler(origins []string) gin.HandlerFunc {
	return CORS{AllowedOrigins: origins}.Handler()
}
