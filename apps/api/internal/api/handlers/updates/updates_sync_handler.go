package updates

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	"github.com/gin-gonic/gin"
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
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Sync failed",
		})
		return
	}
	c.JSON(http.StatusAccepted, result)
}

// SyncVersions is an alias for Sync to match expected handler names.
func (h *UpdatesSyncHandler) SyncVersions(c *gin.Context) {
	h.Sync(c)
}

// GetSyncStatus handles GET /v1/updates/sync/status.
func (h *UpdatesSyncHandler) GetSyncStatus(c *gin.Context) {
	status, err := h.service.GetSyncStatus(c.Request.Context())
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			c.JSON(se.Status, se.ToErrorResponse())
			return
		}
		c.JSON(http.StatusInternalServerError, updates.ErrorResponse{
			Code:    "internal_error",
			Message: "Failed to get sync status",
		})
		return
	}
	c.JSON(http.StatusOK, status)
}
