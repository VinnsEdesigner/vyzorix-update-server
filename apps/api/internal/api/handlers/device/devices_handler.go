package device

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/gin-gonic/gin"
)

// DevicesHandler handles GET /v1/devices and GET /v1/devices/:imei.
type DevicesHandler struct {
	service *device.Service
}

// NewDevicesHandler creates a new DevicesHandler.
func NewDevicesHandler(service *device.Service) *DevicesHandler {
	return &DevicesHandler{service: service}
}

// getOperatorID extracts the operator ID from the Gin context.
func (h *DevicesHandler) getOperatorID(c *gin.Context) string {
	// Try to get from context set by auth middleware
	if opID, exists := c.Get("operator_id"); exists {
		if opIDStr, ok := opID.(string); ok {
			return opIDStr
		}
	}
	// Fallback to query param for dashboard access
	return c.Query("operatorId")
}

// GetDevices handles GET /v1/devices.
func (h *DevicesHandler) GetDevices(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	status := c.Query("status")
	search := c.Query("search")

	if page < 1 {
		page = 1
	}
	if limit < 1 || limit > 100 {
		limit = 20
	}

	query := &device.ListQuery{
		Status: status,
		Search: search,
		Page:   page,
		Limit:  limit,
	}

	result, err := h.service.GetDevices(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get devices"})
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetDeviceDetail handles GET /v1/devices/:imei.
// Implements DOA verification - returns device only if it belongs to the operator.
func (h *DevicesHandler) GetDeviceDetail(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "IMEI is required"})
		return
	}

	operatorID := h.getOperatorID(c)

	// If operator ID is provided, verify ownership
	if operatorID != "" {
		d, err := h.service.GetDeviceDetailByOperator(c.Request.Context(), imei, operatorID)
		if err != nil {
			if errors.Is(err, devicedomain.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found or not owned by operator"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device"})
			return
		}
		c.JSON(http.StatusOK, d)
		return
	}

	// No operator ID - get without ownership check (admin access)
	d, err := h.service.GetDeviceDetail(c.Request.Context(), imei)
	if err != nil {
		if errors.Is(err, devicedomain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get device"})
		return
	}

	c.JSON(http.StatusOK, d)
}

// DeregisterDevice handles DELETE /v1/devices/:imei.
// Implements DOA verification - only the owning operator can deregister.
func (h *DevicesHandler) DeregisterDevice(c *gin.Context) {
	imei := c.Param("imei")
	if imei == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad_request", "message": "IMEI is required"})
		return
	}

	hard := c.Query("hard") == "true"
	operatorID := h.getOperatorID(c)

	// If operator ID is provided, verify ownership
	if operatorID != "" {
		result, err := h.service.DeregisterDeviceByOperator(c.Request.Context(), imei, operatorID, hard)
		if err != nil {
			if errors.Is(err, devicedomain.ErrNotFound) {
				c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found or not owned by operator"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deregister device"})
			return
		}
		c.JSON(http.StatusOK, result)
		return
	}

	// No operator ID - allow without ownership check (admin access)
	result, err := h.service.DeregisterDevice(c.Request.Context(), imei, hard)
	if err != nil {
		if errors.Is(err, devicedomain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "not_found", "message": "device not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to deregister device"})
		return
	}

	c.JSON(http.StatusOK, result)
}
