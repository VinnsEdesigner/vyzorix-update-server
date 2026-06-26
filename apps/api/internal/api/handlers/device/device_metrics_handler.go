package device

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/metrics"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/gin-gonic/gin"
)

// MetricsHandler handles device metrics endpoints.
type MetricsHandler struct {
	metricsSvc *metrics.Service
	devRepo    device.Repository
	logger     *slog.Logger
}

// NewMetricsHandler creates a new metrics handler.
func NewMetricsHandler(metricsSvc *metrics.Service, devRepo device.Repository, logger *slog.Logger) *MetricsHandler {
	return &MetricsHandler{
		metricsSvc: metricsSvc,
		devRepo:    devRepo,
		logger:     logger,
	}
}

// GetMetrics handles GET /v1/device/:id/metrics.
// Returns aggregated metrics for chart visualization.
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Device ID is required"})
		return
	}

	dev, err := h.devRepo.FindByID(ctx, deviceID)
	if err != nil {
		h.logger.Warn("Device not found", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Device not found"})
		return
	}

	req := &metrics.GetMetricsRequest{
		DeviceID:   dev.ID,
		Range:      c.DefaultQuery("range", "6h"),
		Resolution: c.Query("resolution"),
	}

	if st := c.Query("startTime"); st != "" {
		req.StartTime, _ = strconv.ParseInt(st, 10, 64)
	}
	if et := c.Query("endTime"); et != "" {
		req.EndTime, _ = strconv.ParseInt(et, 10, 64)
	}

	response, err := h.metricsSvc.GetDeviceMetrics(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get device metrics", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve metrics"})
		return
	}

	response.Device.IMEI = deviceID

	c.JSON(http.StatusOK, response)
}

// ExportMetrics handles GET /v1/device/:id/metrics/export.
// Exports metrics data in JSON or CSV format.
func (h *MetricsHandler) ExportMetrics(c *gin.Context) {
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

	format := c.DefaultQuery("format", "json")

	req := &metrics.GetTelemetryRequest{
		DeviceID: deviceID,
	}

	if st := c.Query("startTime"); st != "" {
		req.StartTime, _ = strconv.ParseInt(st, 10, 64)
	}
	if et := c.Query("endTime"); et != "" {
		req.EndTime, _ = strconv.ParseInt(et, 10, 64)
	}

	limit := 10000
	if l := c.Query("limit"); l != "" {
		if parsed, parseErr := strconv.Atoi(l); parseErr == nil && parsed > 0 {
			limit = parsed
		}
	}
	req.Limit = limit

	response, err := h.metricsSvc.GetTelemetry(ctx, req)
	if err != nil {
		h.logger.Error("Failed to export metrics", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to export metrics"})
		return
	}

	filename := fmt.Sprintf("metrics_%s_%s.%s", deviceID, time.Now().Format("20060102_150405"), format)

	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		h.writeCSV(c, response)
	default:
		c.JSON(http.StatusOK, response)
	}
}

// writeCSV writes telemetry data as CSV.
func (h *MetricsHandler) writeCSV(c *gin.Context, data *metrics.GetTelemetryResponse) {
	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	_ = writer.Write([]string{"timestamp", "riskScore", "thermalTemp", "bufferLevel", "uptime"})

	for _, frame := range data.Frames {
		_ = writer.Write([]string{
			strconv.FormatInt(frame.Timestamp, 10),
			strconv.FormatFloat(frame.RiskScore, 'f', 2, 64),
			strconv.FormatFloat(frame.ThermalTemp, 'f', 2, 64),
			strconv.FormatFloat(frame.BufferLevel, 'f', 2, 64),
			strconv.FormatInt(frame.Uptime, 10),
		})
	}
}
