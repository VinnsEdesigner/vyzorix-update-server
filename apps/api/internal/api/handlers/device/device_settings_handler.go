package device

import (
	"context"
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appdevice "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"

	"github.com/gin-gonic/gin"
)

// SettingsHandler handles device settings HTTP requests.
type SettingsHandler struct {
	settingsService *appdevice.DeviceSettingsService
	membershipChecker func(ctx context.Context, operatorID, orgID string) error
	presenter *response.Presenter
}

// NewDeviceSettingsHandler creates a new SettingsHandler.
func NewDeviceSettingsHandler(
	settingsService *appdevice.DeviceSettingsService,
	membershipChecker func(ctx context.Context, operatorID, orgID string) error,
	presenter *response.Presenter,
) *SettingsHandler {
	return &SettingsHandler{
		settingsService:  settingsService,
		membershipChecker: membershipChecker,
		presenter:       presenter,
	}
}

// getOrganizationID extracts the organization ID from the Gin context.
func (h *SettingsHandler) getOrganizationID(c *gin.Context) string {
	return middleware.GetOrganizationID(c)
}

// getOperator extracts the operator from the Gin context.
func (h *SettingsHandler) getOperator(c *gin.Context) string {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		return ""
	}
	return op.ID
}

// GetSettings handles GET /v1/devices/:imei/settings.
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	operatorID := h.getOperator(c)
	if operatorID == "" {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := h.getOrganizationID(c)
	if orgID == "" {
		h.presenter.BadRequest(c, "organization context required")
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		h.presenter.BadRequest(c, "device IMEI is required")
		return
	}

	// Check membership
	if h.membershipChecker != nil {
		if err := h.membershipChecker(c.Request.Context(), operatorID, orgID); err != nil {
			h.presenter.Forbidden(c, "access denied")
			return
		}
	}

	settings, err := h.settingsService.GetOrCreateSettings(c.Request.Context(), imei)
	if err != nil {
		if errors.Is(err, devicedomain.ErrNotFound) {
			h.presenter.NotFound(c, "device not found")
			return
		}
		h.presenter.InternalError(c, "failed to get settings")
		return
	}

	h.presenter.OK(c, settings)
}

// UpdateSettings handles PATCH /v1/devices/:imei/settings.
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	operatorID := h.getOperator(c)
	if operatorID == "" {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := h.getOrganizationID(c)
	if orgID == "" {
		h.presenter.BadRequest(c, "organization context required")
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		h.presenter.BadRequest(c, "device IMEI is required")
		return
	}

	// Check membership
	if h.membershipChecker != nil {
		if err := h.membershipChecker(c.Request.Context(), operatorID, orgID); err != nil {
			h.presenter.Forbidden(c, "access denied")
			return
		}
	}

	var req devicedomain.UpdateDeviceSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid request body")
		return
	}

	settings, err := h.settingsService.UpdateSettings(c.Request.Context(), imei, &req)
	if err != nil {
		if errors.Is(err, devicedomain.ErrSettingsNotFound) {
			h.presenter.NotFound(c, "settings not found")
			return
		}
		if errors.Is(err, devicedomain.ErrInvalidThreshold) {
			h.presenter.BadRequest(c, err.Error())
			return
		}
		h.presenter.InternalError(c, "failed to update settings")
		return
	}

	h.presenter.OK(c, settings)
}

// GetThresholds handles GET /v1/devices/:imei/settings/thresholds.
// Returns the effective thresholds using hierarchy: device → org → default
func (h *SettingsHandler) GetThresholds(c *gin.Context) {
	operatorID := h.getOperator(c)
	if operatorID == "" {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := h.getOrganizationID(c)
	if orgID == "" {
		h.presenter.BadRequest(c, "organization context required")
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		h.presenter.BadRequest(c, "device IMEI is required")
		return
	}

	// Check membership - anyone in org can view thresholds
	if h.membershipChecker != nil {
		if err := h.membershipChecker(c.Request.Context(), operatorID, orgID); err != nil {
			h.presenter.Forbidden(c, "access denied")
			return
		}
	}

	thresholds, err := h.settingsService.GetEffectiveThresholds(c.Request.Context(), imei)
	if err != nil {
		if errors.Is(err, devicedomain.ErrNotFound) {
			h.presenter.NotFound(c, "device not found")
			return
		}
		h.presenter.InternalError(c, "failed to get thresholds")
		return
	}

	h.presenter.OK(c, gin.H{
		"thresholds": thresholds,
	})
}

// UpdateThresholds handles PATCH /v1/devices/:imei/settings/thresholds.
func (h *SettingsHandler) UpdateThresholds(c *gin.Context) {
	operatorID := h.getOperator(c)
	if operatorID == "" {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := h.getOrganizationID(c)
	if orgID == "" {
		h.presenter.BadRequest(c, "organization context required")
		return
	}

	imei := c.Param("imei")
	if imei == "" {
		h.presenter.BadRequest(c, "device IMEI is required")
		return
	}

	// Check membership
	if h.membershipChecker != nil {
		if err := h.membershipChecker(c.Request.Context(), operatorID, orgID); err != nil {
			h.presenter.Forbidden(c, "access denied")
			return
		}
	}

	var req devicedomain.UpdateThresholdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid request body")
		return
	}

	settings, err := h.settingsService.UpdateThresholds(c.Request.Context(), imei, &req)
	if err != nil {
		if errors.Is(err, devicedomain.ErrSettingsNotFound) {
			h.presenter.NotFound(c, "settings not found")
			return
		}
		if errors.Is(err, devicedomain.ErrInvalidThreshold) {
			h.presenter.BadRequest(c, err.Error())
			return
		}
		h.presenter.InternalError(c, "failed to update thresholds")
		return
	}

	h.presenter.OK(c, gin.H{
		"thresholds": settings.Thresholds,
	})
}
