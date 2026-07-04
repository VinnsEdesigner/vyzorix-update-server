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

	// Extract operator for DOA check
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized", "message": "Operator context required"})
		return
	}

	deviceID := c.Param("imei")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Device ID is required"})
		return
	}

	// DOA check - verify operator owns this device
	_, err := h.devRepo.FindByIDAndOperator(ctx, deviceID, op.ID)
	if err != nil {
		h.logger.Warn("Device not found or not owned", "deviceID", deviceID, "operatorID", op.ID, "error", err)
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
