package organization

import (
	"errors"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/adapters/response"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/openapi"
	appOrganization "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	apperrors "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/errors"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"

	"github.com/gin-gonic/gin"
)

// Compile-time references for swaggo-annotated openapi DTO types.
var (
	_ openapi.CreateInvitationRequest
	_ openapi.Invitation
	_ openapi.InvitationListResult
	_ openapi.InvitationByTokenResult
	_ openapi.MessageResult
	_ openapi.ErrorResponse
)

type InvitationHandler struct {
	invitationService *appOrganization.InvitationService
	memberService     *appOrganization.MemberService
	presenter         *response.Presenter
}

// NewInvitationHandler creates a new InvitationHandler.
func NewInvitationHandler(
	invitationService *appOrganization.InvitationService,
	memberService *appOrganization.MemberService,
	presenter *response.Presenter,
) *InvitationHandler {
	return &InvitationHandler{
		invitationService: invitationService,
		memberService:     memberService,
		presenter:         presenter,
	}
}

// Create handles POST /v1/invitations.
// @Summary      Create invitation
// @Description  Invites an operator to an organization by email
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        body  body  openapi.CreateInvitationRequest  true  "invitation details"
// @Success      201  {object}  openapi.Invitation  "created invitation"
// @Failure      400  {object}  openapi.ErrorResponse  "invalid input / cannot invite self"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "access denied / max invitations"
// @Failure      404  {object}  openapi.ErrorResponse  "organization not found"
// @Failure      409  {object}  openapi.ErrorResponse  "invitation already exists"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /invitations [post]
func (h *InvitationHandler) Create(c *gin.Context) {
	op := middleware.GetOperatorFromContext(c)
	if op == nil {
		h.presenter.Unauthorized(c, "authentication required")
		return
	}

	var req struct {
		OrganizationID string `json:"organizationId" binding:"required"`
		Email          string `json:"email" binding:"required"`
		Role           string `json:"role" binding:"required"`
		Notes          string `json:"notes"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		h.presenter.BadRequest(c, "Invalid request body")
		return
	}

	// Validate role.
	var invRole organization.OrganizationRole
	switch req.Role {
	case "admin":
		invRole = organization.RoleAdmin
	case "operator":
		invRole = organization.RoleOperator
	case "viewer":
		invRole = organization.RoleViewer
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
		if errors.Is(err, organization.ErrAlreadyExists) {
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
// @Summary      List organization invitations
// @Description  Returns invitations for an organization, optionally filtered by status
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id      path   string  true  "organization ID"
// @Param        status  query  string  false  "filter by status"
// @Success      200  {object}  openapi.InvitationListResult  "invitations"
// @Failure      400  {object}  openapi.ErrorResponse  "org id required / invalid status"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /organizations/{id}/invitations [get]
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

	// Optional status filter.
	var status *organization.InvitationStatus
	statusStr := c.Query("status")
	if statusStr != "" {
		var s organization.InvitationStatus
		switch statusStr {
		case "pending":
			s = organization.InvitationStatusPending
		case "approved":
			s = organization.InvitationStatusApproved
		case "rejected":
			s = organization.InvitationStatusRejected
		case "expired":
			s = organization.InvitationStatusExpired
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
			"id":                inv.ID,
			"organization_id":   inv.OrganizationID,
			"email":             inv.Email,
			"role":              inv.Role,
			"status":            inv.Status,
			"invited_by":        inv.InvitedBy,
			"invited_at":        inv.InvitedAt,
			"responded_at":      inv.RespondedAt,
			"responder_id":      inv.RespondedBy,
			"expires_at":        inv.ExpiresAt,
			"organization_name": inv.OrganizationName,
			"inviter_name":      inv.InviterName,
		}
	}

	h.presenter.OK(c, gin.H{
		"invitations": result,
	})
}

// ListByInviter handles GET /v1/invitations (as inviter).
// @Summary      List sent invitations
// @Description  Returns invitations sent by the authenticated operator
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.InvitationListResult  "invitations"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /invitations [get]
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
			"id":                inv.ID,
			"organization_id":   inv.OrganizationID,
			"email":             inv.Email,
			"role":              inv.Role,
			"status":            inv.Status,
			"invited_at":        inv.InvitedAt,
			"responded_at":      inv.RespondedAt,
			"expires_at":        inv.ExpiresAt,
			"organization_name": inv.OrganizationName,
		}
	}

	h.presenter.OK(c, gin.H{
		"invitations": result,
	})
}

// GetByToken handles GET /v1/invite/:token (public endpoint).
// @Summary      Get invitation by token
// @Description  Returns invitation details for a token (public, used by the invitee)
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        token  path  string  true  "invitation token"
// @Success      200  {object}  openapi.InvitationByTokenResult  "invitation"
// @Failure      400  {object}  openapi.ErrorResponse  "token required"
// @Failure      404  {object}  openapi.ErrorResponse  "invitation not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /invite/{token} [get]
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
		if errors.Is(err, organization.ErrInvitationExpired) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invitation has expired"))
			return
		}
		h.presenter.InternalError(c, "failed to get invitation")
		return
	}

	h.presenter.OK(c, gin.H{
		"id":                inv.ID,
		"organization_id":   inv.OrganizationID,
		"organization_name": inv.OrganizationName,
		"email":             inv.Email,
		"role":              inv.Role,
		"status":            inv.Status,
		"invited_at":        inv.InvitedAt,
		"inviter_name":      inv.InviterName,
		"expires_at":        inv.ExpiresAt,
	})
}

// Accept handles POST /v1/invite/:token/accept.
// @Summary      Accept invitation
// @Description  Accepts an invitation and adds the operator to the organization
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        token  path  string  true  "invitation token"
// @Success      200  {object}  openapi.MessageResult  "invitation accepted"
// @Failure      400  {object}  openapi.ErrorResponse  "token required"
// @Failure      403  {object}  openapi.ErrorResponse  "email mismatch / org at capacity"
// @Failure      404  {object}  openapi.ErrorResponse  "invitation not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /invite/{token}/accept [post]
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
		// Notes is optional, so we can ignore bind errors.
		req.Notes = ""
	}

	_, err := h.invitationService.AcceptInvitation(
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
		if errors.Is(err, organization.ErrInvitationExpired) {
			_ = c.Error(apperrors.NewServerError(apperrors.CodeValidationFailed, "invitation has expired"))
			return
		}
		if errors.Is(err, organization.ErrAlreadyResponded) {
			h.presenter.Conflict(c, "invitation has already been processed")
			return
		}
		if errors.Is(err, organization.ErrEmailMismatch) {
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
// @Summary      Reject invitation
// @Description  Rejects an invitation
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        token  path  string  true  "invitation token"
// @Success      200  {object}  openapi.MessageResult  "invitation rejected"
// @Failure      400  {object}  openapi.ErrorResponse  "token required"
// @Failure      403  {object}  openapi.ErrorResponse  "email does not match"
// @Failure      404  {object}  openapi.ErrorResponse  "invitation not found"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /invite/{token}/reject [post]
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
		if errors.Is(err, organization.ErrAlreadyResponded) {
			h.presenter.Conflict(c, "invitation has already been processed")
			return
		}
		if errors.Is(err, organization.ErrEmailMismatch) {
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
// @Summary      List pending invitations for current operator
// @Description  Returns pending invitations addressed to the authenticated operator's email
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Success      200  {object}  openapi.InvitationListResult  "pending invitations"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /me/invitations [get]
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
// @Summary      Delete invitation
// @Description  Cancels a pending invitation. Only the inviter or an org admin can delete
// @Tags         invitations
// @Accept       json
// @Produce      json
// @Param        X-Organization-ID  header  string  true  "Organization ID"
// @Param        id  path  string  true  "invitation ID"
// @Success      200  {object}  openapi.MessageResult  "invitation deleted"
// @Failure      400  {object}  openapi.ErrorResponse  "invitation id required"
// @Failure      401  {object}  openapi.ErrorResponse  "authentication required"
// @Failure      403  {object}  openapi.ErrorResponse  "only inviter or admin can delete"
// @Failure      404  {object}  openapi.ErrorResponse  "invitation not found"
// @Failure      409  {object}  openapi.ErrorResponse  "only pending invitations can be deleted"
// @Failure      500  {object}  openapi.ErrorResponse  "internal error"
// @Router       /invitations/{id} [delete]
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

	// Get invitation to verify permissions.
	inv, err := h.invitationService.GetInvitationByID(c.Request.Context(), invitationID)
	if err != nil {
		if errors.Is(err, appOrganization.ErrInvitationNotFound) {
			h.presenter.NotFound(c, "invitation not found")
			return
		}
		h.presenter.InternalError(c, "failed to get invitation")
		return
	}

	// Only inviter or org admin can delete.
	isInviter := inv.InvitedBy == op.ID
	isOrgAdmin := false
	if isInviter {
		isOrgAdmin = true // Inviter can always delete.
	} else {
		// Check if user is org admin.
		member, err := h.memberService.GetMembership(c.Request.Context(), op.ID, inv.OrganizationID)
		if err == nil && member.Role.CanManageMembers() {
			isOrgAdmin = true
		}
	}

	if !isInviter && !isOrgAdmin {
		h.presenter.Forbidden(c, "only inviter or organization admin can delete this invitation")
		return
	}

	// Only pending invitations can be deleted.
	if inv.Status != organization.InvitationStatusPending {
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
