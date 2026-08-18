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
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
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
	dev, err := h.devRepo.FindByIDAndOrganization(ctx, deviceID, orgID)
	if err != nil {
		h.logger.Warn("Device not found in organization", "deviceID", deviceID, "organizationID", orgID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found"))
		return
	}

	req := &metrics.GetMetricsRequest{
		DeviceID:       dev.ID,
		OrganizationID: orgID,
		Range:          c.DefaultQuery("range", "6h"),
		Resolution:     c.Query("resolution"),
	}

	// Max allowed range: 90 days in milliseconds.
	const maxTimeWindowMs = 90 * 24 * 60 * 60 * 1000 // 7,776,000,000 ms.

	if st := c.Query("startTime"); st != "" {
		val, intErr := strconv.ParseInt(st, 10, 64)
		if intErr != nil || val < 0 {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid or negative startTime"))
			return
		}
		req.StartTime = val
	}
	if et := c.Query("endTime"); et != "" {
		val, intErr := strconv.ParseInt(et, 10, 64)
		if intErr != nil || val < 0 {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invalid or negative endTime"))
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

	response, err := h.metricsSvc.GetDeviceMetrics(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get device metrics", "deviceID", deviceID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to retrieve metrics"))
		return
	}

	response.Device.IMEI = deviceID

	c.JSON(http.StatusOK, response)
}

// ExportMetrics handles GET /v1/device/:imei/metrics/export.
// Exports metrics data in JSON or CSV format.
func (h *MetricsHandler) ExportMetrics(c *gin.Context) {
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

	format := c.DefaultQuery("format", "json")

	// Calculate time range from range parameter (default: 24h).
	now := time.Now()
	var startTime, endTime time.Time

	if st := c.Query("startTime"); st != "" {
		startTime = time.UnixMilli(parseInt64(st))
		endTime = now
	} else if et := c.Query("endTime"); et != "" {
		endTime = time.UnixMilli(parseInt64(et))
		startTime = endTime.Add(-24 * time.Hour) // Default to 24h range.
	} else {
		// Default to last 24 hours as per spec.
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
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to export metrics"))
		return
	}

	filename := fmt.Sprintf("metrics_%s_%s.%s", deviceID, time.Now().Format("20060102_150405"), format)

	switch format {
	case "csv":
		c.Header("Content-Type", "text/csv")
		c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=%s", filename))
		h.writeCSV(c, response)
	default:
		// For JSON, still provide as file download.
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
