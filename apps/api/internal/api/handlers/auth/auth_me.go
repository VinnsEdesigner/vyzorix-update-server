package auth

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// MeHandler handles GET /v1/auth/me.
type MeHandler struct {
	authService *auth.AuthService
	presenter   *response.Presenter
}

// NewMeHandler creates a new MeHandler.
func NewMeHandler(authService *auth.AuthService, presenter *response.Presenter) *MeHandler {
	return &MeHandler{authService: authService, presenter: presenter}
}

// Handle processes the me request.
// Note: Authentication is handled by cookieAuth middleware which sets the operator in context.
func (h *MeHandler) Handle(c *gin.Context) {
	// Get operator from context (set by cookieAuth middleware).
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	// Get operator's organizations.
	orgs, err := h.authService.GetOperatorOrganizations(c.Request.Context(), op)
	if err != nil {
		h.presenter.InternalError(c, "failed to get organizations")
		return
	}

	// Determine if organization selection is required:.
	// - 0 memberships: needs organization (create/join).
	// - Multiple memberships without valid selected org: needs selection.
	needsOrg := len(orgs) == 0

	// Find selected organization from session or last used.
	var selectedOrg *dto.OrganizationInfo
	if op.LastOrganizationID != "" && !needsOrg {
		for _, org := range orgs {
			if org.ID == op.LastOrganizationID {
				selectedOrg = &dto.OrganizationInfo{
					ID:   org.ID,
					Name: org.Name,
					Role: string(org.Role),
				}
				break
			}
		}
		// If LastOrganizationID is set but not found in active orgs,.
		// user has multiple orgs but the last one is invalid - needs selection.
		if selectedOrg == nil && len(orgs) > 1 {
			needsOrg = true
		}
	}

	// If multiple orgs exist but none is selected, needs organization selection.
	if !needsOrg && len(orgs) > 1 && selectedOrg == nil {
		needsOrg = true
	}

	h.presenter.OK(c, gin.H{
		"id":                    op.ID,
		"email":                 op.Email,
		"name":                  op.Name,
		"mfa_enabled":           op.MFAEnabled,
		"email_verified":        op.EmailVerified,
		"needs_organization":    needsOrg,
		"organizations":         orgs,
		"last_organization_id":  op.LastOrganizationID,
		"selected_organization": selectedOrg,
	})
}
