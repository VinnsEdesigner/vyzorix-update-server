package device

import (
	"log/slog"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/logs"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/gin-gonic/gin"
)

// LogsHandler handles device logs endpoints.
type LogsHandler struct {
	logsSvc *logs.Service
	devRepo device.Repository
	logger  *slog.Logger
}

// Compile-time reference for the ErrorResponse swaggo annotation target.
var _ openapi.ErrorResponse

// NewLogsHandler creates a new logs handler.
func NewLogsHandler(logsSvc *logs.Service, devRepo device.Repository, logger *slog.Logger) *LogsHandler {
	return &LogsHandler{
		logsSvc: logsSvc,
		devRepo: devRepo,
		logger:  logger,
	}
}

// GetLogs handles GET /v1/device/:imei/logs.
// Returns event logs for a device with cursor-based pagination.
// @Summary      List device logs
// @Description  Returns event logs for a device with cursor-based pagination.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id     path  string  true  "device IMEI"
// @Param        limit  query int     false  "result limit"
// @Param        before query string  false  "pagination cursor"
// @Param        level  query string  false  "log level filter"
// @Success      200  {object}  openapi.DeviceLogListResult  "logs"
// @Failure      400  {object}  openapi.ErrorResponse  "device not found / forbidden"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /devices/{id}/logs [get]
func (h *LogsHandler) GetLogs(c *gin.Context) {
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
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to retrieve logs"))
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
