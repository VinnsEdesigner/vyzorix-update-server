package auth

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/auth"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/dto"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.SelectOrganizationRequest
	_ openapi.SelectOrganizationResult
	_ openapi.OrganizationListResult
	_ openapi.LoginRequest
	_ openapi.LoginResult
	_ openapi.ErrorResponse
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
// @Summary      Select organization
// @Description  Switches the operator's active organization context
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        body  body  openapi.SelectOrganizationRequest  true  "organization selection"
// @Success      200  {object}  openapi.SelectOrganizationResult  "selection result"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Router       /auth/organizations/select [post]
func (h *OrganizationHandler) SelectOrganization(c *gin.Context) {
	// Get operator and session from context (set by cookieAuth middleware).
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	sessVal, exists := c.Get("session")
	if !exists {
		h.presenter.Unauthorized(c, "session invalid or expired")
		return
	}
	sess, ok := sessVal.(*session.Session)
	if !ok || sess == nil {
		h.presenter.Unauthorized(c, "session invalid or expired")
		return
	}

	// Parse request.
	var req dto.SelectOrganizationRequest
	if bindErr := c.ShouldBindJSON(&req); bindErr != nil {
		h.presenter.BadRequest(c, "organization_id is required")
		return
	}

	if req.OrganizationID == "" {
		h.presenter.BadRequest(c, "organization_id is required")
		return
	}

	// Select organization.
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
		Role:             string(result.Role),
	})
}

// GetOrganizations handles GET /v1/auth/organizations.
// This endpoint returns all organizations the operator is a member of.
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Success      200  {object}  openapi.OrganizationListResult  "operator's organizations"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /auth/organizations [get]
func (h *OrganizationHandler) GetOrganizations(c *gin.Context) {
	// Get operator from context (set by cookieAuth middleware).
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
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
