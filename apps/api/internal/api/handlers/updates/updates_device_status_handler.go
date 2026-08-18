package updates

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/updates"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	domainupdates "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
	"github.com/gin-gonic/gin"
)

// DeviceStatusHandler handles device callbacks for update status.
type DeviceStatusHandler struct {
	service *updates.PushService
}

// NewDeviceStatusHandler creates a new device status handler.
func NewDeviceStatusHandler(service *updates.PushService) *DeviceStatusHandler {
	return &DeviceStatusHandler{service: service}
}

// DeviceUpdateStatusRequest represents device callback for update status.
type DeviceUpdateStatusRequest struct {
	DispatchID string `json:"dispatchId" binding:"required"`
	DeviceID   string `json:"deviceId" binding:"required"`
	Status     string `json:"status" binding:"required"`
	Error      string `json:"error,omitempty"`
}

// DeviceUpdateStatusResponse represents response to device callback.
type DeviceUpdateStatusResponse struct {
	Message      string `json:"message,omitempty"`
	Acknowledged bool   `json:"acknowledged"`
}

// HandleDeviceUpdateStatus handles POST /v1/updates/device-status.
// This endpoint is called by devices to report update progress.
// The device_id is included in the request body since the device knows its own ID.
// from the FCM payload (device_id field).
func (h *DeviceStatusHandler) HandleDeviceUpdateStatus(c *gin.Context) {
	var req DeviceUpdateStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid request"))
		return
	}

	// Validate status value.
	validStatuses := map[string]domainupdates.DevicePushStatus{
		"in_progress": domainupdates.DevicePushStatusInProgress,
		"completed":   domainupdates.DevicePushStatusCompleted,
		"failed":      domainupdates.DevicePushStatusFailed,
	}

	status, ok := validStatuses[req.Status]
	if !ok {
		c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid status. Must be: in_progress, completed, or failed"))
		return
	}

	// Update device status via service.
	err := h.service.UpdateDeviceStatusByDispatch(c.Request.Context(), req.DispatchID, req.DeviceID, status, req.Error)
	if err != nil {
		if err == domainupdates.ErrPushNotFound {
			c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Push not found"))
			return
		}
		if err == domainupdates.ErrDeviceNotFound {
			c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found in push"))
			return
		}
		c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to update status"))
		return
	}

	c.JSON(http.StatusOK, DeviceUpdateStatusResponse{
		Acknowledged: true,
		Message:      "Status updated",
	})
}
