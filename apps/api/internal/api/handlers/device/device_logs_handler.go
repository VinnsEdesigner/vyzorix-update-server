package device

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/gin-gonic/gin"
)

// LogsHandler handles device logs endpoints.
type LogsHandler struct {
	logsSvc *logs.Service
	devRepo device.Repository
	logger  *slog.Logger
}

// NewLogsHandler creates a new logs handler.
func NewLogsHandler(logsSvc *logs.Service, devRepo device.Repository, logger *slog.Logger) *LogsHandler {
	return &LogsHandler{
		logsSvc: logsSvc,
		devRepo: devRepo,
		logger:  logger,
	}
}

// GetLogs handles GET /v1/device/:id/logs.
// Returns event logs for a device with cursor-based pagination.
func (h *LogsHandler) GetLogs(c *gin.Context) {
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

	eventType := c.Query("type")
	limit := 100
	if l := c.Query("limit"); l != "" {
		if parsed := parseInt64(l); parsed > 0 {
			limit = int(parsed)
		}
	}
	cursor := c.Query("cursor")

	var startTime, endTime int64
	if st := c.Query("startTime"); st != "" {
		startTime = parseInt64(st)
	}
	if et := c.Query("endTime"); et != "" {
		endTime = parseInt64(et)
	}

	req := &logs.ListLogsRequest{
		DeviceID:  deviceID,
		EventType: eventType,
		StartTime: startTime,
		EndTime:   endTime,
		Limit:     limit,
		Cursor:    cursor,
	}

	response, err := h.logsSvc.GetDeviceLogs(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get device logs", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve logs"})
		return
	}

	c.JSON(http.StatusOK, response)
}

// parseInt64 parses an int64 from string.
func parseInt64(s string) int64 {
	var n int64
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int64(c-'0')
	}
	return n
}
