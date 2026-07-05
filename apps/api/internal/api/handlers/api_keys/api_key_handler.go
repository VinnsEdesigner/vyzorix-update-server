package api_keys

import (
	"net/http"
	"strconv"

	apikeyapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/api_key"
	apikeydomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/api_key"
	"github.com/gin-gonic/gin"
)

// Handler handles API key management endpoints.
type Handler struct {
	service *apikeyapp.Service
}

// NewHandler creates a new API key handler.
func NewHandler(service *apikeyapp.Service) *Handler {
	return &Handler{service: service}
}

// RegisterRoutes registers the API key management routes.
func (h *Handler) RegisterRoutes(r *gin.RouterGroup) {
	keys := r.Group("/api-keys")
	{
		keys.POST("", h.CreateKey)
		keys.GET("", h.ListKeys)
		keys.GET("/:keyId", h.GetKey)
		keys.PATCH("/:keyId", h.UpdateKey)
		keys.DELETE("/:keyId", h.RevokeKey)
		keys.POST("/:keyId/rotate", h.RotateKey)
	}
}

// CreateKey creates a new API key.
func (h *Handler) CreateKey(c *gin.Context) {
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "operator not found",
		})
		return
	}

	var req apikeydomain.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	result, err := h.service.GenerateKey(c.Request.Context(), operatorID.(string), &req)
	if err != nil {
		status := getHTTPStatus(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	// Return the full key only on creation
	c.JSON(http.StatusCreated, gin.H{
		"id":          result.ID,
		"name":        result.Name,
		"api_key":     result.FullKey, // Full key - only time it's shown!
		"key_prefix":  result.KeyPrefix,
		"scope":       result.Scope,
		"expires_at":  result.ExpiresAt,
		"created_at":  result.CreatedAt,
	})
}

// ListKeys lists all API keys for the authenticated operator.
func (h *Handler) ListKeys(c *gin.Context) {
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "operator not found",
		})
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.service.ListKeys(c.Request.Context(), operatorID.(string), page, limit)
	if err != nil {
		status := getHTTPStatus(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":                   result.Keys,
		"pagination":             result.Pagination,
		"monthly_limit":         result.MonthlyLimit,
		"keys_created_this_month": result.KeysCreatedThisMonth,
	})
}

// GetKey gets a single API key by ID.
func (h *Handler) GetKey(c *gin.Context) {
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "operator not found",
		})
		return
	}

	keyID := c.Param("keyId")

	key, err := h.service.GetKey(c.Request.Context(), operatorID.(string), keyID)
	if err != nil {
		status := getHTTPStatus(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, key.ToResponse())
}

// UpdateKey updates an API key (rename, change scope).
func (h *Handler) UpdateKey(c *gin.Context) {
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "operator not found",
		})
		return
	}

	keyID := c.Param("keyId")

	var req apikeydomain.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "validation_error",
			"message": err.Error(),
		})
		return
	}

	key, err := h.service.UpdateKey(c.Request.Context(), operatorID.(string), keyID, &req)
	if err != nil {
		status := getHTTPStatus(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, key.ToResponse())
}

// RevokeKey revokes an API key.
func (h *Handler) RevokeKey(c *gin.Context) {
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "operator not found",
		})
		return
	}

	keyID := c.Param("keyId")

	err := h.service.RevokeKey(c.Request.Context(), operatorID.(string), keyID)
	if err != nil {
		status := getHTTPStatus(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.Status(http.StatusNoContent)
}

// RotateKey rotates an API key, generating a new key and invalidating the old one.
func (h *Handler) RotateKey(c *gin.Context) {
	operatorID, exists := c.Get("operator_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error":   "unauthorized",
			"message": "operator not found",
		})
		return
	}

	keyID := c.Param("keyId")

	result, err := h.service.RotateKey(c.Request.Context(), operatorID.(string), keyID)
	if err != nil {
		status := getHTTPStatus(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	// Return the new full key
	c.JSON(http.StatusOK, gin.H{
		"id":          result.ID,
		"name":        result.Name,
		"api_key":     result.FullKey, // New full key - only time it's shown!
		"key_prefix":  result.KeyPrefix,
		"scope":       result.Scope,
		"expires_at":  result.ExpiresAt,
		"created_at":  result.CreatedAt,
	})
}

// getHTTPStatus returns the appropriate HTTP status for an error.
func getHTTPStatus(err error) int {
	switch err {
	case apikeydomain.ErrAPIKeyNotFound:
		return http.StatusNotFound
	case apikeydomain.ErrAPIKeyExpired, apikeydomain.ErrAPIKeyRevoked, apikeydomain.ErrAPIKeyInactive:
		return http.StatusUnauthorized
	case apikeydomain.ErrInsufficientScope:
		return http.StatusForbidden
	case apikeydomain.ErrMonthlyLimitExceeded, apikeydomain.ErrKeyNameConflict:
		return http.StatusForbidden
	case apikeydomain.ErrRateLimitExceeded:
		return http.StatusTooManyRequests
	case apikeydomain.ErrAPIKeyRequired, apikeydomain.ErrInvalidScope, apikeydomain.ErrKeyNameTooLong, apikeydomain.ErrInvalidExpiryDays:
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}
