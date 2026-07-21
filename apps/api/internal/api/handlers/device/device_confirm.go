package device

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	"github.com/gin-gonic/gin"
)

// InboxCleanup defines the interface for cleaning up inbox entries after device confirmation.
type InboxCleanup interface {
	// DeleteByIMEI deletes all inbox entries for a given IMEI.
	DeleteByIMEI(ctx context.Context, imei string) error
}

// ConfirmHandler handles POST /v1/device/confirm.
type ConfirmHandler struct {
	deviceService *device.Service
	inboxCleanup  InboxCleanup
	logger        *slog.Logger
}

// NewConfirmHandler creates a new ConfirmHandler.
func NewConfirmHandler(deviceService *device.Service) *ConfirmHandler {
	return &ConfirmHandler{
		deviceService: deviceService,
		logger:        slog.Default(),
	}
}

// NewConfirmHandlerWithCleanup creates a new ConfirmHandler with inbox cleanup capability.
func NewConfirmHandlerWithCleanup(deviceService *device.Service, inboxCleanup InboxCleanup) *ConfirmHandler {
	return &ConfirmHandler{
		deviceService: deviceService,
		inboxCleanup:  inboxCleanup,
		logger:        slog.Default(),
	}
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

		if errors.Is(err, device.ErrDeviceAlreadyConfirmed) {
			c.JSON(http.StatusConflict, gin.H{
				"error":   "already_confirmed",
				"message": "Device has already been confirmed. The command secret is single-use.",
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "Failed to confirm device registration",
		})
		return
	}

	// Clean up inbox entry after successful confirmation
	// This prevents inbox records from accumulating forever after device registration completes
	if h.inboxCleanup != nil {
		if err := h.inboxCleanup.DeleteByIMEI(c.Request.Context(), req.IMEI); err != nil {
			// Log but don't fail the request - confirmation was successful
			h.logger.Warn("failed to clean up inbox entry after device confirmation",
				"imei", req.IMEI,
				"error", err,
			)
		} else {
			h.logger.Info("cleaned up inbox entry after device confirmation",
				"imei", req.IMEI,
			)
		}
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
