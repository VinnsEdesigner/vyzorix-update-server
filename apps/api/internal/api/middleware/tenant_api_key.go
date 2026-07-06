package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	keys "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	domain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	"github.com/gin-gonic/gin"
)



// PathType defines the authentication boundary for a path.
type PathType int

const (
	PathTypeUnknown        PathType = iota
	PathTypePublic                  // No auth required
	PathTypeInfrastructure          // Env API Key (TokenSecret)
	PathTypeSessionOnly             // Session Cookie required
	PathTypeDeviceAuth              // HMAC Signature
	PathTypeTenant                  // Session OR API Key + Scope
)

// PathBoundary maps path patterns to their authentication requirements.
var PathBoundaries = map[string]PathType{
	// PUBLIC - No auth required
	"/health":              PathTypePublic,
	"/v1/auth/":           PathTypePublic,
	"/v1/device/register":  PathTypePublic,
	"/v1/device/inbox":     PathTypePublic,
	"/v1/device/confirm":   PathTypePublic,
	"/v1/device/":          PathTypePublic, // /v1/device/:imei/status - device status check
	"/metrics":             PathTypePublic, // Prometheus scraping

	// INFRASTRUCTURE - TokenSecret (env var) - handled at route level
	// /admin/*, /internal/*, /healthz

	// SESSION ONLY - Session Cookie required (no API key)
	"/bin/":          PathTypeSessionOnly,
	"/v1/dashboard/": PathTypeSessionOnly,
	"/v1/api-keys/":  PathTypeSessionOnly,
	"/api/v1/apk/":   PathTypeSessionOnly,

	// DEVICE AUTH - HMAC Signature - handled by device middleware
	// /device/:imei/command, /device/:imei/fcm-token
	// These use HMAC, not session/API key

	// TENANT - Session OR API Key + Scope
	"/v1/devices/":            PathTypeTenant,
	"/v1/command/":            PathTypeTenant,
	"/v1/telemetry/":          PathTypeTenant,
	"/v1/updates/":            PathTypeTenant,
	"/v1/connections/":        PathTypeTenant,
	"/v1/device/diagnostics/": PathTypeTenant,
}

// ClassifyPath determines the PathType for a given path.
// It checks exact matches first, then prefix matches in order of longest prefix first.
// This ensures /v1/device/diagnostics/ is matched before /v1/device/.
func ClassifyPath(path string) PathType {
	// Check exact matches first
	if pt, ok := PathBoundaries[path]; ok {
		return pt
	}

	// Check prefix matches - iterate in order of longest prefix first
	// to ensure more specific paths take precedence over shorter prefixes
	longestPrefix := ""
	var longestPathType PathType
	for prefix, pt := range PathBoundaries {
		if strings.HasPrefix(path, prefix) {
			if len(prefix) > len(longestPrefix) {
				longestPrefix = prefix
				longestPathType = pt
			}
		}
	}
	if longestPrefix != "" {
		return longestPathType
	}

	return PathTypeTenant // Default to TENANT
}

// IsPublicPath returns true if the path is PUBLIC (no auth required).
func IsPublicPath(path string) bool {
	return ClassifyPath(path) == PathTypePublic
}

// IsInfrastructurePath returns true if the path is INFRASTRUCTURE (Env API Key).
func IsInfrastructurePath(path string) bool {
	return ClassifyPath(path) == PathTypeInfrastructure
}

// IsSessionOnlyPath returns true if the path requires Session Cookie.
func IsSessionOnlyPath(path string) bool {
	return ClassifyPath(path) == PathTypeSessionOnly
}

// IsDeviceAuthPath returns true if the path requires HMAC authentication.
func IsDeviceAuthPath(path string) bool {
	return ClassifyPath(path) == PathTypeDeviceAuth
}

// IsTenantPath returns true if the path is TENANT (Session OR API Key + Scope).
func IsTenantPath(path string) bool {
	return ClassifyPath(path) == PathTypeTenant
}

// =============================================================================
// TENANT API KEY AUTHENTICATION MIDDLEWARE
// =============================================================================

// TenantAPIKeyAuth provides tenant API key authentication middleware.
type TenantAPIKeyAuth struct {
	service     *keys.APIKeyService
	auditLogger *audit.Logger
}

// NewTenantAPIKeyAuth creates a new TenantAPIKeyAuth middleware.
func NewTenantAPIKeyAuth(service *keys.APIKeyService, auditLogger *audit.Logger) *TenantAPIKeyAuth {
	return &TenantAPIKeyAuth{
		service:     service,
		auditLogger: auditLogger,
	}
}

// Middleware returns the Gin middleware function for tenant API key authentication.
// This handles TENANT paths: Session OR API Key + Scope.
func (t *TenantAPIKeyAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		pathType := ClassifyPath(path)

		// Skip for non-tenant paths (let other middleware handle)
		if pathType != PathTypeTenant {
			c.Next()
			return
		}

		// Skip admin and API key management paths - these require session only
		if strings.HasPrefix(path, "/v1/admin/") ||
			strings.HasPrefix(path, "/v1/api-keys/") {
			c.Next()
			return
		}

		// Check if already authenticated via session
		if _, exists := c.Get("operator_id"); exists {
			c.Next()
			return
		}

		// Extract API key from header
		apiKey := extractAPIKeyFromHeader(c)
		if apiKey == "" {
			// Log missing API key attempt
			if t.auditLogger != nil {
				t.auditLogger.APIKeyFailed(
					c.Request.Context(),
					"",
					"",
					c.ClientIP(),
					c.GetHeader("User-Agent"),
					"missing",
				)
			}
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "api_key_required",
				"message": "X-API-Key header or session required",
			})
			return
		}

		// Validate the key using the service
		key, err := t.service.ValidateKey(c.Request.Context(), apiKey)
		if err != nil {
			// Log failed authentication attempt
			if t.auditLogger != nil {
				// Extract key prefix for logging (last 8 chars before the key ID)
				keyPrefix := ""
				if len(apiKey) > 8 {
					keyPrefix = apiKey[:8]
				}
				reason := "invalid_api_key"
				if err.Error() == "api key has expired" {
					reason = "expired"
				} else if err.Error() == "api key has been revoked" {
					reason = "revoked"
				}
				t.auditLogger.APIKeyFailed(
					c.Request.Context(),
					"", // operatorID unknown for failed auth
					keyPrefix,
					c.ClientIP(),
					c.GetHeader("User-Agent"),
					reason,
				)
			}

			statusCode := http.StatusUnauthorized
			if err.Error() == "api key has expired" {
				statusCode = http.StatusUnauthorized
			}
			c.AbortWithStatusJSON(statusCode, gin.H{
				"error":   "invalid_api_key",
				"message": err.Error(),
			})
			return
		}

		// Store key info in context for downstream use
		// CRITICAL: Also set operator_id for handlers that check it (like keys_handler.go)
		c.Set("operator_id", key.OperatorID)
		c.Set("api_key_id", key.ID)
		c.Set("api_key_scope", string(key.Scope))
		c.Set("api_key_name", key.Name)
		c.Set("auth_type", "tenant_api_key")

		// Increment usage asynchronously
		go func() {
			bgCtx := context.Background()
			_ = t.service.IncrementUsage(bgCtx, key.ID)
		}()

		c.Next()
	}
}

// extractAPIKeyFromHeader extracts the API key from the X-API-Key header.
func extractAPIKeyFromHeader(c *gin.Context) string {
	// Try X-API-Key header first
	apiKey := c.GetHeader("X-API-Key")
	if apiKey != "" {
		return apiKey
	}

	// Try Authorization header with Bearer prefix
	auth := c.GetHeader("Authorization")
	if strings.HasPrefix(auth, "Bearer ") {
		return strings.TrimPrefix(auth, "Bearer ")
	}

	return ""
}

// ScopeEnforcementFunc is a function that determines required scope based on HTTP method.
type ScopeEnforcementFunc func(method string) domain.Scope

// ScopeEnforcement returns a middleware that enforces scope based on HTTP method.
// It uses a scope determination function so different routes can have different requirements.
func (t *TenantAPIKeyAuth) ScopeEnforcement(scopeFn ScopeEnforcementFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Only enforce scope for API key auth (not session auth)
		authType, exists := c.Get("auth_type")
		if !exists || authType != "tenant_api_key" {
			// Session auth - skip scope enforcement
			c.Next()
			return
		}

		scopeVal, exists := c.Get("api_key_scope")
		if !exists {
			c.Next()
			return
		}

		scopeStr, ok := scopeVal.(string)
		if !ok {
			c.Next()
			return
		}

		// Determine required scope based on HTTP method
		requiredScope := scopeFn(c.Request.Method)
		keyScope := domain.Scope(scopeStr)

		if !hasScope(keyScope, requiredScope) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_scope",
				"message": "API key scope insufficient for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// hasScope checks if the key's scope includes the required scope.
func hasScope(keyScope, requiredScope domain.Scope) bool {
	switch requiredScope {
	case domain.ScopeRead:
		return keyScope.CanRead()
	case domain.ScopeWrite:
		return keyScope.CanWrite()
	case domain.ScopeAdmin:
		return keyScope.CanDelete()
	default:
		return false
	}
}

// MethodToScope converts an HTTP method to the required scope.
func MethodToScope(method string) domain.Scope {
	switch method {
	case "GET", "HEAD", "OPTIONS":
		return domain.ScopeRead
	case "POST", "PUT", "PATCH":
		return domain.ScopeWrite
	case "DELETE":
		return domain.ScopeAdmin
	default:
		return domain.ScopeRead
	}
}
