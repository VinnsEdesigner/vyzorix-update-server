package command

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/gin-gonic/gin"
)

// HistoryHandler handles command history endpoints.
type HistoryHandler struct {
	historySvc *command.HistoryService
	logger     *slog.Logger
}

// NewHistoryHandler creates a new command history handler.
func NewHistoryHandler(historySvc *command.HistoryService, logger *slog.Logger) *HistoryHandler {
	return &HistoryHandler{
		historySvc: historySvc,
		logger:     logger,
	}
}

// GetHistory handles GET /v1/device/:imei/commands.
// Returns paginated command history for a device.
func (h *HistoryHandler) GetHistory(c *gin.Context) {
	ctx := c.Request.Context()

	deviceID := c.Param("imei")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Device ID is required"})
		return
	}

	status := c.Query("status")
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	var startTime, endTime int64
	if st := c.Query("startTime"); st != "" {
		parsed, err := strconv.ParseInt(st, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid startTime format"})
			return
		}
		startTime = parsed
	}
	if et := c.Query("endTime"); et != "" {
		parsed, err := strconv.ParseInt(et, 10, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "Invalid endTime format"})
			return
		}
		endTime = parsed
	}

	req := &command.GetHistoryRequest{
		DeviceID:  deviceID,
		Status:    status,
		Page:      page,
		Limit:     limit,
		StartTime: startTime,
		EndTime:   endTime,
	}

	resp, err := h.historySvc.GetHistory(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get command history", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve command history"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
