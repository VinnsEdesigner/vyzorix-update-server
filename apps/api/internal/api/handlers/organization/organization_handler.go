package organization

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appOrganization "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"

	"github.com/gin-gonic/gin"
)

// OrganizationHandler handles organization-related HTTP requests.
type OrganizationHandler struct {
	orgService    *appOrganization.OrganizationService
	memberService *appOrganization.MemberService
	presenter    *response.Presenter
}

// NewOrganizationHandler creates a new OrganizationHandler.
func NewOrganizationHandler(
	orgService *appOrganization.OrganizationService,
	memberService *appOrganization.MemberService,
	presenter *response.Presenter,
) *OrganizationHandler {
	return &OrganizationHandler{
		orgService:    orgService,
		memberService: memberService,
		presenter:     presenter,
	}
}

// Create handles POST /v1/organizations.
func (h *OrganizationHandler) Create(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	var req struct {
		Name       string `json:"name" binding:"required"`
		MaxMembers int    `json:"maxMembers"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request body")
		return
	}

	if req.Name == "" {
		h.presenter.BadRequest(c, "organization name is required")
		return
	}

	org, err := h.orgService.CreateOrganization(c.Request.Context(), op.ID, req.Name, req.MaxMembers)
	if err != nil {
		if errors.Is(err, appOrganization.ErrMaxOrgsReached) {
			h.presenter.Forbidden(c, "maximum 2 active organizations allowed")
			return
		}
		if errors.Is(err, organization.ErrOrganizationExists) {
			h.presenter.Conflict(c, "organization with this name already exists")
			return
		}
		h.presenter.InternalError(c, "failed to create organization")
		return
	}

	h.presenter.Created(c, gin.H{
		"id":           org.ID,
		"name":         org.Name,
		"created_by":   org.CreatedBy,
		"created_at":   org.CreatedAt,
		"max_members":  org.MaxMembers,
		"is_active":    org.IsActive,
	})
}

// List handles GET /v1/organizations.
func (h *OrganizationHandler) List(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgs, err := h.orgService.ListOrganizations(c.Request.Context(), op.ID)
	if err != nil {
		h.presenter.InternalError(c, "failed to list organizations")
		return
	}

	result := make([]gin.H, len(orgs))
	for i, org := range orgs {
		result[i] = gin.H{
			"id":           org.ID,
			"name":         org.Name,
			"created_by":   org.CreatedBy,
			"created_at":   org.CreatedAt,
			"updated_at":   org.UpdatedAt,
			"max_members":  org.MaxMembers,
			"is_active":    org.IsActive,
			"member_count": org.MemberCount,
		}
	}

	h.presenter.OK(c, gin.H{
		"organizations": result,
	})
}

// Get handles GET /v1/organizations/:id.
func (h *OrganizationHandler) Get(c *gin.Context) {
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

	// Check if operator is a member of this org
	_, err := h.memberService.GetMembership(c.Request.Context(), op.ID, orgID)
	if err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	org, err := h.orgService.GetOrganizationWithMembers(c.Request.Context(), orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			h.presenter.NotFound(c, "organization not found")
			return
		}
		h.presenter.InternalError(c, "failed to get organization")
		return
	}

	h.presenter.OK(c, gin.H{
		"id":           org.ID,
		"name":         org.Name,
		"created_by":   org.CreatedBy,
		"created_at":   org.CreatedAt,
		"updated_at":   org.UpdatedAt,
		"max_members":  org.MaxMembers,
		"is_active":    org.IsActive,
		"member_count": org.MemberCount,
	})
}

// Update handles PATCH /v1/organizations/:id.
func (h *OrganizationHandler) Update(c *gin.Context) {
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

	// Check if operator can manage organization settings
	if err := h.memberService.CheckCanManageOrganization(c.Request.Context(), op.ID, orgID); err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	var req struct {
		Name       *string `json:"name"`
		MaxMembers *int    `json:"maxMembers"`
		IsActive   *bool   `json:"isActive"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request body")
		return
	}

	org, err := h.orgService.UpdateOrganization(c.Request.Context(), orgID, req.Name, req.MaxMembers, req.IsActive)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			h.presenter.NotFound(c, "organization not found")
			return
		}
		h.presenter.InternalError(c, "failed to update organization")
		return
	}

	h.presenter.OK(c, gin.H{
		"id":           org.ID,
		"name":         org.Name,
		"created_by":   org.CreatedBy,
		"created_at":   org.CreatedAt,
		"updated_at":   org.UpdatedAt,
		"max_members":  org.MaxMembers,
		"is_active":    org.IsActive,
		"member_count": org.MemberCount,
	})
}

// Delete handles DELETE /v1/organizations/:id.
func (h *OrganizationHandler) Delete(c *gin.Context) {
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

	// Check if operator can manage organization (must be super_admin)
	membership, err := h.memberService.GetMembership(c.Request.Context(), op.ID, orgID)
	if err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	if !membership.Role.CanDeleteOrganization() {
		h.presenter.Forbidden(c, "only super_admin can delete organization")
		return
	}

	if err := h.orgService.DeleteOrganization(c.Request.Context(), orgID); err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			h.presenter.NotFound(c, "organization not found")
			return
		}
		h.presenter.InternalError(c, "failed to delete organization")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "organization deleted",
	})
}
