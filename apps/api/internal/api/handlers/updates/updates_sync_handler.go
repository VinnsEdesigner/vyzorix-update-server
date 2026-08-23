package updates

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.UpdateSyncResponse
	_ openapi.UpdateSyncStatusResult
	_ openapi.ErrorResponse
)

// UpdatesSyncHandler handles sync-related HTTP requests.
type UpdatesSyncHandler struct {
	service     *updates.Service
	auditLogger *audit.Logger
}

// NewUpdatesSyncHandler creates a new UpdatesSyncHandler.
func NewUpdatesSyncHandler(service *updates.Service, auditLogger *audit.Logger) *UpdatesSyncHandler {
	return &UpdatesSyncHandler{
		service:     service,
		auditLogger: auditLogger,
	}
}

// Sync handles POST /v1/updates/sync.
// @Summary      Sync versions from GitHub
// @Description  Triggers a GitHub-release sync (admin only)
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      202  {object}  openapi.UpdateSyncResponse  "sync started"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      403  {object}  openapi.ErrorResponse  "admin required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/sync [post]
func (h *UpdatesSyncHandler) Sync(c *gin.Context) {
	operator := middleware.GetOperatorFromContext(c)
	operatorID := ""
	if operator != nil {
		operatorID = operator.ID
	}

	// Audit log sync start.
	if h.auditLogger != nil {
		h.auditLogger.UpdateSyncStarted(
			c.Request.Context(),
			operatorID,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
	}

	result, err := h.service.SyncFromGitHub(c.Request.Context())
	if err != nil {
		// Audit log sync failure.
		if h.auditLogger != nil {
			h.auditLogger.UpdateSyncFailed(
				c.Request.Context(),
				operatorID,
				c.ClientIP(),
				c.GetHeader("User-Agent"),
				err.Error(),
			)
		}
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Sync failed"))
		return
	}
	c.JSON(http.StatusAccepted, result)
}

// SyncVersions is an alias for Sync to match expected handler names.
func (h *UpdatesSyncHandler) SyncVersions(c *gin.Context) {
	h.Sync(c)
}

// GetSyncStatus handles GET /v1/updates/sync/status.
// @Summary      Get sync status
// @Description  Returns the current GitHub-release sync status
// @Tags         updates
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.UpdateSyncStatusResult  "sync status"
// @Failure      401  {object}  openapi.ErrorResponse  "not authenticated"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /updates/sync/status [get]
func (h *UpdatesSyncHandler) GetSyncStatus(c *gin.Context) {
	status, err := h.service.GetSyncStatus(c.Request.Context())
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			_ = c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to get sync status"))
		return
	}
	c.JSON(http.StatusOK, status)
}
