package device

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
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

// GetMetrics handles GET /v1/device/:imei/metrics.
// Returns aggregated metrics for chart visualization.
func (h *MetricsHandler) GetMetrics(c *gin.Context) {
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
	dev, err := h.devRepo.FindByIDAndOrganization(ctx, deviceID, orgID)
	if err != nil {
		h.logger.Warn("Device not found in organization", "deviceID", deviceID, "organizationID", orgID, "error", err)
		c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "Device not found"})
		return
	}

	req := &metrics.GetMetricsRequest{
		DeviceID:       dev.ID,
		OrganizationID: orgID,
		Range:          c.DefaultQuery("range", "6h"),
		Resolution:     c.Query("resolution"),
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

// ExportMetrics handles GET /v1/device/:imei/metrics/export.
// Exports metrics data in JSON or CSV format.
func (h *MetricsHandler) ExportMetrics(c *gin.Context) {
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

	format := c.DefaultQuery("format", "json")

	// Calculate time range from range parameter (default: 24h)
	now := time.Now()
	var startTime, endTime time.Time
	
	if st := c.Query("startTime"); st != "" {
		startTime = time.UnixMilli(parseInt64(st))
		endTime = now
	} else if et := c.Query("endTime"); et != "" {
		endTime = time.UnixMilli(parseInt64(et))
		startTime = endTime.Add(-24 * time.Hour) // Default to 24h range
	} else {
		// Default to last 24 hours as per spec
		endTime = now
		startTime = endTime.Add(-24 * time.Hour)
	}

	req := &metrics.GetTelemetryRequest{
		DeviceID:  deviceID,
		StartTime: startTime.UnixMilli(),
		EndTime:   endTime.UnixMilli(),
		Limit:     10000,
	}

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
		// For JSON, still provide as file download
		c.Header("Content-Type", "application/json")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
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
