package device

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// TelemetryHandler handles device telemetry endpoints.
type TelemetryHandler struct {
	metricsSvc *metrics.Service
	devRepo    device.Repository
	logger     *slog.Logger
}

// NewTelemetryHandler creates a new telemetry handler.
func NewTelemetryHandler(metricsSvc *metrics.Service, devRepo device.Repository, logger *slog.Logger) *TelemetryHandler {
	return &TelemetryHandler{
		metricsSvc: metricsSvc,
		devRepo:    devRepo,
		logger:     logger,
	}
}

// GetTelemetry handles GET /v1/device/:imei/telemetry.
// Returns raw telemetry frames.
func (h *TelemetryHandler) GetTelemetry(c *gin.Context) {
	ctx := c.Request.Context()

	// Extract operator for auth check.
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Operator context required"))
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context required"))
		return
	}

	deviceID := c.Param("imei")
	if deviceID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Device ID is required"))
		return
	}

	// Verify device belongs to this organization.
	_, err := h.devRepo.FindByIDAndOrganization(ctx, deviceID, orgID)
	if err != nil {
		h.logger.Warn("Device not found in organization", "deviceID", deviceID, "organizationID", orgID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found"))
		return
	}

	req := &metrics.GetTelemetryRequest{
		DeviceID: deviceID,
	}

	// Max allowed range: 90 days in milliseconds.
	const maxTimeWindowMs = 90 * 24 * 60 * 60 * 1000 // 7,776,000,000 ms.

	if st := c.Query("startTime"); st != "" {
		val, intErr := strconv.ParseInt(st, 10, 64)
		if intErr != nil {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid startTime format"))
			return
		}
		if val < 0 {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "startTime must be >= 0"))
			return
		}
		req.StartTime = val
	}
	if et := c.Query("endTime"); et != "" {
		val, intErr := strconv.ParseInt(et, 10, 64)
		if intErr != nil {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid endTime format"))
			return
		}
		req.EndTime = val
	}
	// Enforce max query window after both values are parsed.
	if req.StartTime > 0 && req.EndTime > 0 {
		if req.EndTime <= req.StartTime {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "endTime must be greater than startTime"))
			return
		}
		if req.EndTime-req.StartTime > maxTimeWindowMs {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "time range exceeds maximum allowed window of 90 days"))
			return
		}
	}
	if l := c.Query("limit"); l != "" {
		val, intErr := strconv.Atoi(l)
		if intErr != nil {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid limit format"))
			return
		}
		req.Limit = val
	}

	response, err := h.metricsSvc.GetTelemetry(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get telemetry", "deviceID", deviceID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to retrieve telemetry"))
		return
	}

	c.JSON(http.StatusOK, response)
}
