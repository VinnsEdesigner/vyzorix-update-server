// Package origin provides WebSocket origin validation.
package origin

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// Validator validates WebSocket origins against an allowed list.
type Validator struct {
	allowedOrigins map[string]bool
	log            *slog.Logger
	allowWildcard  bool
}

// NewValidator creates a validator with allowed origins from config.
func NewValidator(origins []string) *Validator {
	allowed := make(map[string]bool)
	allowWildcard := false

	for _, origin := range origins {
		origin = strings.TrimSpace(origin)
		if origin == "" {
			continue
		}

		if origin == "*" {
			allowWildcard = true
			continue
		}

		normalized := strings.ToLower(origin)
		allowed[origin] = true
		allowed[normalized] = true
		allowed[strings.ToUpper(origin)] = true
	}

	return &Validator{
		allowedOrigins: allowed,
		allowWildcard:  allowWildcard,
	}
}

// SetLogger sets the logger for security events.
func (v *Validator) SetLogger(log *slog.Logger) {
	v.log = log
}

// Validate checks if the origin is allowed.
func (v *Validator) Validate(origin string) bool {
	if origin == "" {
		return true
	}

	if v.allowWildcard {
		return true
	}

	if v.allowedOrigins[origin] {
		return true
	}

	normalized := strings.ToLower(origin)
	if v.allowedOrigins[normalized] {
		return true
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	if u.Scheme == "http" || u.Scheme == "ws" {
		return false
	}

	normalized = strings.ToLower(origin)
	return v.allowedOrigins[normalized]
}

// ValidateWithDetails returns validation result with details for logging.
func (v *Validator) ValidateWithDetails(origin string) (bool, string) {
	if origin == "" {
		return true, "empty origin (non-browser client)"
	}

	if v.allowWildcard {
		return true, "wildcard origin allowed"
	}

	if v.allowedOrigins[origin] {
		return true, "direct match"
	}

	normalized := strings.ToLower(origin)
	if v.allowedOrigins[normalized] {
		return true, "case-insensitive match"
	}

	u, err := url.Parse(origin)
	if err != nil {
		return false, "malformed origin URL"
	}

	if u.Scheme == "http" || u.Scheme == "ws" {
		return false, "non-secure scheme rejected"
	}

	return false, "origin not in allowed list"
}

// CheckOrigin returns a function compatible with gorilla/websocket Upgrader.
func (v *Validator) CheckOrigin() func(*http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		valid, reason := v.ValidateWithDetails(origin)

		if !valid && origin != "" {
			if v.log != nil {
				path := ""
				if r.URL != nil {
					path = r.URL.Path
				}
				v.log.Warn("websocket origin rejected",
					"origin", origin,
					"reason", reason,
					"remoteAddr", r.RemoteAddr,
					"path", path,
				)
			}
		}

		return valid
	}
}

// CheckOriginWithoutLogging returns a function without logging (for testing).
func (v *Validator) CheckOriginWithoutLogging() func(*http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return v.Validate(origin)
	}
}

// AllowedOrigins returns a copy of the allowed origins.
func (v *Validator) AllowedOrigins() []string {
	origins := make([]string, 0, len(v.allowedOrigins))
	for origin := range v.allowedOrigins {
		origins = append(origins, origin)
	}
	return origins
}

// IsWildcardAllowed returns whether wildcard origin is configured.
func (v *Validator) IsWildcardAllowed() bool {
	return v.allowWildcard
}
