package auth

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"

	"github.com/gin-gonic/gin"
)

// OrganizationHandler handles organization selection requests.
type OrganizationHandler struct {
	authService *auth.AuthService
	presenter   *response.Presenter
}

// NewOrganizationHandler creates a new OrganizationHandler.
func NewOrganizationHandler(authService *auth.AuthService, presenter *response.Presenter) *OrganizationHandler {
	return &OrganizationHandler{authService: authService, presenter: presenter}
}

// SelectOrganization handles POST /v1/auth/organizations/select.
// This endpoint allows operators with multiple organization memberships to switch between them.
func (h *OrganizationHandler) SelectOrganization(c *gin.Context) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	// Validate session and get operator
	sess, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "session invalid or expired")
			return
		}
		h.presenter.InternalError(c, "an error occurred")
		return
	}

	// Parse request
	var req dto.SelectOrganizationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "organization_id is required",
		})
		return
	}

	if req.OrganizationID == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":   "invalid_request",
			"message": "organization_id is required",
		})
		return
	}

	// Select organization
	result, err := h.authService.SelectOrganization(c.Request.Context(), op.ID, sess.ID, req.OrganizationID)
	if err != nil {
		if errors.Is(err, application.ErrForbidden) {
			h.presenter.Forbidden(c, "you are not a member of this organization")
			return
		}
		if errors.Is(err, application.ErrUnauthorized) {
			h.presenter.Unauthorized(c, "session invalid or expired")
			return
		}
		h.presenter.InternalError(c, "an error occurred selecting organization")
		return
	}

	h.presenter.OK(c, dto.SelectOrganizationResponse{
		OrganizationID:   result.OrganizationID,
		OrganizationName: result.OrganizationName,
		Role:            string(result.Role),
	})
}

// GetOrganizations handles GET /v1/auth/organizations.
// This endpoint returns all organizations the operator is a member of.
func (h *OrganizationHandler) GetOrganizations(c *gin.Context) {
	sessionID, err := c.Cookie("vyz_session")
	if err != nil {
		h.presenter.Unauthorized(c, "not authenticated")
		return
	}

	_, op, err := h.authService.ValidateSession(c.Request.Context(), sessionID)
	if err != nil {
		if errors.Is(err, application.ErrUnauthorized) || errors.Is(err, application.ErrTokenExpired) {
			h.presenter.Unauthorized(c, "session invalid or expired")
			return
		}
		h.presenter.InternalError(c, "an error occurred")
		return
	}

	orgs, err := h.authService.GetOperatorOrganizations(c.Request.Context(), op)
	if err != nil {
		h.presenter.InternalError(c, "an error occurred fetching organizations")
		return
	}

	h.presenter.OK(c, gin.H{
		"organizations": orgs,
	})
}
