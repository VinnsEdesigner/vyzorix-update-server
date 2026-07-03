package device

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	hub "github.com/VinnsEdesigner/vyzorix/apps/api/internal/ws"

	"github.com/gin-gonic/gin"
)

// ListHandler handles GET /v1/dashboard/devices.
type ListHandler struct {
	deviceService *device.Service
	hub           *hub.Hub
}

// NewListHandler creates a new ListHandler.
func NewListHandler(deviceService *device.Service, hub *hub.Hub) *ListHandler {
	return &ListHandler{
		deviceService: deviceService,
		hub:           hub,
	}
}

// Handle processes the device listing request with pagination.
func (h *ListHandler) Handle(c *gin.Context) {
	ctx := c.Request.Context()

	// Parse pagination parameters.
	limit := 50

	if l := c.Query("limit"); l != "" {
		if parsed, err := strconv.Atoi(l); err == nil && parsed > 0 {
			limit = parsed
			if limit > 100 {
				limit = 100
			}
		}
	}

	offset := 0

	if o := c.Query("offset"); o != "" {
		if parsed, err := strconv.Atoi(o); err == nil && parsed >= 0 {
			offset = parsed
		}
	}

	// Parse online filter.
	var filterOnline *bool

	switch c.Query("online") {
	case "true":
		v := true
		filterOnline = &v
	case "false":
		v := false
		filterOnline = &v
	}

	// Get paginated devices.
	response, err := h.deviceService.List(ctx, limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error", "message": "Invalid request"})
		return
	}

	// Build response with online status from hub.
	type DeviceRow struct {
		DeviceID    string `json:"deviceId"`
		AppVersion  string `json:"appVersion"`
		DeviceClass string `json:"deviceClass"`
		LastSeen    int64  `json:"lastSeen"`
		Online      bool   `json:"online"`
	}

	devices := make([]DeviceRow, 0, len(response.Devices))

	for _, d := range response.Devices {
		// Check if device is online via WebSocket hub.
		isOnline := h.isDeviceOnline(d.ID) || d.Online

		// Apply online filter.
		if filterOnline != nil && isOnline != *filterOnline {
			continue
		}

		devices = append(devices, DeviceRow{
			DeviceID:    d.ID,
			Online:      isOnline,
			LastSeen:    d.LastSeen,
			AppVersion:  d.AppVersion,
			DeviceClass: d.Model, // Use Model as DeviceClass
		})

		// Stop if we have enough results.
		if len(devices) >= limit {
			break
		}
	}

	result := gin.H{"devices": devices}
	if response.Pagination.Total > offset+len(devices) {
		result["nextCursor"] = offset + len(devices)
	}

	result["total"] = response.Pagination.Total

	c.JSON(http.StatusOK, result)
}

// isDeviceOnline checks if a device has an active WebSocket connection via the hub.
func (h *ListHandler) isDeviceOnline(deviceID string) bool {
	if h.hub == nil {
		return false
	}

	return h.hub.Online(deviceID)
}

// ListByOperator handles GET /v1/dashboard/devices?operatorId=<id>.
// Returns all devices for a specific operator.
func (h *ListHandler) ListByOperator(c *gin.Context) {
	ctx := c.Request.Context()

	operatorID := c.Query("operatorId")
	if operatorID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
		return
	}

	devices, err := h.deviceService.ListByOperatorEntity(ctx, operatorID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	type DeviceRow struct {
		ID           string `json:"id"`
		DeviceID     string `json:"deviceId"`
		AppVersion   string `json:"appVersion"`
		DeviceClass  string `json:"deviceClass"`
		Online       bool   `json:"online"`
		LastSeen     int64  `json:"lastSeen"`
		RegisteredAt int64  `json:"registeredAt"`
	}

	result := make([]DeviceRow, len(devices))

	for i, d := range devices {
		isOnline := h.isDeviceOnline(d.ID) || d.Online
		result[i] = DeviceRow{
			ID:           d.ID,
			DeviceID:     d.ID,
			AppVersion:   d.AppVersion,
			DeviceClass:  d.DeviceClass,
			Online:       isOnline,
			LastSeen:     d.LastSeen,
			RegisteredAt: d.RegisteredAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": result,
		"total":   len(result),
	})
}

// GetDevice handles GET /v1/device/:id.
func (h *ListHandler) GetDevice(c *gin.Context) {
	deviceID := c.Param("id")
	if deviceID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request"})
		return
	}

	d, err := h.deviceService.GetDevice(c.Request.Context(), deviceID)
	if err != nil {
		if err == devicedomain.ErrNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})

		return
	}

	isOnline := h.isDeviceOnline(d.ID) || d.Online

	c.JSON(http.StatusOK, gin.H{
		"id":                  d.ID,
		"device_id":           d.ID,
		"firebase_install_id": d.FirebaseInstallID,
		"app_version":         d.AppVersion,
		"device_class":        d.DeviceClass,
		"online":              isOnline,
		"last_seen":           d.LastSeen,
		"registered_at":       d.RegisteredAt,
	})
}

// Count handles GET /v1/device/count.
func (h *ListHandler) Count(c *gin.Context) {
	count, err := h.deviceService.Count(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "internal_error"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":      count,
		"serverTime": c.GetInt64("serverTime"),
	})
}
