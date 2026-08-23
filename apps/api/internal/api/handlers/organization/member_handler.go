package organization

import (
	"context"
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
	_ openapi.OrganizationMember
	_ openapi.OrganizationMemberListResult
	_ openapi.UpdateMemberRoleRequest
	_ openapi.MessageResult
	_ openapi.ErrorResponse
)

// MemberHandler handles organization member-related HTTP requests.
type MemberHandler struct {
	memberService *appOrganization.MemberService
	presenter     *response.Presenter
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

// MembershipChecker returns an adapter that implements the OrganizationMembershipChecker interface.
// This allows the handler's memberService to be used by the org membership middleware.
func (h *MemberHandler) MembershipChecker() *MemberServiceAdapter {
	return &MemberServiceAdapter{memberService: h.memberService}
}

// MemberServiceAdapter implements middleware.OrganizationMembershipChecker.
// using the organization's member service.
type MemberServiceAdapter struct {
	memberService *appOrganization.MemberService
}

// GetMembership checks if an operator is a member of an organization.
func (a *MemberServiceAdapter) GetMembership(ctx context.Context, operatorID, orgID string) (interface{}, error) {
	if a.memberService == nil {
		return nil, errors.New("member service not configured")
	}
	return a.memberService.GetMembership(ctx, operatorID, orgID)
}

// List handles GET /v1/organizations/:id/members.
// @Summary      List organization members
// @Description  Returns all members of an organization
// @Tags         members
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "organization ID"
// @Success      200  {object}  openapi.OrganizationMemberListResult  "members"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/members [get]
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

	// Use membership from context (set by OrganizationMembership middleware).
	// If middleware didn't run, fall back to service call.
	member := middleware.GetMembership(c)
	if member == nil {
		var err error
		_, err = h.memberService.GetMembership(c.Request.Context(), op.ID, orgID)
		if err != nil {
			h.presenter.Forbidden(c, "access denied")
			return
		}
	} else {
		// Check minimum role: require at least operator role to list members.
		if member.Role.Level() < organization.RoleOperator.Level() {
			h.presenter.Forbidden(c, "insufficient permissions: operator role required")
			return
		}
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
// @Summary      Get organization member
// @Description  Returns a single organization member by ID
// @Tags         members
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id        path  string  true  "organization ID"
// @Param        memberId  path  string  true  "member ID"
// @Success      200  {object}  openapi.OrganizationMember  "member"
// @Failure      400  {object}  openapi.ErrorResponse  "org id and member id required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "member not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/members/{memberId} [get]
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

	// Use membership from context (set by OrganizationMembership middleware).
	// If middleware didn't run, fall back to service call.
	ctxMember := middleware.GetMembership(c)
	if ctxMember == nil {
		var err error
		_, err = h.memberService.GetMembership(c.Request.Context(), op.ID, orgID)
		if err != nil {
			h.presenter.Forbidden(c, "access denied")
			return
		}
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

	// Verify member belongs to the org.
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
// @Summary      Update member role
// @Description  Updates a member's role. Cannot change role to super_admin
// @Tags         members
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id        path  string  true  "organization ID"
// @Param        memberId  path  string  true  "member ID"
// @Param        body  body  openapi.UpdateMemberRoleRequest  true  "new role"
// @Success      200  {object}  openapi.OrganizationMember  "updated member"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid role / input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied / insufficient role"
// @Failure      404  {object}  openapi.ErrorResponse  "member not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/members/{memberId}/role [patch]
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

	// Check if operator can manage members.
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

	// Validate role.
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
// @Summary      Remove organization member
// @Description  Removes a member from an organization. Cannot remove the last super_admin or yourself
// @Tags         members
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id        path  string  true  "organization ID"
// @Param        memberId  path  string  true  "member ID"
// @Success      200  {object}  openapi.MessageResult  "member removed"
// @Failure      400  {object}  openapi.ErrorResponse  "org id and member id required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "cannot remove last super_admin / self / higher role"
// @Failure      404  {object}  openapi.ErrorResponse  "member not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/members/{memberId} [delete]
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
// @Summary      Transfer ownership
// @Description  Transfers organization ownership to another member
// @Tags         members
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id         path  string  true  "organization ID"
// @Param        memberId   path  string  true  "member ID"
// @Success      200  {object}  openapi.MessageResult  "transfer ownership result"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "member not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/members/{memberId}/transfer [post]
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

// Suspend handles POST /v1/organizations/:id/members/:memberId/suspend.
// @Summary      Suspend member
// @Description  Suspends an organization member
// @Tags         members
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id         path  string  true  "organization ID"
// @Param        memberId   path  string  true  "member ID"
// @Success      200  {object}  openapi.MessageResult  "suspend member result"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "member not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/members/{memberId}/suspend [post]
func (h *MemberHandler) Suspend(c *gin.Context) {
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

	err := h.memberService.SuspendMember(c.Request.Context(), orgID, memberID, op.ID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			h.presenter.NotFound(c, "member not found")
			return
		}
		if errors.Is(err, appOrganization.ErrCannotModifyHigherRole) {
			h.presenter.Forbidden(c, "cannot suspend member with equal or higher role")
			return
		}
		if err.Error() == "cannot suspend yourself" {
			h.presenter.Forbidden(c, "cannot suspend yourself")
			return
		}
		if err.Error() == "member is already suspended" {
			h.presenter.BadRequest(c, "member is already suspended")
			return
		}
		h.presenter.InternalError(c, "failed to suspend member")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "member suspended",
	})
}

// Reinstate handles POST /v1/organizations/:id/members/:memberId/reinstate.
// @Summary      Reinstate member
// @Description  Reinstates a suspended organization member
// @Tags         members
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id         path  string  true  "organization ID"
// @Param        memberId   path  string  true  "member ID"
// @Success      200  {object}  openapi.MessageResult  "reinstate member result"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied"
// @Failure      404  {object}  openapi.ErrorResponse  "member not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/members/{memberId}/reinstate [post]
func (h *MemberHandler) Reinstate(c *gin.Context) {
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

	err := h.memberService.ReinstateMember(c.Request.Context(), orgID, memberID, op.ID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			h.presenter.NotFound(c, "member not found")
			return
		}
		if errors.Is(err, appOrganization.ErrCannotModifyHigherRole) {
			h.presenter.Forbidden(c, "cannot reinstate member with equal or higher role")
			return
		}
		if err.Error() == "cannot reinstate yourself" {
			h.presenter.Forbidden(c, "cannot reinstate yourself")
			return
		}
		if err.Error() == "member is not suspended" {
			h.presenter.BadRequest(c, "member is not suspended")
			return
		}
		h.presenter.InternalError(c, "failed to reinstate member")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "member reinstated",
	})
}
