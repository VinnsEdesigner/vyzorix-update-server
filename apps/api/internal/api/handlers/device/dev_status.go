package device

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// StatusHandler handles GET /v1/device/:imei/status.
type StatusHandler struct {
	deviceService *device.Service
}

// NewStatusHandler creates a new StatusHandler.
func NewStatusHandler(deviceService *device.Service) *StatusHandler {
	return &StatusHandler{deviceService: deviceService}
}

// Handle processes the device status request.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Router       /devices/status [get]
func (h *StatusHandler) Handle(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "device ID is required"))
		return
	}

	status, err := h.deviceService.GetStatus(c.Request.Context(), imei)
	if err != nil {
		if err == application.ErrDeviceNotFound {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "device not found"))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to retrieve device status"))

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_id":    status.DeviceID,
		"online":       status.Online,
		"last_seen":    status.LastSeen,
		"app_version":  status.AppVersion,
		"device_class": status.DeviceClass,
	})
}
