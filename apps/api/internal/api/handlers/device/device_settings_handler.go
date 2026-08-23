package device

import (
	"context"
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	appdevice "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/device"
	devicedomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.DeviceSettingsResult
	_ openapi.UpdateDeviceSettingsRequest
	_ openapi.ThresholdsResult
	_ openapi.ErrorResponse
)

// SettingsHandler handles device settings HTTP requests.
type SettingsHandler struct {
	settingsService   *appdevice.DeviceSettingsService
	membershipChecker func(ctx context.Context, operatorID, orgID string) error
	presenter         *response.Presenter
}

// NewDeviceSettingsHandler creates a new SettingsHandler.
func NewDeviceSettingsHandler(
	settingsService *appdevice.DeviceSettingsService,
	membershipChecker func(ctx context.Context, operatorID, orgID string) error,
	presenter *response.Presenter,
) *SettingsHandler {
	return &SettingsHandler{
		settingsService:   settingsService,
		membershipChecker: membershipChecker,
		presenter:         presenter,
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
// @Summary      Get device settings
// @Description  Returns device-level settings
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Success      200  {object}  openapi.DeviceSettingsResult  "device settings"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required / org context required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "device not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /devices/{imei}/settings [get]
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

	// Check membership.
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
// @Summary      Update device settings
// @Description  Updates device-level settings
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Param        body  body  openapi.UpdateDeviceSettingsRequest  true  "settings updates"
// @Success      200  {object}  openapi.DeviceSettingsResult  "updated device settings"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required / invalid body"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "settings not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /devices/{imei}/settings [patch]
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

	// Check membership.
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
// @Summary      Get device thresholds
// @Description  Returns the effective thresholds using hierarchy: device → org → default
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Success      200  {object}  openapi.ThresholdsResult  "thresholds"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required / org context required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /devices/{imei}/settings/thresholds [get]
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

	// Check membership - anyone in org can view thresholds.
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
// @Summary      Update device thresholds
// @Description  Updates device-level alert thresholds
// @Tags         devices
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        imei  path  string  true  "device IMEI"
// @Param        body  body  openapi.ThresholdUpdateRequest  true  "threshold updates"
// @Success      200  {object}  openapi.ThresholdsResult  "updated thresholds"
// @Failure      400  {object}  openapi.ErrorResponse  "IMEI required / invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /devices/{imei}/settings/thresholds [patch]
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

	// Check membership.
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
