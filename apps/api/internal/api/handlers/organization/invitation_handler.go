package organization

import (
	"errors"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	appOrganization "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"

	"github.com/gin-gonic/gin"
)
type InvitationHandler struct {
	invitationService *appOrganization.InvitationService
	memberService    *appOrganization.MemberService
	presenter        *response.Presenter
}

// NewInvitationHandler creates a new InvitationHandler.
func NewInvitationHandler(
	invitationService *appOrganization.InvitationService,
	memberService *appOrganization.MemberService,
	presenter *response.Presenter,
) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		memberService:    memberService,
		presenter:        presenter,
	}
}

// Create handles POST /v1/invitations.
func (h *InvitationHandler) Create(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	var req struct {
		OrganizationID string `json:"organizationId" binding:"required"`
		Email         string `json:"email" binding:"required"`
		Role          string `json:"role" binding:"required"`
		Notes         string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request body")
		return
	}

	// Validate role
	var invRole invitation.InvitationRole
	switch req.Role {
	case "admin":
		invRole = invitation.InvitationRoleAdmin
	case "operator":
		invRole = invitation.InvitationRoleOperator
	case "viewer":
		invRole = invitation.InvitationRoleViewer
	default:
		h.presenter.BadRequest(c, "invalid role")
		return
	}

	inv, err := h.invitationService.CreateInvitation(
		c.Request.Context(),
		req.OrganizationID,
		op.ID,
		req.Email,
		invRole,
		req.Notes,
	)
	if err != nil {
		if errors.Is(err, organization.ErrForbidden) {
			h.presenter.Forbidden(c, "access denied")
			return
		}
		if errors.Is(err, organization.ErrNotFound) {
			h.presenter.NotFound(c, "organization not found")
			return
		}
		if errors.Is(err, invitation.ErrAlreadyExists) {
			h.presenter.Conflict(c, "invitation already exists for this email")
			return
		}
		if errors.Is(err, appOrganization.ErrMaxInvitationsReached) {
			h.presenter.Forbidden(c, "maximum pending invitations reached")
			return
		}
		if errors.Is(err, appOrganization.ErrCannotInviteSelf) {
			h.presenter.BadRequest(c, "cannot invite yourself")
			return
		}
		h.presenter.InternalError(c, "failed to create invitation")
		return
	}

	h.presenter.Created(c, gin.H{
		"id":              inv.ID,
		"organization_id": inv.OrganizationID,
		"email":           inv.Email,
		"role":            inv.Role,
		"status":          inv.Status,
		"token":           inv.Token,
		"invited_by":      inv.InvitedBy,
		"invited_at":      inv.InvitedAt,
		"expires_at":      inv.ExpiresAt,
	})
}

// ListByOrganization handles GET /v1/organizations/:id/invitations.
func (h *InvitationHandler) ListByOrganization(c *gin.Context) {
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

	// Optional status filter
	var status *invitation.InvitationStatus
	statusStr := c.Query("status")
	if statusStr != "" {
		var s invitation.InvitationStatus
		switch statusStr {
		case "pending":
			s = invitation.InvitationStatusPending
		case "approved":
			s = invitation.InvitationStatusApproved
		case "rejected":
			s = invitation.InvitationStatusRejected
		case "expired":
			s = invitation.InvitationStatusExpired
		default:
			h.presenter.BadRequest(c, "invalid status filter")
			return
		}
		status = &s
	}

	invitations, err := h.invitationService.ListInvitationsByOrganization(c.Request.Context(), orgID, status)
	if err != nil {
		h.presenter.InternalError(c, "failed to list invitations")
		return
	}

	result := make([]gin.H, len(invitations))
	for i, inv := range invitations {
		result[i] = gin.H{
			"id":               inv.ID,
			"organization_id":  inv.OrganizationID,
			"email":            inv.Email,
			"role":             inv.Role,
			"status":           inv.Status,
			"invited_by":       inv.InvitedBy,
			"invited_at":       inv.InvitedAt,
			"responded_at":     inv.RespondedAt,
			"responder_id":     inv.ResponderID,
			"expires_at":       inv.ExpiresAt,
			"organization_name": inv.OrganizationName,
			"inviter_name":     inv.InviterName,
		}
	}

	h.presenter.OK(c, gin.H{
		"invitations": result,
	})
}

// ListByInviter handles GET /v1/invitations (as inviter).
func (h *InvitationHandler) ListByInviter(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	invitations, err := h.invitationService.ListInvitationsByInviter(c.Request.Context(), op.ID)
	if err != nil {
		h.presenter.InternalError(c, "failed to list invitations")
		return
	}

	result := make([]gin.H, len(invitations))
	for i, inv := range invitations {
		result[i] = gin.H{
			"id":               inv.ID,
			"organization_id":  inv.OrganizationID,
			"email":            inv.Email,
			"role":             inv.Role,
			"status":           inv.Status,
			"invited_at":       inv.InvitedAt,
			"responded_at":     inv.RespondedAt,
			"expires_at":       inv.ExpiresAt,
			"organization_name": inv.OrganizationName,
		}
	}

	h.presenter.OK(c, gin.H{
		"invitations": result,
	})
}

// GetByToken handles GET /v1/invite/:token (public endpoint).
func (h *InvitationHandler) GetByToken(c *gin.Context) {
	token := c.Param("token")
	if token == "" {
		h.presenter.BadRequest(c, "token is required")
		return
	}

	inv, err := h.invitationService.GetInvitationByToken(c.Request.Context(), token)
	if err != nil {
		if errors.Is(err, appOrganization.ErrInvitationNotFound) {
			h.presenter.NotFound(c, "invitation not found")
			return
		}
		if errors.Is(err, invitation.ErrExpired) {
			c.JSON(http.StatusGone, gin.H{"error": "gone", "message": "invitation has expired"})
			return
		}
		h.presenter.InternalError(c, "failed to get invitation")
		return
	}

	h.presenter.OK(c, gin.H{
		"id":               inv.ID,
		"organization_id":  inv.OrganizationID,
		"organization_name": inv.OrganizationName,
		"email":            inv.Email,
		"role":             inv.Role,
		"status":           inv.Status,
		"invited_at":       inv.InvitedAt,
		"inviter_name":     inv.InviterName,
		"expires_at":       inv.ExpiresAt,
	})
}

// Accept handles POST /v1/invite/:token/accept.
func (h *InvitationHandler) Accept(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	token := c.Param("token")
	if token == "" {
		h.presenter.BadRequest(c, "token is required")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		// Notes is optional, so we can ignore bind errors
		req.Notes = ""
	}

	err := h.invitationService.AcceptInvitation(
		c.Request.Context(),
		token,
		op.ID,
		op.Email,
		req.Notes,
	)
	if err != nil {
		if errors.Is(err, appOrganization.ErrInvitationNotFound) {
			h.presenter.NotFound(c, "invitation not found")
			return
		}
		if errors.Is(err, invitation.ErrExpired) {
			c.JSON(http.StatusGone, gin.H{"error": "gone", "message": "invitation has expired"})
			return
		}
		if errors.Is(err, invitation.ErrAlreadyResponded) {
			h.presenter.Conflict(c, "invitation has already been processed")
			return
		}
		if errors.Is(err, invitation.ErrEmailMismatch) {
			h.presenter.Forbidden(c, "email does not match invitation")
			return
		}
		if errors.Is(err, appOrganization.ErrAlreadyOrgMember) {
			h.presenter.Conflict(c, "you are already a member of this organization")
			return
		}
		if errors.Is(err, appOrganization.ErrOrgAtCapacity) {
			h.presenter.Forbidden(c, "organization has reached its member limit")
			return
		}
		h.presenter.InternalError(c, "failed to accept invitation")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "invitation accepted",
	})
}

// Reject handles POST /v1/invite/:token/reject.
func (h *InvitationHandler) Reject(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	token := c.Param("token")
	if token == "" {
		h.presenter.BadRequest(c, "token is required")
		return
	}

	var req struct {
		Notes string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		req.Notes = ""
	}

	err := h.invitationService.RejectInvitation(
		c.Request.Context(),
		token,
		op.ID,
		op.Email,
		req.Notes,
	)
	if err != nil {
		if errors.Is(err, appOrganization.ErrInvitationNotFound) {
			h.presenter.NotFound(c, "invitation not found")
			return
		}
		if errors.Is(err, invitation.ErrAlreadyResponded) {
			h.presenter.Conflict(c, "invitation has already been processed")
			return
		}
		if errors.Is(err, invitation.ErrEmailMismatch) {
			h.presenter.Forbidden(c, "email does not match invitation")
			return
		}
		h.presenter.InternalError(c, "failed to reject invitation")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "invitation rejected",
	})
}

// ListPendingForEmail handles GET /v1/me/invitations.
func (h *InvitationHandler) ListPendingForEmail(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	invitations, err := h.invitationService.ListPendingInvitationsForEmail(c.Request.Context(), op.Email)
	if err != nil {
		h.presenter.InternalError(c, "failed to list invitations")
		return
	}

	result := make([]gin.H, len(invitations))
	for i, inv := range invitations {
		result[i] = gin.H{
			"id":                inv.ID,
			"organization_id":   inv.OrganizationID,
			"organization_name": inv.OrganizationName,
			"email":             inv.Email,
			"role":              inv.Role,
			"status":            inv.Status,
			"invited_at":        inv.InvitedAt,
			"inviter_name":      inv.InviterName,
			"expires_at":        inv.ExpiresAt,
		}
	}

	h.presenter.OK(c, gin.H{
		"invitations": result,
	})
}

// Delete handles DELETE /v1/invitations/:id (cancel/delete invitation).
func (h *InvitationHandler) Delete(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	invitationID := c.Param("id")
	if invitationID == "" {
		h.presenter.BadRequest(c, "invitation id is required")
		return
	}

	// Get invitation to verify permissions
	inv, err := h.invitationService.GetInvitationByID(c.Request.Context(), invitationID)
	if err != nil {
		if errors.Is(err, appOrganization.ErrInvitationNotFound) {
			h.presenter.NotFound(c, "invitation not found")
			return
		}
		h.presenter.InternalError(c, "failed to get invitation")
		return
	}

	// Only inviter or org admin can delete
	isInviter := inv.InvitedBy == op.ID
	isOrgAdmin := false
	if isInviter {
		isOrgAdmin = true // Inviter can always delete
	} else {
		// Check if user is org admin
		member, err := h.memberService.GetMembership(c.Request.Context(), op.ID, inv.OrganizationID)
		if err == nil && member.Role.CanManageMembers() {
			isOrgAdmin = true
		}
	}

	if !isInviter && !isOrgAdmin {
		h.presenter.Forbidden(c, "only inviter or organization admin can delete this invitation")
		return
	}

	// Only pending invitations can be deleted
	if inv.Status != invitation.InvitationStatusPending {
		h.presenter.Conflict(c, "only pending invitations can be deleted")
		return
	}

	if err := h.invitationService.CancelInvitation(c.Request.Context(), invitationID); err != nil {
		h.presenter.InternalError(c, "failed to delete invitation")
		return
	}

	h.presenter.OK(c, gin.H{
		"message": "invitation deleted",
	})
}
