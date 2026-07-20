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

	// Extract operator for auth check
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Operator context required"})
		return
	}

	// Get organization ID from context
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

	// Verify device belongs to this organization
	_, err := h.devRepo.FindByIDAndOrganization(ctx, deviceID, orgID)
	if err != nil {
		h.logger.Warn("Device not found in organization", "deviceID", deviceID, "organizationID", orgID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Device not found"})
		return
	}

	req := &metrics.GetTelemetryRequest{
		DeviceID: deviceID,
	}

	if st := c.Query("startTime"); st != "" {
		req.StartTime, _ = strconv.ParseInt(st, 10, 64)
	}
	if et := c.Query("endTime"); et != "" {
		req.EndTime, _ = strconv.ParseInt(et, 10, 64)
	}
	if l := c.Query("limit"); l != "" {
		req.Limit, _ = strconv.Atoi(l)
	}

	response, err := h.metricsSvc.GetTelemetry(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get telemetry", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve telemetry"})
		return
	}

	c.JSON(http.StatusOK, response)
}
