package device

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"

	"github.com/gin-gonic/gin"
)

// UpdaterHandler handles device update endpoints.
type UpdaterHandler struct {
	deviceService *device.Service
}

// NewUpdaterHandler creates a new UpdaterHandler.
func NewUpdaterHandler(deviceService *device.Service) *UpdaterHandler {
	return &UpdaterHandler{deviceService: deviceService}
}

// UpdateFCMToken handles PATCH /v1/device/:imei/fcm-token.
func (h *UpdaterHandler) UpdateFCMToken(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "device ID is required"})
		return
	}

	var req struct {
		FCMToken string `json:"fcmToken" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid request"})
		return
	}

	err := h.deviceService.UpdateFCMToken(c.Request.Context(), imei, req.FCMToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to update FCM token"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// Delete handles DELETE /v1/device/:imei.
func (h *UpdaterHandler) Delete(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "device ID is required"})
		return
	}

	err := h.deviceService.Delete(c.Request.Context(), imei)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "failed to delete device"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
