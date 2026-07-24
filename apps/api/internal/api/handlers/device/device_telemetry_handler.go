package device

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
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
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Operator context required"})
		return
	}

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "organization context required"})
		return
	}

	deviceID := c.Param("imei")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Device ID is required"})
		return
	}

	// Verify device belongs to this organization.
	_, err := h.devRepo.FindByIDAndOrganization(ctx, deviceID, orgID)
	if err != nil {
		h.logger.Warn("Device not found in organization", "deviceID", deviceID, "organizationID", orgID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Device not found"})
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
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid startTime format"})
			return
		}
		if val < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "startTime must be >= 0"})
			return
		}
		req.StartTime = val
	}
	if et := c.Query("endTime"); et != "" {
		val, intErr := strconv.ParseInt(et, 10, 64)
		if intErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid endTime format"})
			return
		}
		req.EndTime = val
	}
	// Enforce max query window after both values are parsed.
	if req.StartTime > 0 && req.EndTime > 0 {
		if req.EndTime <= req.StartTime {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "endTime must be greater than startTime"})
			return
		}
		if req.EndTime-req.StartTime > maxTimeWindowMs {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "time range exceeds maximum allowed window of 90 days"})
			return
		}
	}
	if l := c.Query("limit"); l != "" {
		val, intErr := strconv.Atoi(l)
		if intErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "invalid limit format"})
			return
		}
		req.Limit = val
	}

	response, err := h.metricsSvc.GetTelemetry(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get telemetry", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve telemetry"})
		return
	}

	c.JSON(http.StatusOK, response)
}
