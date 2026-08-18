package auth

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	apikeyapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	apikeydomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// Handler handles API key management endpoints.
type Handler struct {
	service     *apikeyapp.APIKeyService
	auditLogger audit.AuditLogger
}

// NewHandler creates a new API key handler.
func NewHandler(service *apikeyapp.APIKeyService, auditLogger audit.AuditLogger) *Handler {
	return &Handler{
		service:     service,
		auditLogger: auditLogger,
	}
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
	operatorIDVal, exists := c.Get("operator_id")
	if !exists {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "operator not found"))

		return
	}
	operatorID, ok := operatorIDVal.(string)
	if !ok {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "invalid operator id"))

		return
	}

	var req apikeydomain.CreateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))

		return
	}

	result, err := h.service.GenerateKey(c.Request.Context(), operatorID, middleware.GetOrganizationID(c), &req)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	// Audit log successful key creation.
	h.auditLogger.APIKeyCreated(
		c.Request.Context(),
		operatorID,
		result.ID,
		result.Name,
		result.KeyPrefix,
		string(result.Scope),
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	// Return the full key only on creation.
	c.JSON(http.StatusCreated, gin.H{
		"id":              result.ID,
		"operator_id":     result.OperatorID,
		"name":            result.Name,
		"api_key":         result.FullKey, // Full key - only time it's shown!.
		"key_prefix":      result.KeyPrefix,
		"scope":           result.Scope,
		"expires_at":      result.ExpiresAt,
		"is_active":       result.IsActive,
		"request_count":   result.RequestCount,
		"created_at":      result.CreatedAt,
		"updated_at":      result.UpdatedAt,
		"last_request_at": result.LastRequest,
		"revoked_at":      result.RevokedAt,
	})
}

// ListKeys lists all API keys for the authenticated operator.
func (h *Handler) ListKeys(c *gin.Context) {
	operatorIDVal, exists := c.Get("operator_id")
	if !exists {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "operator not found"))

		return
	}
	operatorID, ok := operatorIDVal.(string)
	if !ok {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "invalid operator id"))

		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.service.ListKeys(c.Request.Context(), operatorID, page, limit)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":                    result.Keys,
		"pagination":              result.Pagination,
		"monthly_limit":           result.MonthlyLimit,
		"keys_created_this_month": result.KeysCreatedThisMonth,
	})
}

// GetKey gets a single API key by ID.
func (h *Handler) GetKey(c *gin.Context) {
	operatorIDVal, exists := c.Get("operator_id")
	if !exists {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "operator not found"))

		return
	}
	operatorID, ok := operatorIDVal.(string)
	if !ok {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "invalid operator id"))

		return
	}

	keyID := c.Param("keyId")

	key, err := h.service.GetKey(c.Request.Context(), operatorID, keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
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
	operatorIDVal, exists := c.Get("operator_id")
	if !exists {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "operator not found"))

		return
	}
	operatorID, ok := operatorIDVal.(string)
	if !ok {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "invalid operator id"))

		return
	}

	keyID := c.Param("keyId")

	var req apikeydomain.UpdateAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))

		return
	}

	key, err := h.service.UpdateKey(c.Request.Context(), operatorID, keyID, &req)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	// Audit log successful key update.
	// Build changes summary.
	changes := ""
	if req.Name != nil {
		changes += "name"
	}
	if req.Scope != nil {
		if changes != "" {
			changes += ", "
		}
		changes += "scope"
	}
	h.auditLogger.APIKeyUpdated(
		c.Request.Context(),
		operatorID,
		keyID,
		key.Name,
		changes,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	c.JSON(http.StatusOK, key.ToResponse())
}

// RevokeKey revokes an API key.
func (h *Handler) RevokeKey(c *gin.Context) {
	operatorIDVal, exists := c.Get("operator_id")
	if !exists {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "operator not found"))

		return
	}
	operatorID, ok := operatorIDVal.(string)
	if !ok {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "invalid operator id"))

		return
	}

	keyID := c.Param("keyId")

	// Get key info for audit before revoking.
	key, err := h.service.GetKey(c.Request.Context(), operatorID, keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	err = h.service.RevokeKey(c.Request.Context(), operatorID, keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	// Audit log successful key revocation.
	h.auditLogger.APIKeyRevoked(
		c.Request.Context(),
		operatorID,
		keyID,
		key.Name,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	c.Status(http.StatusNoContent)
}

// RotateKey rotates an API key, generating a new key and invalidating the old one.
func (h *Handler) RotateKey(c *gin.Context) {
	operatorIDVal, exists := c.Get("operator_id")
	if !exists {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "operator not found"))

		return
	}
	operatorID, ok := operatorIDVal.(string)
	if !ok {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "invalid operator id"))

		return
	}

	keyID := c.Param("keyId")

	result, err := h.service.RotateKey(c.Request.Context(), operatorID, keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	// Audit log successful key rotation.
	h.auditLogger.APIKeyRotated(
		c.Request.Context(),
		operatorID,
		result.ID,
		result.Name,
		c.ClientIP(),
		c.GetHeader("User-Agent"),
	)

	// Return the new full key.
	c.JSON(http.StatusOK, gin.H{
		"id":              result.ID,
		"operator_id":     result.OperatorID,
		"name":            result.Name,
		"api_key":         result.FullKey, // New full key - only time it's shown!.
		"key_prefix":      result.KeyPrefix,
		"scope":           result.Scope,
		"expires_at":      result.ExpiresAt,
		"is_active":       result.IsActive,
		"request_count":   result.RequestCount,
		"created_at":      result.CreatedAt,
		"updated_at":      result.UpdatedAt,
		"last_request_at": result.LastRequest,
		"revoked_at":      result.RevokedAt,
	})
}
