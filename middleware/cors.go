package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

type CORS struct{ AllowedOrigins []string }

func (co CORS) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if co.allowed(origin) {
			if origin == "" {
				origin = "*"
			}
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Vary", "Origin")
		}
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Vyzorix-Nonce, X-Vyzorix-Timestamp, X-Vyzorix-Signature, X-Vyzorix-Token")
		if c.Request.Method == "OPTIONS" {
			c.Writer.WriteHeader(204)
			return
		}
		c.Next()
	}
}

func (co CORS) allowed(origin string) bool {
	if origin == "" {
		return true
	}
	for _, v := range co.AllowedOrigins {
		if v == "*" || strings.EqualFold(v, origin) {
			return true
		}
	}
	return false
}

func CORSHandler(origins []string) gin.HandlerFunc {
	return CORS{AllowedOrigins: origins}.Handler()
}
