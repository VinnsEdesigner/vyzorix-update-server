package command

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/gin-gonic/gin"
)

// HistoryHandler handles command history endpoints.
type HistoryHandler struct {
	historySvc *command.HistoryService
	devRepo    device.Repository
	logger     *slog.Logger
}

// NewHistoryHandler creates a new command history handler.
func NewHistoryHandler(historySvc *command.HistoryService, devRepo device.Repository, logger *slog.Logger) *HistoryHandler {
	return &HistoryHandler{
		historySvc: historySvc,
		devRepo:    devRepo,
		logger:     logger,
	}
}

// GetHistory handles GET /v1/device/:imei/commands.
// Returns paginated command history for a device.
func (h *HistoryHandler) GetHistory(c *gin.Context) {
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

	status := c.Query("status")
	page := 1
	limit := 20

	if p := c.Query("page"); p != "" {
		if parsed, err := strconv.Atoi(p); err == nil && parsed > 0 {
			page = parsed
		}
	}

	if l := c.Query("limit"); l != "" {
		parsed, err := strconv.Atoi(l)
		if err != nil || parsed <= 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "limit must be a positive integer"})
			return
		}
		if parsed > 100 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "limit cannot exceed 100"})
			return
		}
		limit = parsed
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
	if startTime > 0 && endTime > 0 && startTime >= endTime {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "startTime must be before endTime"})
		return
	}

	req := &command.GetHistoryRequest{
		DeviceID:       deviceID,
		OrganizationID: orgID,
		Status:         status,
		Page:           page,
		Limit:          limit,
		StartTime:      startTime,
		EndTime:        endTime,
	}

	resp, err := h.historySvc.GetHistory(ctx, req)
	if err != nil {
		h.logger.Error("Failed to get command history", "deviceID", deviceID, "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Failed to retrieve command history"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
