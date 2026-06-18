package device

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"

	"github.com/gin-gonic/gin"
)

// StatusHandler handles GET /v1/device/:id/status.
type StatusHandler struct {
	deviceService *device.Service
}

// NewStatusHandler creates a new StatusHandler.
func NewStatusHandler(deviceService *device.Service) *StatusHandler {
	return &StatusHandler{deviceService: deviceService}
}

// Handle processes the device status request.
func (h *StatusHandler) Handle(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "device_id is required"})
		return
	}

	status, err := h.deviceService.GetStatus(c.Request.Context(), deviceID)
	if err != nil {
		if err == application.ErrDeviceNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "device_not_found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device status"})
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
