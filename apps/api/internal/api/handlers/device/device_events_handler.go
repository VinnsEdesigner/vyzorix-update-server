package device

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/event"
	"github.com/gin-gonic/gin"
)

// EventsHandler handles device events API endpoints.
type EventsHandler struct {
	eventRepo event.Repository
	devRepo   device.Repository
	logger    *slog.Logger
}

// NewEventsHandler creates a new EventsHandler.
func NewEventsHandler(eventRepo event.Repository, devRepo device.Repository, logger *slog.Logger) *EventsHandler {
	return &EventsHandler{
		eventRepo: eventRepo,
		devRepo:   devRepo,
		logger:    logger,
	}
}

// findDevice verifies device belongs to organization.
func (h *EventsHandler) findDevice(ctx context.Context, deviceID, orgID string) (*device.Device, error) {
	if h.devRepo == nil {
		//nolint:nilnil // Intentionally returns nil,nil to skip verification when devRepo is unavailable.
		return nil, nil
	}
	return h.devRepo.FindByIDAndOrganization(ctx, deviceID, orgID)
}

// GetEvents handles GET /v1/device/:imei/events.
// Returns event history for a device with filtering and pagination.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Router       /devices/{id}/events [get]
func (h *EventsHandler) GetEvents(c *gin.Context) {
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
	if _, err := h.findDevice(ctx, deviceID, orgID); err != nil {
		h.logger.Warn("Device not found in organization", "deviceID", deviceID, "organizationID", orgID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Device not found"))
		return
	}

	// Parse query parameters.
	filter := &event.EventFilter{
		Limit: 100,
	}

	// Parse event types.
	if types := c.QueryArray("types"); len(types) > 0 {
		for _, t := range types {
			filter.EventTypes = append(filter.EventTypes, event.EventType(t))
		}
	}

	// Parse severities.
	if severities := c.QueryArray("severities"); len(severities) > 0 {
		for _, s := range severities {
			filter.Severities = append(filter.Severities, event.Severity(s))
		}
	}

	// Parse limit.
	if limit := c.Query("limit"); limit != "" {
		if l := parseInt(limit); l > 0 && l <= 500 {
			filter.Limit = l
		}
	}

	// Parse offset.
	if offset := c.Query("offset"); offset != "" {
		if o := parseInt(offset); o > 0 {
			filter.Offset = o
		}
	}

	// Parse time range.
	if startTime := c.Query("startTime"); startTime != "" {
		if t := parseTime(startTime); !t.IsZero() {
			filter.StartTime = t
		}
	}

	if endTime := c.Query("endTime"); endTime != "" {
		if t := parseTime(endTime); !t.IsZero() {
			filter.EndTime = t
		}
	}

	// Query events.
	result, err := h.eventRepo.GetByDevice(ctx, deviceID, filter)
	if err != nil {
		h.logger.Error("Failed to get device events", "deviceID", deviceID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to retrieve events"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":  result.Events,
		"hasMore": result.HasMore,
		"count":   len(result.Events),
	})
}

// GetEventsByType handles GET /v1/events/types/:type.
// Returns events of a specific type across all accessible devices.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Router       /devices/{id}/events/{type} [get]
func (h *EventsHandler) GetEventsByType(c *gin.Context) {
	ctx := c.Request.Context()

	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Operator context required"))
		return
	}

	eventType := c.Param("type")
	if eventType == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Event type is required"))
		return
	}

	filter := &event.EventFilter{
		Limit: 100,
	}

	if limit := c.Query("limit"); limit != "" {
		if l := parseInt(limit); l > 0 && l <= 500 {
			filter.Limit = l
		}
	}

	if offset := c.Query("offset"); offset != "" {
		if o := parseInt(offset); o > 0 {
			filter.Offset = o
		}
	}

	result, err := h.eventRepo.GetByType(ctx, event.EventType(eventType), filter)
	if err != nil {
		h.logger.Error("Failed to get events by type", "eventType", eventType, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to retrieve events"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events":  result.Events,
		"hasMore": result.HasMore,
		"count":   len(result.Events),
	})
}

// GetRecentEvents handles GET /v1/events/recent.
// Returns most recent events across all accessible devices.
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Router       /devices/{id}/events/recent [get]
func (h *EventsHandler) GetRecentEvents(c *gin.Context) {
	ctx := c.Request.Context()

	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Operator context required"))
		return
	}

	limit := 50
	if l := c.Query("limit"); l != "" {
		if parsed := parseInt(l); parsed > 0 && parsed <= 200 {
			limit = parsed
		}
	}

	events, err := h.eventRepo.GetRecent(ctx, limit)
	if err != nil {
		h.logger.Error("Failed to get recent events", "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Failed to retrieve events"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"events": events,
		"count":  len(events),
	})
}

// GetEventByID handles GET /v1/events/:id.
// Returns a single event by ID.
func (h *EventsHandler) GetEventByID(c *gin.Context) {
	ctx := c.Request.Context()

	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeAuthTokenInvalid, "Operator context required"))
		return
	}

	eventID := c.Param("id")
	if eventID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "Event ID is required"))
		return
	}

	evt, err := h.eventRepo.GetByID(ctx, eventID)
	if err != nil {
		h.logger.Warn("Event not found", "eventID", eventID, "error", err)
		_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "Event not found"))
		return
	}

	c.JSON(http.StatusOK, evt)
}

// parseInt parses an int from string.
func parseInt(s string) int {
	var n int
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0
		}
		n = n*10 + int(c-'0')
	}
	return n
}

// parseTime parses a time string (ISO8601 or Unix timestamp).
func parseTime(s string) time.Time {
	// Try Unix milliseconds first.
	if t := parseUnixMs(s); !t.IsZero() {
		return t
	}

	// Try Unix seconds.
	if t := parseUnix(s); !t.IsZero() {
		return t
	}

	// Try RFC3339.
	if t, err := time.Parse(time.RFC3339, s); err == nil {
		return t
	}

	return time.Time{}
}

// parseUnixMs parses Unix milliseconds.
func parseUnixMs(s string) time.Time {
	ms := parseInt(s)
	if ms == 0 {
		return time.Time{}
	}
	return time.UnixMilli(int64(ms))
}

// parseUnix parses Unix seconds.
func parseUnix(s string) time.Time {
	sec := parseInt(s)
	if sec == 0 {
		return time.Time{}
	}
	return time.Unix(int64(sec), 0)
}
