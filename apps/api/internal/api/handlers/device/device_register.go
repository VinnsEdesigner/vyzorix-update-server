package device

import (
	"errors"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// RegisterHandler handles POST /v1/device/register.
type RegisterHandler struct {
	deviceService *device.Service
}

// NewRegisterHandler creates a new RegisterHandler.
func NewRegisterHandler(deviceService *device.Service) *RegisterHandler {
	return &RegisterHandler{deviceService: deviceService}
}

// Handle processes the device registration request.
func (h *RegisterHandler) Handle(c *gin.Context) {
	var req struct {
		DeviceID          string `json:"deviceId" binding:"required"`
		FirebaseInstallID string `json:"firebaseInstallId" binding:"required"`
		FCMToken          string `json:"fcmToken"`
		AppVersion        string `json:"appVersion"`
		DeviceClass       string `json:"deviceClass"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid JSON in request body"})
		return
	}

	if req.DeviceID == "" || req.FirebaseInstallID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "deviceId and firebaseInstallId are required"})
		return
	}

	dtoReq := &dto.RegisterDeviceRequest{
		DeviceID:          req.DeviceID,
		FirebaseInstallID: req.FirebaseInstallID,
		FCMToken:          req.FCMToken,
		AppVersion:        req.AppVersion,
		DeviceClass:       req.DeviceClass,
	}

	result, err := h.deviceService.Register(c.Request.Context(), dtoReq)
	if err != nil {
		if errors.Is(err, device.ErrDeviceHijack) {
			c.JSON(http.StatusConflict, gin.H{"error": "conflict", "message": "device registration hijack detected"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to register device"})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"device_id":      result.DeviceID,
		"command_secret": result.CommandSecret,
		"registered_at":  result.RegisteredAt,
		"server_time":    time.Now().UnixMilli(),
	})
}
