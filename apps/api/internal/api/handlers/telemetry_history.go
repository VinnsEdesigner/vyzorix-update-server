package handlers

import (
	"encoding/csv"
	"fmt"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	"github.com/gin-gonic/gin"
)

// TelemetryHistoryConfig holds configuration for telemetry history.
type TelemetryHistoryConfig struct {
	// MaxResults is the maximum number of results to return (default 1000)
	MaxResults int
	// CacheTTL is the cache TTL for recent queries (default 1 hour)
	CacheTTL time.Duration
	// EnableExport enables CSV/JSON export (default true)
	EnableExport bool
}

// DefaultTelemetryHistoryConfig returns the default configuration.
func DefaultTelemetryHistoryConfig() *TelemetryHistoryConfig {
	return &TelemetryHistoryConfig{
		MaxResults:   1000,
		CacheTTL:     1 * time.Hour,
		EnableExport: true,
	}
}

// TelemetryHistoryHandler handles telemetry history requests.
type TelemetryHistoryHandler struct {
	log           *slog.Logger
	telemetryRepo *storage.TelemetryRepository
	config        *TelemetryHistoryConfig
}

// NewTelemetryHistoryHandler creates a new TelemetryHistoryHandler.
func NewTelemetryHistoryHandler(
	log *slog.Logger,
	telemetryRepo *storage.TelemetryRepository,
	cfg *TelemetryHistoryConfig,
) *TelemetryHistoryHandler {
	if cfg == nil {
		cfg = DefaultTelemetryHistoryConfig()
	}

	return &TelemetryHistoryHandler{
		log:           log,
		telemetryRepo: telemetryRepo,
		config:        cfg,
	}
}



// verifyDeviceInOrganization verifies that a device belongs to the given organization.
func (h *TelemetryHistoryHandler) verifyDeviceInOrganization(c *gin.Context, deviceID, orgID string) bool {
	if h.deviceRepo == nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "device repository not available",
		})
		return false
	}

	d, err := h.deviceRepo.FindByIDAndOrganization(c.Request.Context(), deviceID, orgID)
	if err != nil {
		if err == device.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "not_found",
				"message": "device not found in organization",
			})
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "internal_error",
				"message": "failed to verify device",
			})
		}
		return false
	}

	// Verify the device ID matches (deviceID could be IMEI)
	if d.ID != deviceID {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "device not found in organization",
		})
		return false
	}

	return true
}

// QueryHistoryRequest represents a telemetry history query request.
type QueryHistoryRequest struct {
	DeviceID  string `json:"deviceId" form:"deviceId"`
	Format    string `json:"format" form:"format"`
	StartTime int64  `json:"startTime" form:"startTime"`
	EndTime   int64  `json:"endTime" form:"endTime"`
	Limit     int    `json:"limit" form:"limit"`
}

// QueryHistoryResponse represents the response for telemetry history.
type QueryHistoryResponse struct {
	DeviceID   string           `json:"deviceId"`
	Entries    []telemetryEntry `json:"entries"`
	TotalCount int              `json:"totalCount"`
	StartTime  int64            `json:"startTime"`
	EndTime    int64            `json:"endTime"`
	QueryTime  int64            `json:"queryTime"` // Server-side query duration in ms
}

// TelemetryEntry represents a single telemetry entry for the API response.
type telemetryEntry struct {
	ReceivedAt  time.Time `json:"receivedAt"`
	ID          string    `json:"id"`
	DeviceID    string    `json:"deviceId"`
	Payload     string    `json:"payload,omitempty"`
	RiskScore   int       `json:"riskScore,omitempty"`
	BufferLevel int       `json:"bufferLevel,omitempty"`
	ThermalTemp float64   `json:"thermalTemp,omitempty"`
}

// Query handles GET /v1/telemetry/history
// Query telemetry history for a device within a time range.
func (h *TelemetryHistoryHandler) Query(c *gin.Context) {
	// Require organization context for multi-tenant isolation
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "organization context required",
		})
		return
	}

	var req QueryHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "invalid query parameters",
		})
		return
	}

	// Verify device belongs to organization
	if !h.verifyDeviceInOrganization(c, req.DeviceID, orgID) {
		return
	}

	startTime := time.Now()

	// Query telemetry
	entries, err := h.telemetryRepo.ListSince(
		c.Request.Context(),
		req.DeviceID,
		req.StartTime,
		h.config.MaxResults,
	)
	if err != nil {
		h.log.Error("failed to query telemetry", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to query telemetry",
		})
		return
	}

	// Convert domain entries to response entries
	responseEntries := make([]telemetryEntry, len(entries))
	for i, e := range entries {
		responseEntries[i] = telemetryEntry{
			ReceivedAt:  e.ReceivedAt,
			ID:          e.ID,
			DeviceID:    e.DeviceID,
			Payload:     e.Payload,
			RiskScore:   e.RiskScore,
			BufferLevel: e.BufferLevel,
			ThermalTemp: e.ThermalTemp,
		}
	}

	response := QueryHistoryResponse{
		DeviceID:   req.DeviceID,
		Entries:    responseEntries,
		TotalCount: len(entries),
		StartTime:  req.StartTime,
		EndTime:    req.EndTime,
		QueryTime:  time.Since(startTime).Milliseconds(),
	}

	switch req.Format {
	case "csv":
		h.exportCSV(c, response)
		return
	case "json":
	}
	c.JSON(http.StatusOK, response)
}

// exportCSV exports telemetry history as CSV.
func (h *TelemetryHistoryHandler) exportCSV(c *gin.Context, data QueryHistoryResponse) {
	if !h.config.EnableExport {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "CSV export is disabled",
		})

		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", fmt.Sprintf(
		"attachment; filename=telemetry_%s_%d_%d.csv",
		data.DeviceID,
		data.StartTime,
		data.EndTime,
	))

	writer := csv.NewWriter(c.Writer)
	defer writer.Flush()

	// Write header
	_ = writer.Write([]string{
		"id",
		"device_id",
		"received_at",
		"risk_score",
		"buffer_level",
		"thermal_temp",
	})

	// Write data
	for _, e := range data.Entries {
		_ = writer.Write([]string{
			e.ID,
			e.DeviceID,
			e.ReceivedAt.Format(time.RFC3339),
			strconv.Itoa(e.RiskScore),
			strconv.Itoa(e.BufferLevel),
			fmt.Sprintf("%.2f", e.ThermalTemp),
		})
	}
}

// ExportJSON exports telemetry history as JSON file.
func (h *TelemetryHistoryHandler) ExportJSON(c *gin.Context) {
	if !h.config.EnableExport {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "JSON export is disabled",
		})

		return
	}

	var req QueryHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "invalid query parameters",
		})

		return
	}

	// Query telemetry
	entries, err := h.telemetryRepo.ListSince(
		c.Request.Context(),
		req.DeviceID,
		req.StartTime,
		h.config.MaxResults,
	)
	if err != nil {
		h.log.Error("failed to query telemetry for export", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to query telemetry",
		})

		return
	}

	c.Header("Content-Type", "application/json")
	c.Header("Content-Disposition", fmt.Sprintf(
		"attachment; filename=telemetry_%s_%d_%d.json",
		req.DeviceID,
		req.StartTime,
		req.EndTime,
	))

	c.JSON(http.StatusOK, gin.H{
		"deviceId":   req.DeviceID,
		"exportedAt": time.Now().Format(time.RFC3339),
		"count":      len(entries),
		"entries":    entries,
	})
}

// GetLatest handles GET /v1/telemetry/latest/:deviceId
// Gets the latest telemetry entry for a device.
func (h *TelemetryHistoryHandler) GetLatest(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "deviceId is required",
		})

		return
	}

	entries, err := h.telemetryRepo.List(c.Request.Context(), deviceID, 1)
	if err != nil {
		h.log.Error("failed to get latest telemetry", "err", err, "deviceId", deviceID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to get latest telemetry",
		})

		return
	}

	if len(entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "no telemetry found for device",
		})

		return
	}

	e := entries[0]
	c.JSON(http.StatusOK, gin.H{
		"id":          e.ID,
		"deviceId":    e.DeviceID,
		"receivedAt":  e.ReceivedAt,
		"riskScore":   e.RiskScore,
		"bufferLevel": e.BufferLevel,
		"thermalTemp": e.ThermalTemp,
	})
}

// GetStats handles GET /v1/telemetry/stats/:deviceId
// Gets telemetry statistics for a device.
func (h *TelemetryHistoryHandler) GetStats(c *gin.Context) {
	deviceID := c.Param("deviceId")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "deviceId is required",
		})

		return
	}

	// Get last 100 entries for stats calculation
	entries, err := h.telemetryRepo.List(c.Request.Context(), deviceID, 100)
	if err != nil {
		h.log.Error("failed to get telemetry for stats", "err", err, "deviceId", deviceID)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to get telemetry stats",
		})

		return
	}

	if len(entries) == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "no telemetry found for device",
		})

		return
	}

	// Calculate statistics
	var totalRisk, totalBuffer, totalTemp int

	minRisk, maxRisk := 999, -1
	minTemp, maxTemp := 999.0, -999.0

	for _, e := range entries {
		totalRisk += e.RiskScore
		totalBuffer += e.BufferLevel
		totalTemp += int(e.ThermalTemp * 100)

		if e.RiskScore < minRisk {
			minRisk = e.RiskScore
		}

		if e.RiskScore > maxRisk {
			maxRisk = e.RiskScore
		}

		if e.ThermalTemp < minTemp {
			minTemp = e.ThermalTemp
		}

		if e.ThermalTemp > maxTemp {
			maxTemp = e.ThermalTemp
		}
	}

	count := len(entries)
	// Safety: should never happen due to early return, but double-check
	if count == 0 {
		c.JSON(http.StatusNotFound, gin.H{
			"error":   "not_found",
			"message": "no telemetry found for device",
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"deviceId":    deviceID,
		"sampleCount": count,
		"latestEntry": entries[0].ReceivedAt,
		"oldestEntry": entries[count-1].ReceivedAt,
		"riskScore": gin.H{
			"avg": float64(totalRisk) / float64(count),
			"min": minRisk,
			"max": maxRisk,
		},
		"bufferLevel": gin.H{
			"avg": float64(totalBuffer) / float64(count),
		},
		"thermalTemp": gin.H{
			"avg": float64(totalTemp) / float64(count) / 100,
			"min": minTemp,
			"max": maxTemp,
		},
	})
}

// CleanupOld handles DELETE /v1/telemetry/cleanup
// Cleans up telemetry older than the specified timestamp.
func (h *TelemetryHistoryHandler) CleanupOld(c *gin.Context) {
	// Require admin or operator role
	op := middleware.GetOperatorFromContext(c)
	if op == nil || (!op.IsAdmin() && !op.IsOperator()) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "authentication required",
		})
		return
	}

	orgID := middleware.GetOrganizationID(c)
	if orgID == "" || !op.IsAdminIn(orgID) {
		c.JSON(http.StatusForbidden, gin.H{
			"error":   "forbidden",
			"message": "admin role required",
		})
		return
	}

	var req struct {
		OlderThan int64 `json:"olderThan" form:"olderThan"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "olderThan timestamp required",
		})

		return
	}

	if req.OlderThan <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "bad_request",
			"message": "olderThan must be a positive timestamp",
		})

		return
	}

	deleted, err := h.telemetryRepo.DeleteOlderThan(c.Request.Context(), req.OlderThan)
	if err != nil {
		h.log.Error("failed to cleanup old telemetry", "err", err)
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "internal_error",
			"message": "failed to cleanup old telemetry",
		})

		return
	}

	h.log.Info("telemetry cleanup completed",
		"deletedCount", deleted,
		"operatorId", op.ID,
	)

	c.JSON(http.StatusOK, gin.H{
		"deleted":   deleted,
		"olderThan": req.OlderThan,
	})
}
