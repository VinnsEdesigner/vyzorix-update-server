package middleware

import "net/http"

type Authenticator struct {
	TokenSecret       string
	DevelopmentBypass bool
}

func (a Authenticator) Dashboard(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if a.DevelopmentBypass || r.Header.Get("Authorization") == "Bearer "+a.TokenSecret || r.Header.Get("X-Vyzorix-Token") == a.TokenSecret {
			next.ServeHTTP(w, r)
			return
		}
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})
}
