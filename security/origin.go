package security

import (
	"log/slog"
	"net/http"
	"net/url"
	"strings"
)

// OriginValidator validates WebSocket origins against an allowed list.
type OriginValidator struct {
	allowedOrigins map[string]bool
	log            *slog.Logger
	allowWildcard  bool
}

// NewOriginValidator creates a validator with allowed origins from config.
func NewOriginValidator(origins []string) *OriginValidator {
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

		// Normalize to lowercase for case-insensitive comparison
		normalized := strings.ToLower(origin)
		allowed[origin] = true
		allowed[normalized] = true
		allowed[strings.ToUpper(origin)] = true
	}

	return &OriginValidator{
		allowedOrigins: allowed,
		allowWildcard:  allowWildcard,
	}
}

// SetLogger sets the logger for security events.
func (v *OriginValidator) SetLogger(log *slog.Logger) {
	v.log = log
}

// Validate checks if the origin is allowed.
// Returns true if:
//   - Origin is empty (non-browser client like curl, mobile app)
//   - Origin is "*" wildcard and allowed
//   - Origin matches an entry in the allowed list
//   - Origin matches scheme://host (port-sensitive)
func (v *OriginValidator) Validate(origin string) bool {
	// Empty origin - non-browser client (curl, mobile app, server-to-server)
	if origin == "" {
		return true
	}

	// Wildcard allowed
	if v.allowWildcard {
		return true
	}

	// Direct match
	if v.allowedOrigins[origin] {
		return true
	}

	// Case-insensitive match
	normalized := strings.ToLower(origin)
	if v.allowedOrigins[normalized] {
		return true
	}

	// Parse and validate structure
	u, err := url.Parse(origin)
	if err != nil {
		return false
	}

	// Check scheme - only allow secure connections in production
	// ws:// and http:// are only allowed in development
	if u.Scheme == "http" || u.Scheme == "ws" {
		return false
	}

	// Reject non-secure schemes
	if u.Scheme != "https" && u.Scheme != "wss" && u.Scheme != "" {
		return false
	}

	return false
}

// ValidateWithDetails returns validation result with details for logging.
func (v *OriginValidator) ValidateWithDetails(origin string) (bool, string) {
	// Empty origin - always allowed (non-browser clients)
	if origin == "" {
		return true, "empty origin (non-browser client)"
	}

	// Wildcard allowed
	if v.allowWildcard {
		return true, "wildcard origin allowed"
	}

	// Direct or case-insensitive match
	if v.allowedOrigins[origin] {
		return true, "direct match"
	}

	normalized := strings.ToLower(origin)
	if v.allowedOrigins[normalized] {
		return true, "case-insensitive match"
	}

	// Parse for detailed rejection reason
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
func (v *OriginValidator) CheckOrigin() func(*http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		valid, reason := v.ValidateWithDetails(origin)

		if !valid && origin != "" {
			if v.log != nil {
				// Extract path safely - URL may be nil in some cases
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
func (v *OriginValidator) CheckOriginWithoutLogging() func(*http.Request) bool {
	return func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		return v.Validate(origin)
	}
}

// AllowedOrigins returns a copy of the allowed origins.
func (v *OriginValidator) AllowedOrigins() []string {
	origins := make([]string, 0, len(v.allowedOrigins))
	for origin := range v.allowedOrigins {
		origins = append(origins, origin)
	}
	return origins
}

// IsWildcardAllowed returns whether wildcard origin is configured.
func (v *OriginValidator) IsWildcardAllowed() bool {
	return v.allowWildcard
}
