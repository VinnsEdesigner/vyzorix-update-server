package middleware

import (
	"net/http"
	"strings"
)

type CORS struct{ AllowedOrigins []string }

func (c CORS) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if c.allowed(origin) {
			if origin == "" {
				origin = "*"
			}
			w.Header().Set("Access-Control-Allow-Origin", origin)
			w.Header().Set("Vary", "Origin")
		}
		w.Header().Set("Access-Control-Allow-Methods", "GET, HEAD, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type, X-Vyzorix-Nonce, X-Vyzorix-Timestamp, X-Vyzorix-Signature, X-Vyzorix-Token")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (c CORS) allowed(origin string) bool {
	if origin == "" {
		return true
	}
	for _, v := range c.AllowedOrigins {
		if v == "*" || strings.EqualFold(v, origin) {
			return true
		}
	}
	return false
}
