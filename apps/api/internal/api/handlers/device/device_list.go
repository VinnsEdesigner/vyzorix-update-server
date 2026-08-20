package device

import (
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
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

	// Get organization ID from context.
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context is required"))
		return
	}

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

	// Get paginated devices filtered by organization.
	response, err := h.deviceService.ListByOrganization(ctx, orgID, limit, offset)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "Invalid request"))
		return
	}

	// Build response with online status from hub.
	type DeviceRow struct {
		ID             string `json:"id"`
		Imei           string `json:"imei"`
		OrganizationID string `json:"organization_id,omitempty"`
		DeviceName     string `json:"device_name"`
		Model          string `json:"model"`
		Manufacturer   string `json:"manufacturer"`
		AppVersion     string `json:"app_version"`
		Status         string `json:"status"`
		LastSeen       int64  `json:"last_seen"`
		RegisteredAt   int64  `json:"registered_at"`
	}

	devices := make([]DeviceRow, 0, len(response.Devices))

	for _, d := range response.Devices {
		// Check if device is online via WebSocket hub.
		isOnline := h.isDeviceOnline(d.ID) || d.Online

		// Apply online filter.
		if filterOnline != nil && isOnline != *filterOnline {
			continue
		}

		// Determine status string based on online state and lifecycle.
		status := "offline"
		if d.DeregisteredAt != nil {
			status = "deregistered"
		} else if isOnline {
			status = "online"
		}

		devices = append(devices, DeviceRow{
			ID:             d.ID,
			Imei:           d.ID, // ID field is the device IMEI.
			OrganizationID: d.OrganizationID,
			DeviceName:     d.DeviceName,
			Model:          d.Model,
			Manufacturer:   d.Manufacturer,
			AppVersion:     d.AppVersion,
			Status:         status,
			LastSeen:       d.LastSeen,
			RegisteredAt:   d.RegisteredAt,
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
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "operatorId is required"))
		return
	}

	devices, err := h.deviceService.ListByOperatorEntity(ctx, operatorID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to list devices by operator"))
		return
	}

	type DeviceRow struct {
		ID             string `json:"id"`
		Imei           string `json:"imei"`
		OrganizationID string `json:"organization_id"`
		DeviceName     string `json:"device_name"`
		Model          string `json:"model"`
		Manufacturer   string `json:"manufacturer"`
		AppVersion     string `json:"app_version"`
		Status         string `json:"status"`
		LastSeen       int64  `json:"last_seen"`
		RegisteredAt   int64  `json:"registered_at"`
	}

	result := make([]DeviceRow, len(devices))

	for i, d := range devices {
		isOnline := h.isDeviceOnline(d.ID) || d.Online

		// Determine status string based on online state and lifecycle.
		status := "offline"
		if d.DeregisteredAt != nil {
			status = "deregistered"
		} else if isOnline {
			status = "online"
		}

		result[i] = DeviceRow{
			ID:             d.ID,
			Imei:           d.ID,
			OrganizationID: d.OrganizationID,
			DeviceName:     d.DeviceName,
			Model:          d.Model,
			Manufacturer:   d.Manufacturer,
			AppVersion:     d.AppVersion,
			Status:         status,
			LastSeen:       d.LastSeen,
			RegisteredAt:   d.RegisteredAt,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"devices": result,
		"total":   len(result),
	})
}

// GetDevice handles GET /v1/device/:imei.
func (h *ListHandler) GetDevice(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
		return
	}

	d, err := h.deviceService.GetDevice(c.Request.Context(), imei)
	if err != nil {
		if err == devicedomain.ErrNotFound {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "device not found"))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to retrieve device"))

		return
	}

	isOnline := h.isDeviceOnline(d.ID) || d.Online

	// Determine status string based on online state and lifecycle.
	status := "offline"
	if d.DeregisteredAt != nil {
		status = "deregistered"
	} else if isOnline {
		status = "online"
	}

	c.JSON(http.StatusOK, gin.H{
		"id":            d.ID,
		"imei":          d.ID,
		"device_name":   d.DeviceName,
		"model":         d.Model,
		"manufacturer":  d.Manufacturer,
		"app_version":   d.AppVersion,
		"status":        status,
		"last_seen":     d.LastSeen,
		"registered_at": d.RegisteredAt,
	})
}

// Count handles GET /v1/device/count.
func (h *ListHandler) Count(c *gin.Context) {
	orgID := middleware.GetOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization context is required"))
		return
	}

	count, err := h.deviceService.CountByOrganization(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to count devices"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":      count,
		"serverTime": c.GetInt64("serverTime"),
	})
}

// CountByOrganization handles GET /v1/device/count?organizationId=<id>.
// Returns the count of devices for a specific organization.
func (h *ListHandler) CountByOrganization(c *gin.Context) {
	orgID := c.Query("organizationId")
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organizationId is required"))
		return
	}

	count, err := h.deviceService.CountByOrganization(c.Request.Context(), orgID)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to count devices"))
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"count":      count,
		"serverTime": c.GetInt64("serverTime"),
	})
}

func (h *ListHandler) GetTags(c *gin.Context) {
	imei := c.Param("imei")
	d, err := h.deviceService.GetDevice(c.Request.Context(), imei)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "device not found"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": d.Tags})
}

func (h *ListHandler) SetTags(c *gin.Context) {
	imei := c.Param("imei")
	var body struct {
		Tags []string `json:"tags"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid body"})
		return
	}
	if err := h.deviceService.SetDeviceTags(c.Request.Context(), imei, body.Tags); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to set tags"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"tags": body.Tags})
}

func (h *ListHandler) AddTag(c *gin.Context) {
	imei := c.Param("imei")
	tag := c.Param("tag")
	if err := h.deviceService.AddDeviceTag(c.Request.Context(), imei, tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to add tag"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"added": tag})
}

func (h *ListHandler) RemoveTag(c *gin.Context) {
	imei := c.Param("imei")
	tag := c.Param("tag")
	if err := h.deviceService.RemoveDeviceTag(c.Request.Context(), imei, tag); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to remove tag"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"removed": tag})
}
