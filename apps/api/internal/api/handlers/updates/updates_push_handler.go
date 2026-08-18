package updates

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/audit"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// UpdatesPushHandler handles push-related HTTP requests.
type UpdatesPushHandler struct {
	service     *updates.Service
	auditLogger *audit.Logger
}

// NewUpdatesPushHandler creates a new UpdatesPushHandler.
func NewUpdatesPushHandler(service *updates.Service, auditLogger *audit.Logger) *UpdatesPushHandler {
	return &UpdatesPushHandler{
		service:     service,
		auditLogger: auditLogger,
	}
}

// PushUpdate handles POST /v1/updates/push.
func (h *UpdatesPushHandler) PushUpdate(c *gin.Context) {
	var req struct {
		ScheduledAt *int64   `json:"scheduledAt,omitempty"`
		Version     string   `json:"version" binding:"required"`
		InstallType string   `json:"installType" binding:"required,oneof=immediate scheduled"`
		DeviceIDs   []string `json:"deviceIds" binding:"required,min=1"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, err.Error()))
		return
	}

	operator := middleware.GetOperatorFromContext(c)
	if operator == nil {
		c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Operator not found"))
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	pushReq := &updates.PushUpdateRequest{
		Version:        req.Version,
		DeviceIDs:      req.DeviceIDs,
		InstallType:    req.InstallType,
		ScheduledAt:    req.ScheduledAt,
		OrganizationID: orgID,
	}

	pushResp, err := h.service.PushUpdate(c.Request.Context(), pushReq, operator.ID)
	if err != nil {
		if se := updates.AsServiceError(err); se != nil {
			c.Error(apperrors.NewServerErrorFromStatus(se.Status, se.Message))
			return
		}
		c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to push update"))
		return
	}

	// Audit log the push.
	if h.auditLogger != nil {
		h.auditLogger.UpdatePushed(
			c.Request.Context(),
			operator.ID,
			pushResp.PushID,
			pushResp.Version,
			pushResp.Devices.Total,
			c.ClientIP(),
			c.GetHeader("User-Agent"),
		)
	}

	c.JSON(http.StatusAccepted, pushResp)
}
