package organization

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appOrganization "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"

	"github.com/gin-gonic/gin"
)

// MemberHandler handles organization member-related HTTP requests.
type MemberHandler struct {
	memberService *appOrganization.MemberService
	presenter    *response.Presenter
}

// NewMemberHandler creates a new MemberHandler.
func NewMemberHandler(
	memberService *appOrganization.MemberService,
	presenter *response.Presenter,
) *MemberHandler {
	return &MemberHandler{
		memberService: memberService,
		presenter:     presenter,
	}
}

// List handles GET /v1/organizations/:id/members.
func (h *MemberHandler) List(c *gin.Context) {
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

	members, err := h.memberService.ListMembers(c.Request.Context(), orgID)
	if err != nil {
		h.presenter.InternalError(c, "failed to list members")
		return
	}

	result := make([]gin.H, len(members))
	for i, member := range members {
		result[i] = gin.H{
			"id":              member.ID,
			"organization_id": member.OrganizationID,
			"operator_id":     member.OperatorID,
			"role":            member.Role,
			"invited_by":      member.InvitedBy,
			"joined_at":       member.JoinedAt,
			"status":          string(member.Lifecycle),
			"operator_name":   member.OperatorName,
			"operator_email":  member.OperatorEmail,
		}
	}

	h.presenter.OK(c, gin.H{
		"members": result,
	})
}

// Get handles GET /v1/organizations/:id/members/:memberId.
func (h *MemberHandler) Get(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	memberID := c.Param("memberId")

	if orgID == "" || memberID == "" {
		h.presenter.BadRequest(c, "organization id and member id are required")
		return
	}

	// Check if operator is a member of this org
	_, err := h.memberService.GetMembership(c.Request.Context(), op.ID, orgID)
	if err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	member, err := h.memberService.GetMember(c.Request.Context(), memberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			h.presenter.NotFound(c, "member not found")
			return
		}
		h.presenter.InternalError(c, "failed to get member")
		return
	}

	// Verify member belongs to the org
	if member.OrganizationID != orgID {
		h.presenter.NotFound(c, "member not found")
		return
	}

	h.presenter.OK(c, gin.H{
		"id":              member.ID,
		"organization_id": member.OrganizationID,
		"operator_id":     member.OperatorID,
		"role":            member.Role,
		"invited_by":      member.InvitedBy,
		"joined_at":       member.JoinedAt,
		"removed_at":      member.RemovedAt,
		"status":          string(member.Lifecycle),
		"operator_name":   member.OperatorName,
		"operator_email":  member.OperatorEmail,
	})
}

// UpdateRole handles PATCH /v1/organizations/:id/members/:memberId.
func (h *MemberHandler) UpdateRole(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	memberID := c.Param("memberId")

	if orgID == "" || memberID == "" {
		h.presenter.BadRequest(c, "organization id and member id are required")
		return
	}

	// Check if operator can manage members
	if err := h.memberService.CheckCanManageMembers(c.Request.Context(), op.ID, orgID); err != nil {
		h.presenter.Forbidden(c, "access denied")
		return
	}

	var req struct {
		Role string `json:"role" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request body")
		return
	}

	// Validate role
	var newRole organization.OrganizationRole
	switch req.Role {
	case "admin":
		newRole = organization.RoleAdmin
	case "operator":
		newRole = organization.RoleOperator
	case "viewer":
		newRole = organization.RoleViewer
	case "super_admin":
		h.presenter.BadRequest(c, "cannot change role to super_admin")
		return
	default:
		h.presenter.BadRequest(c, "invalid role")
		return
	}

	member, err := h.memberService.UpdateMemberRole(c.Request.Context(), orgID, memberID, op.ID, newRole)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			h.presenter.NotFound(c, "member not found")
			return
		}
		if errors.Is(err, appOrganization.ErrCannotModifyHigherRole) {
			h.presenter.Forbidden(c, "cannot modify member with equal or higher role")
			return
		}
		h.presenter.InternalError(c, "failed to update member role")
		return
	}

	h.presenter.OK(c, gin.H{
		"id":              member.ID,
		"organization_id": member.OrganizationID,
		"operator_id":     member.OperatorID,
		"role":            member.Role,
		"status":          string(member.Lifecycle),
		"operator_name":   member.OperatorName,
		"operator_email":  member.OperatorEmail,
	})
}

// Remove handles DELETE /v1/organizations/:id/members/:memberId.
func (h *MemberHandler) Remove(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	memberID := c.Param("memberId")

	if orgID == "" || memberID == "" {
		h.presenter.BadRequest(c, "organization id and member id are required")
		return
	}

	err := h.memberService.RemoveMember(c.Request.Context(), orgID, memberID, op.ID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			h.presenter.NotFound(c, "member not found")
			return
		}
		if errors.Is(err, appOrganization.ErrCannotRemoveLastSuperAdmin) {
			h.presenter.Forbidden(c, "cannot remove the last super_admin")
			return
		}
		if errors.Is(err, appOrganization.ErrCannotModifyHigherRole) {
			h.presenter.Forbidden(c, "cannot remove member with equal or higher role")
			return
		}
		if err.Error() == "cannot remove yourself from organization" {
			h.presenter.Forbidden(c, "cannot remove yourself from organization")
			return
		}
		h.presenter.InternalError(c, "failed to remove member")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "member removed",
	})
}

// TransferOwnership handles POST /v1/organizations/:id/members/:memberId/transfer.
func (h *MemberHandler) TransferOwnership(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	orgID := c.Param("id")
	memberID := c.Param("memberId")

	if orgID == "" || memberID == "" {
		h.presenter.BadRequest(c, "organization id and member id are required")
		return
	}

	err := h.memberService.TransferOwnership(c.Request.Context(), orgID, op.ID, memberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			h.presenter.NotFound(c, "member not found")
			return
		}
		if errors.Is(err, organization.ErrForbidden) {
			h.presenter.Forbidden(c, "only super_admin can transfer ownership")
			return
		}
		h.presenter.InternalError(c, "failed to transfer ownership")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "ownership transferred successfully",
	})
}
