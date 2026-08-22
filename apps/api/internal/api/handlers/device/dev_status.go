package device

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.DeviceStatus
	_ openapi.DeviceEventListResult
	_ openapi.DeviceLogListResult
	_ openapi.ErrorResponse
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
// @Summary      Get device status
// @Description  Returns the live status (online, last_seen, app version) for a device.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Success      200  {object}  openapi.DeviceStatus  "device status"
// @Failure      400  {object}  openapi.ErrorResponse  "device not found / forbidden"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /devices/{imei}/status [get]
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
