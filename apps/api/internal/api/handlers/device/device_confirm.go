package device

import (
	"errors"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/gin-gonic/gin"
)

// ConfirmHandler handles POST /v1/device/confirm.
type ConfirmHandler struct {
	deviceService *device.Service
}

// NewConfirmHandler creates a new ConfirmHandler.
func NewConfirmHandler(deviceService *device.Service) *ConfirmHandler {
	return &ConfirmHandler{deviceService: deviceService}
}

// Handle processes the device confirmation request.
// Device calls this after receiving the commandSecret via FCM to finalize registration.
func (h *ConfirmHandler) Handle(c *gin.Context) {
	var req struct {
		IMEI          string `json:"imei" binding:"required"`
		CommandSecret string `json:"commandSecret" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "Invalid JSON in request body: imei and commandSecret are required",
		})
		return
	}

	if req.IMEI == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "IMEI is required",
		})
		return
	}

	if req.CommandSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "commandSecret is required",
		})
		return
	}

	d, err := h.deviceService.ConfirmDevice(c.Request.Context(), req.IMEI, req.CommandSecret)
	if err != nil {
		if errors.Is(err, device.ErrDeviceNotFound) {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "Device not found. Registration may not have been approved yet.",
			})
			return
		}

		if errors.Is(err, device.ErrInvalidCommandSecret) {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid command secret",
			})
			return
		}

		if errors.Is(err, device.ErrCommandSecretNotSet) {
			c.JSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "Device command secret not set. Registration may not have been approved yet.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to confirm device registration",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_id":      d.ID,
		"imei":           d.ID,
		"confirmed":       true,
		"online":         d.Online,
		"registered_at":   d.RegisteredAt,
		"server_time":     time.Now().UnixMilli(),
	})
}
