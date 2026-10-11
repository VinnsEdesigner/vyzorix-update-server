package admin

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	apikeyapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/keys"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	apikeydomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain"
	"github.com/gin-gonic/gin"
)

func apiKeysToAdminOpenAPI(in []apikeydomain.APIKeyResponse) []openapi.AdminAPIKey {
	out := make([]openapi.AdminAPIKey, len(in))
	for i, k := range in {
		item := openapi.AdminAPIKey{
			ID:           k.ID,
			Name:         k.Name,
			KeyPrefix:    k.KeyPrefix,
			Scope:        string(k.Scope),
			RequestCount: k.RequestCount,
			IsActive:     k.IsActive,
			CreatedAt:    k.CreatedAt.UnixMilli(),
			UpdatedAt:    k.UpdatedAt.UnixMilli(),
		}
		if k.ExpiresAt != nil {
			item.ExpiresAt = int64Ptr(k.ExpiresAt.UnixMilli())
		}
		if k.LastRequest != nil {
			item.LastRequestAt = int64Ptr(k.LastRequest.UnixMilli())
		}
		if k.RevokedAt != nil {
			item.RevokedAt = int64Ptr(k.RevokedAt.UnixMilli())
		}
		out[i] = item
	}
	return out
}

func adminKeysByOperatorToOpenAPI(in []apikeydomain.AdminAPIKeyResponse) []openapi.AdminAPIKey {
	out := make([]openapi.AdminAPIKey, len(in))
	for i, k := range in {
		item := openapi.AdminAPIKey{
			OperatorID:   k.OperatorID,
			OperatorName: k.OperatorName,
			ID:           k.ID,
			Name:         k.Name,
			KeyPrefix:    k.KeyPrefix,
			Scope:        string(k.Scope),
			RequestCount: k.RequestCount,
			IsActive:     k.IsActive,
			CreatedAt:    k.CreatedAt.UnixMilli(),
			UpdatedAt:    k.UpdatedAt.UnixMilli(),
		}
		if k.ExpiresAt != nil {
			item.ExpiresAt = int64Ptr(k.ExpiresAt.UnixMilli())
		}
		if k.LastRequest != nil {
			item.LastRequestAt = int64Ptr(k.LastRequest.UnixMilli())
		}
		if k.RevokedAt != nil {
			item.RevokedAt = int64Ptr(k.RevokedAt.UnixMilli())
		}
		out[i] = item
	}
	return out
}

func apikeydomainPagination(p apikeydomain.Pagination) openapi.Pagination {
	return openapi.Pagination{Page: p.Page, Limit: p.Limit, Total: int64(p.Total), TotalPages: p.TotalPages}
}

func int64Ptr(v int64) *int64 { return &v }

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.APIKeyListResult
	_ openapi.GlobalAPIKeyStatsResult
	_ openapi.OperatorAPIKeyStatsResult
	_ openapi.SuccessResult
	_ openapi.ErrorResponse
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

// ListAllKeys handles GET /v1/admin/api-keys.
// @Summary      List all API keys
// @Description  Lists all API keys across all operators (super admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        page   query int  false  "page number (default 1)"
// @Param        limit  query int  false  "page size (default 20)"
// @Param        operator_id  query string  false  "filter by operator ID"
// @Param        search  query string  false  "search by key name"
// @Success      200  {object}  openapi.AdminAPIKeyListResult  "API keys with operator identity"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Router       /admin/api-keys [get]
func (h *SuperAdminHandler) ListAllKeys(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))

	if limit <= 0 || limit > 100 {
		limit = 20
	}

	operatorID := c.Query("operator_id")
	search := c.Query("search")

	result, err := h.service.ListAllKeys(c.Request.Context(), page, limit, operatorID, search)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, openapi.ErrorResponse{
			Error:   apikeydomain.ErrorCode(err),
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, openapi.AdminAPIKeyListResult{
		Keys:       adminKeysByOperatorToOpenAPI(result.Keys),
		Pagination: apikeydomainPagination(result.Pagination),
	})
}

// GetOperatorKeys handles GET /v1/admin/api-keys/operator/:operatorId.
// @Summary      List operator API keys
// @Description  Lists all API keys for a specific operator (super admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        operatorId  path  string  true  "operator ID"
// @Param        page   query int  false  "page number (default 1)"
// @Param        limit  query int  false  "page size (default 20)"
// @Success      200  {object}  openapi.AdminAPIKeyListResult  "operator API keys"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Router       /admin/api-keys/operator/{operatorId} [get]
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
		c.JSON(status, openapi.ErrorResponse{
			Error:   apikeydomain.ErrorCode(err),
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, openapi.AdminAPIKeyListResult{
		Keys:                 apiKeysToAdminOpenAPI(result.Keys),
		Pagination:           apikeydomainPagination(result.Pagination),
		MonthlyLimit:         result.MonthlyLimit,
		KeysCreatedThisMonth: result.KeysCreatedThisMonth,
	})
}

// ForceRevokeKey handles DELETE /v1/admin/api-keys/:keyId.
// @Summary      Force revoke API key
// @Description  Force revokes an API key for any operator (super admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        keyId  path  string  true  "API key ID"
// @Success      200  {object}  openapi.SuccessResult  "key revoked"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Failure      404  {object}  openapi.ErrorResponse  "key not found"
// @Router       /admin/api-keys/{keyId} [delete]
func (h *SuperAdminHandler) ForceRevokeKey(c *gin.Context) {
	keyID := c.Param("keyId")

	// Get key info for audit before revoking.
	key, err := h.service.GetKey(c.Request.Context(), "", keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, openapi.ErrorResponse{
			Error:   apikeydomain.ErrorCode(err),
			Message: err.Error(),
		})
		return
	}

	err = h.service.ForceRevokeKey(c.Request.Context(), keyID)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, openapi.ErrorResponse{
			Error:   apikeydomain.ErrorCode(err),
			Message: err.Error(),
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

// GetGlobalStats handles GET /v1/admin/api-keys/stats.
// @Summary      Get global API key stats
// @Description  Returns global API key statistics (super admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.GlobalAPIKeyStatsResult  "global stats"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Router       /admin/api-keys/stats [get]
func (h *SuperAdminHandler) GetGlobalStats(c *gin.Context) {
	stats, err := h.service.GetGlobalStats(c.Request.Context())
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, openapi.ErrorResponse{
			Error:   apikeydomain.ErrorCode(err),
			Message: err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetOperatorStats handles GET /v1/admin/api-keys/stats/operator/:operatorId.
// @Summary      Get operator API key stats
// @Description  Returns API key statistics for a specific operator (super admin only)
// @Tags         admin
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        operatorId  path  string  true  "operator ID"
// @Success      200  {object}  openapi.OperatorAPIKeyStatsResult  "operator stats"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "super_admin required"
// @Router       /admin/api-keys/stats/operator/{operatorId} [get]
func (h *SuperAdminHandler) GetOperatorStats(c *gin.Context) {
	operatorID := c.Param("operatorId")

	page := 1
	limit := 20

	result, err := h.service.ListKeys(c.Request.Context(), operatorID, page, limit)
	if err != nil {
		status := apikeydomain.HTTPStatusCode(err)
		c.JSON(status, openapi.ErrorResponse{
			Error:   apikeydomain.ErrorCode(err),
			Message: err.Error(),
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

	c.JSON(http.StatusOK, openapi.OperatorAPIKeyStatsResult{
		OperatorID:           operatorID,
		TotalKeys:            totalKeys,
		ActiveKeys:           activeKeys,
		RevokedKeys:          totalKeys - activeKeys,
		KeysCreatedThisMonth: result.KeysCreatedThisMonth,
		MonthlyLimit:         result.MonthlyLimit,
	})
}
