package api_keys

import (
	"net/http"
	"strconv"

	apikeyapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/api_key"
	apikeydomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/api_key"
	"github.com/gin-gonic/gin"
)

// SuperAdminHandler handles super admin API key endpoints.
type SuperAdminHandler struct {
	service *apikeyapp.Service
}

// NewSuperAdminHandler creates a new super admin API key handler.
func NewSuperAdminHandler(service *apikeyapp.Service) *SuperAdminHandler {
	return &SuperAdminHandler{service: service}
}

// RegisterRoutes registers the super admin API key routes.
func (h *SuperAdminHandler) RegisterRoutes(r *gin.RouterGroup) {
	keys := r.Group("/admin/api-keys")
	{
		keys.GET("", h.ListAllKeys)
		keys.GET("/operator/:operatorId", h.GetOperatorKeys)
		keys.DELETE("/:keyId", h.ForceRevokeKey)
		keys.GET("/stats", h.GetGlobalStats)
		keys.GET("/stats/operator/:operatorId", h.GetOperatorStats)
	}
}

// ListAllKeys lists all API keys across all operators (super admin only).
func (h *SuperAdminHandler) ListAllKeys(c *gin.Context) {
	_, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	_, _ = strconv.Atoi(c.DefaultQuery("limit", "20"))

	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "use /admin/api-keys/operator/:operatorId instead",
	})
}

// GetOperatorKeys lists all API keys for a specific operator (super admin only).
func (h *SuperAdminHandler) GetOperatorKeys(c *gin.Context) {
	operatorID := c.Param("operatorId")

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	result, err := h.service.ListKeys(c.Request.Context(), operatorID, page, limit)
	if err != nil {
		status := getHTTPStatus(err)
		c.JSON(status, gin.H{
			"error":   apikeydomain.ErrorCode(err),
			"message": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"keys":                    result.Keys,
		"pagination":              result.Pagination,
		"monthly_limit":          result.MonthlyLimit,
		"keys_created_this_month": result.KeysCreatedThisMonth,
	})
}

// ForceRevokeKey force revokes an API key for any operator (super admin only).
func (h *SuperAdminHandler) ForceRevokeKey(c *gin.Context) {
	keyID := c.Param("keyId")

	err := h.service.ForceRevokeKey(c.Request.Context(), keyID)
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

// GetGlobalStats returns global API key statistics (super admin only).
func (h *SuperAdminHandler) GetGlobalStats(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "statistics endpoint not yet implemented",
	})
}

// GetOperatorStats returns API key statistics for a specific operator (super admin only).
func (h *SuperAdminHandler) GetOperatorStats(c *gin.Context) {
	operatorID := c.Param("operatorId")

	page := 1
	limit := 20

	result, err := h.service.ListKeys(c.Request.Context(), operatorID, page, limit)
	if err != nil {
		status := getHTTPStatus(err)
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
		"keys_created_this_month":  result.KeysCreatedThisMonth,
		"monthly_limit":           result.MonthlyLimit,
	})
}
