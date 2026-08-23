package organization

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	appOrganization "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.OrganizationSettingsResult
	_ openapi.UpdateOrganizationSettingsRequest
	_ openapi.ThresholdsResult
	_ openapi.ThresholdUpdateRequest
	_ openapi.ErrorResponse
)

// SettingsHandler handles organization settings HTTP requests.
type SettingsHandler struct {
	settingsService *appOrganization.OrganizationSettingsService
	memberService   *appOrganization.MemberService
	presenter       *response.Presenter
}

// NewSettingsHandler creates a new SettingsHandler.
func NewSettingsHandler(
	settingsService *appOrganization.OrganizationSettingsService,
	memberService *appOrganization.MemberService,
	presenter *response.Presenter,
) *SettingsHandler {
	return &SettingsHandler{
		settingsService: settingsService,
		memberService:   memberService,
		presenter:       presenter,
	}
}

// GetSettings handles GET /v1/organizations/:id/settings.
// @Summary      Get organization settings
// @Description  Returns organization-level settings
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "organization ID"
// @Success      200  {object}  openapi.OrganizationSettingsResult  "organization settings"
// @Failure      400  {object}  openapi.ErrorResponse  "org id required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/settings [get]
func (h *SettingsHandler) GetSettings(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	if orgID == "" {
		h.presenter.BadRequest(c, "organization id is required")
		return
	}

	// Check membership and permissions.
	if err := h.memberService.CheckCanManageOrganization(c.Request.Context(), op.ID, orgID); err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	settings, err := h.settingsService.GetOrCreateSettings(c.Request.Context(), orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			h.presenter.NotFound(c, "organization not found")
			return
		}
		h.presenter.InternalError(c, "failed to get settings")
		return
	}

	h.presenter.OK(c, settings)
}

// UpdateSettings handles PATCH /v1/organizations/:id/settings.
// @Summary      Update organization settings
// @Description  Updates organization-level settings
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id    path  string  true  "organization ID"
// @Param        body  body  openapi.UpdateOrganizationSettingsRequest  true  "settings updates"
// @Success      200  {object}  openapi.OrganizationSettingsResult  "updated organization settings"
// @Failure      400  {object}  openapi.ErrorResponse  "org id required / invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "settings not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/settings [patch]
func (h *SettingsHandler) UpdateSettings(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	if orgID == "" {
		h.presenter.BadRequest(c, "organization id is required")
		return
	}

	// Check membership and permissions.
	if err := h.memberService.CheckCanManageOrganization(c.Request.Context(), op.ID, orgID); err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	var req organization.UpdateOrganizationSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid request body")
		return
	}

	settings, err := h.settingsService.UpdateSettings(c.Request.Context(), orgID, &req)
	if err != nil {
		if errors.Is(err, organization.ErrSettingsNotFound) {
			h.presenter.NotFound(c, "settings not found")
			return
		}
		if errors.Is(err, organization.ErrInvalidThreshold) {
			h.presenter.BadRequest(c, err.Error())
			return
		}
		h.presenter.InternalError(c, "failed to update settings")
		return
	}

	h.presenter.OK(c, settings)
}

// GetThresholds handles GET /v1/organizations/:id/settings/thresholds.
// @Summary      Get organization thresholds
// @Description  Returns organization-level alert thresholds
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "organization ID"
// @Success      200  {object}  openapi.ThresholdsResult  "thresholds"
// @Failure      400  {object}  openapi.ErrorResponse  "org id required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/settings/thresholds [get]
func (h *SettingsHandler) GetThresholds(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	if orgID == "" {
		h.presenter.BadRequest(c, "organization id is required")
		return
	}

	// Check membership - anyone in org can view thresholds.
	if err := h.memberService.CheckMembership(c.Request.Context(), op.ID, orgID); err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	settings, err := h.settingsService.GetOrCreateSettings(c.Request.Context(), orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			h.presenter.NotFound(c, "organization not found")
			return
		}
		h.presenter.InternalError(c, "failed to get thresholds")
		return
	}

	h.presenter.OK(c, gin.H{
		"thresholds": settings.DefaultThresholds,
	})
}

// UpdateThresholds handles PATCH /v1/organizations/:id/settings/thresholds.
// @Summary      Update organization thresholds
// @Description  Updates organization-level alert thresholds
// @Tags         organizations
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id    path  string  true  "organization ID"
// @Param        body  body  openapi.ThresholdUpdateRequest  true  "threshold updates"
// @Success      200  {object}  openapi.ThresholdsResult  "updated thresholds"
// @Failure      400  {object}  openapi.ErrorResponse  "org id required / invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/settings/thresholds [patch]
func (h *SettingsHandler) UpdateThresholds(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	if orgID == "" {
		h.presenter.BadRequest(c, "organization id is required")
		return
	}

	// Check membership and permissions.
	if err := h.memberService.CheckCanManageOrganization(c.Request.Context(), op.ID, orgID); err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	var req organization.UpdateThresholdsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "invalid request body")
		return
	}

	settings, err := h.settingsService.UpdateThresholds(c.Request.Context(), orgID, &req)
	if err != nil {
		if errors.Is(err, organization.ErrSettingsNotFound) {
			h.presenter.NotFound(c, "settings not found")
			return
		}
		if errors.Is(err, organization.ErrInvalidThreshold) {
			h.presenter.BadRequest(c, err.Error())
			return
		}
		h.presenter.InternalError(c, "failed to update thresholds")
		return
	}

	h.presenter.OK(c, gin.H{
		"thresholds": settings.DefaultThresholds,
	})
}
