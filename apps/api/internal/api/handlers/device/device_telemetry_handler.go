package device

import (
	"log/slog"
	"net/http"
	"strconv"

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

// GetTelemetry handles GET /v1/device/:id/telemetry.
// Returns raw telemetry frames.
func (h *TelemetryHandler) GetTelemetry(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Device ID is required"})
		return
	}

	_, err := h.devRepo.FindByID(ctx, deviceID)
	if err != nil {
		h.logger.Warn("Device not found", "deviceID", deviceID, "error", err)
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
