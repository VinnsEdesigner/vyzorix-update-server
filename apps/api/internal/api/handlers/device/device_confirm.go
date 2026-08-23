package device

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.DeviceConfirmRequest
	_ openapi.DeviceConfirmResult
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

// Handle handles POST /v1/device/confirm.
// @Summary      Confirm device registration
// @Description  Confirms a device registration from an approved inbox entry
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        body  body  openapi.DeviceConfirmRequest  true  "confirmation request (imei + commandSecret)"
// @Success      200  {object}  openapi.DeviceConfirmResult  "confirmed device"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      404  {object}  openapi.ErrorResponse  "device/inbox entry not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /device/confirm [post]
func (h *ConfirmHandler) Handle(c *gin.Context) {
	var req struct {
		IMEI          string `json:"imei" binding:"required"`
		CommandSecret string `json:"commandSecret" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Invalid JSON in request body: imei and commandSecret are required"))

		return
	}

	if req.IMEI == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))

		return
	}

	if req.CommandSecret == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "commandSecret is required"))

		return
	}

	d, err := h.deviceService.ConfirmDevice(c.Request.Context(), req.IMEI, req.CommandSecret)
	if err != nil {
		if errors.Is(err, device.ErrDeviceNotFound) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found. Registration may not have been approved yet."))

			return
		}

		if errors.Is(err, device.ErrInvalidCommandSecret) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Invalid command secret"))

			return
		}

		if errors.Is(err, device.ErrCommandSecretNotSet) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Device command secret not set. Registration may not have been approved yet."))

			return
		}

		if errors.Is(err, device.ErrDeviceAlreadyConfirmed) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceConflict, "Device has already been confirmed. The command secret is single-use."))

			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to confirm device registration"))

		return
	}

	// Clean up inbox entry after successful confirmation.
	// This prevents inbox records from accumulating forever after device registration completes.
	if h.inboxCleanup != nil {
		if err := h.inboxCleanup.DeleteByIMEI(c.Request.Context(), req.IMEI); err != nil {
			// Log but don't fail the request - confirmation was successful.
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
		"device_id":     d.ID,
		"imei":          d.ID,
		"confirmed":     true,
		"online":        d.Online,
		"registered_at": d.RegisteredAt,
		"server_time":   time.Now().UnixMilli(),
	})
}
