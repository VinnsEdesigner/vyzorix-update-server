package middleware

import (
	"context"
	"net/http"
	"strings"

	domain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	keys "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// =============================================================================
// AUTHENTICATION BOUNDARY DEFINITIONS
// Based on MULTI_CLIENT_API_KEY_SYSTEM.md Section 2 - Endpoint Authentication
// =============================================================================

// PathType defines the authentication boundary for a path
type PathType int

const (
	PathTypeUnknown PathType = iota
	PathTypePublic           // No auth required
	PathTypeInfrastructure   // Env API Key (TokenSecret)
	PathTypeSessionOnly      // Session Cookie required
	PathTypeDeviceAuth       // HMAC Signature
	PathTypeTenant           // Session OR API Key + Scope
)

// PathBoundary maps path patterns to their authentication requirements
var PathBoundaries = map[string]PathType{
	// PUBLIC - No auth required
	"/health":                       PathTypePublic,
	"/healthz":                      PathTypePublic, // Protected at route level
	"/v1/auth/":                     PathTypePublic,
	"/v1/device/register":           PathTypePublic,
	"/v1/device/public/":            PathTypePublic,
	"/v1/device/inbox":             PathTypePublic,
	"/v1/device/confirm":            PathTypePublic,
	"/metrics":                      PathTypePublic, // Prometheus scraping

	// INFRASTRUCTURE - Env API Key (TokenSecret) - handled at route level
	// /admin/*, /internal/*

	// SESSION ONLY - Session Cookie required
	"/bin/":           PathTypeSessionOnly,
	"/v1/dashboard/":  PathTypeSessionOnly,
	"/v1/api-keys/":   PathTypeSessionOnly,
	"/api/v1/apk/":   PathTypeSessionOnly,

	// DEVICE AUTH - HMAC Signature - handled by device middleware
	// /v1/device/:imei/command, /v1/device/:imei/fcm-token

	// TENANT - Session OR API Key + Scope (default)
	"/v1/devices/":            PathTypeTenant,
	"/v1/device/":             PathTypeTenant,
	"/v1/command/":            PathTypeTenant,
	"/v1/telemetry/":          PathTypeTenant,
	"/v1/updates/":            PathTypeTenant,
	"/v1/device/diagnostics/": PathTypeTenant,
}

// ClassifyPath determines the PathType for a given path
func ClassifyPath(path string) PathType {
	// Check exact matches first
	if pt, ok := PathBoundaries[path]; ok {
		return pt
	}

	// Check prefix matches
	for prefix, pt := range PathBoundaries {
		if strings.HasPrefix(path, prefix) {
			return pt
		}
	}

	return PathTypeTenant // Default to TENANT
}

// IsPublicPath returns true if the path is PUBLIC (no auth required)
func IsPublicPath(path string) bool {
	return ClassifyPath(path) == PathTypePublic
}

// IsInfrastructurePath returns true if the path is INFRASTRUCTURE (Env API Key)
func IsInfrastructurePath(path string) bool {
	return ClassifyPath(path) == PathTypeInfrastructure
}

// IsSessionOnlyPath returns true if the path requires Session Cookie
func IsSessionOnlyPath(path string) bool {
	return ClassifyPath(path) == PathTypeSessionOnly
}

// IsDeviceAuthPath returns true if the path requires HMAC authentication
func IsDeviceAuthPath(path string) bool {
	return ClassifyPath(path) == PathTypeDeviceAuth
}

// IsTenantPath returns true if the path is TENANT (Session OR API Key + Scope)
func IsTenantPath(path string) bool {
	return ClassifyPath(path) == PathTypeTenant
}

// =============================================================================
// TENANT API KEY AUTHENTICATION MIDDLEWARE
// =============================================================================

// TenantAPIKeyAuth provides tenant API key authentication middleware.
type TenantAPIKeyAuth struct {
	service   *keys.Service
	keyPrefix string
}

// NewTenantAPIKeyAuth creates a new TenantAPIKeyAuth middleware.
func NewTenantAPIKeyAuth(service *keys.Service, keyPrefix string) *TenantAPIKeyAuth {
	return &TenantAPIKeyAuth{
		service:   service,
		keyPrefix: keyPrefix,
	}
}

// Middleware returns the Gin middleware function for tenant API key authentication.
// This handles TENANT paths: Session OR API Key + Scope
func (t *TenantAPIKeyAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path
		pathType := ClassifyPath(path)

		// Skip for non-tenant paths (let other middleware handle)
		if pathType != PathTypeTenant {
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
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "api_key_required",
				"message": "X-API-Key header or session required",
			})
			return
		}

		// Validate the key using the service
		key, err := t.service.ValidateKey(c.Request.Context(), apiKey)
		if err != nil {
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
		c.Set("api_key_id", key.ID)
		c.Set("api_key_operator_id", key.OperatorID)
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

// ScopeEnforcement returns a middleware that enforces scope based on HTTP method.
func (t *TenantAPIKeyAuth) ScopeEnforcement(requiredScope domain.Scope) gin.HandlerFunc {
	return func(c *gin.Context) {
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
		scope := domain.Scope(scopeStr)
		if !hasScope(scope, requiredScope) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "insufficient_scope",
				"message": "insufficient scope for this operation",
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

// methodToScope converts an HTTP method to the required scope.
func methodToScope(method string) domain.Scope {
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

// GenerateKey generates a new API key with the configured prefix.
func (t *TenantAPIKeyAuth) GenerateKey() (string, string, error) {
	uidStr := uuid.New().String()
	fullKey := t.keyPrefix + "_" + uidStr
	prefix := fullKey[:12]
	return fullKey, prefix, nil
}
