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
func (t *TenantAPIKeyAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		path := c.Request.URL.Path

		// Skip authentication for public paths
		if isPublicPath(path) {
			c.Next()
			return
		}

		// Extract API key from header
		apiKey := extractAPIKeyFromHeader(c)
		if apiKey == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "api_key_required",
				"message": "X-API-Key header is required",
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

// isPublicPath returns true if the path is public and doesn't require API key auth.
func isPublicPath(path string) bool {
	publicPrefixes := []string{
		"/health",
		
		// metrics is public for Prometheus scraping
		// healthz is infrastructure-protected via load balancer
		"/v1/metrics",
		"/v1/auth/",
		"/v1/device/register",
		"/v1/device/inbox",
		"/v1/device/confirm",
		"/v1/device/public/",
		"/admin/",
		"/internal/",
	}

	for _, prefix := range publicPrefixes {
		if strings.HasPrefix(path, prefix) {
			return true
		}
	}
	return false
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
