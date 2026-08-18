package device

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
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

// getOrganizationID extracts the organization ID from the Gin context.
func (h *DevicesHandler) getOrganizationID(c *gin.Context) string {
	return middleware.GetOrganizationID(c)
}

// GetDevices handles GET /v1/devices.
func (h *DevicesHandler) GetDevices(c *gin.Context) {
	orgID := h.getOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization_id is required"))
		return
	}

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
		OrganizationID: orgID,
		Status:         status,
		Search:         search,
		Page:           page,
		Limit:          limit,
	}

	result, err := h.service.GetDevices(c.Request.Context(), query)
	if err != nil {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to get devices"))
		return
	}

	c.JSON(http.StatusOK, result)
}

// GetDeviceDetail handles GET /v1/devices/:imei.
// Requires organization context for multi-tenant isolation.
func (h *DevicesHandler) GetDeviceDetail(c *gin.Context) {
	orgID := h.getOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization_id is required"))
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
		return
	}

	d, err := h.service.GetDeviceDetailByOrganization(c.Request.Context(), imei, orgID)
	if err != nil {
		if errors.Is(err, devicedomain.ErrNotFound) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "device not found in organization"))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to get device"))
		return
	}

	c.JSON(http.StatusOK, d)
}

// DeregisterDevice handles DELETE /v1/devices/:imei.
// Requires organization context for multi-tenant isolation.
func (h *DevicesHandler) DeregisterDevice(c *gin.Context) {
	orgID := h.getOrganizationID(c)
	if orgID == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "organization_id is required"))
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "IMEI is required"))
		return
	}

	hard := c.Query("hard") == "true"

	result, err := h.service.DeregisterDeviceByOrganization(c.Request.Context(), imei, orgID, hard)
	if err != nil {
		if errors.Is(err, devicedomain.ErrNotFound) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeResourceNotFound, "device not found in organization"))
			return
		}
		_ = c.Error(apperrors.NewServerError(apperrors.CodeInternalServerError, "failed to deregister device"))
		return
	}

	c.JSON(http.StatusOK, result)
}
