package admin

import (
	"net/http"
	"strconv"

	apikeyapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	apikeydomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	"github.com/gin-gonic/gin"
)

// SuperAdminHandler handles super admin API key endpoints.
type SuperAdminHandler struct {
	service     *apikeyapp.APIKeyService
	auditLogger *audit.Logger
}

// NewSuperAdminHandler creates a new super admin API key handler.
func NewSuperAdminHandler(service *apikeyapp.APIKeyService, auditLogger *audit.Logger) *SuperAdminHandler {
	return &SuperAdminHandler{
		service:     service,
		auditLogger: auditLogger,
	}
}

// RegisterRoutes registers the super admin API key routes.
// Note: The /admin/api-keys prefix is already added by the caller in setupAdminRoutes.
func (h *SuperAdminHandler) RegisterRoutes(r *gin.RouterGroup) {
	// r is already the /admin/api-keys group from setupAdminRoutes.
	r.GET("", h.ListAllKeys)
	r.GET("/operator/:operatorId", h.GetOperatorKeys)
	r.DELETE("/:keyId", h.ForceRevokeKey)
	r.GET("/stats", h.GetGlobalStats)
	r.GET("/stats/operator/:operatorId", h.GetOperatorStats)
}

// ListAllKeys lists all API keys across all operators (super admin only).

func (h *SuperAdminHandler) ListAllKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	result, err := h.service.ListAllKeys(c.Request.Context(), page, limit)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":       result.Keys,
		"pagination": result.Pagination,
	})
}

// GetOperatorKeys lists all API keys for a specific operator (super admin only).

func (h *SuperAdminHandler) GetOperatorKeys(c *gin.Context) {
	operatorID := c.Param("operatorId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

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

// ForceRevokeKey force revokes an API key for any operator (super admin only).
func (h *SuperAdminHandler) ForceRevokeKey(c *gin.Context) {
	keyID := c.Param("keyId")

	// Get key info for audit before revoking.
	key, err := h.service.GetKey(c.Request.Context(), "", keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	err = h.service.ForceRevokeKey(c.Request.Context(), keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	// Audit log force revocation by super admin.
	if h.auditLogger != nil {
		h.auditLogger.APIKeyRevoked(
			c.Request.Context(),
			key.OperatorID,
			keyID,
			key.Name,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
	}

	c.Status(http.StatusNoContent)
}

// GetGlobalStats returns global API key statistics (super admin only).
func (h *SuperAdminHandler) GetGlobalStats(c *gin.Context) {
	stats, err := h.service.GetGlobalStats(c.Request.Context())
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"total_active_keys": stats.TotalActiveKeys,
		"max_per_month":     stats.MaxPerMonth,
	})
}

// GetOperatorStats returns API key statistics for a specific operator (super admin only).
func (h *SuperAdminHandler) GetOperatorStats(c *gin.Context) {
	operatorID := c.Param("operatorId")

	page := 1
	limit := 20

	result, err := h.service.ListKeys(c.Request.Context(), operatorID, page, limit)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	totalKeys := int64(result.Pagination.Total)
	activeKeys := int64(0)
	for _, key := range result.Keys {
		if key.IsActive {
			activeKeys++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"operator_id":             operatorID,
		"total_keys":              totalKeys,
		"active_keys":             activeKeys,
		"revoked_keys":            totalKeys - activeKeys,
		"keys_created_this_month": result.KeysCreatedThisMonth,
		"monthly_limit":           result.MonthlyLimit,
	})
}
