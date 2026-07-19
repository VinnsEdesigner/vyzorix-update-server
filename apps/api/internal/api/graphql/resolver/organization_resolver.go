// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"fmt"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	orgapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/graphql-go/graphql"
	"time"
)

// ============================================================
// Organization Queries
// ============================================================

// GetOrganization resolves the organization query.
func (r *Resolver) GetOrganization(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	id, ok := p.Args["id"].(string)
	if !ok || id == "" {
		return nil, r.Presenter.BadRequestError("organization ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is a member of the organization
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, id); err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	org, err := r.OrgService.GetOrganization(ctx, id)
	if err != nil {
		return nil, r.Presenter.NotFoundError("organization not found")
	}

	return r.orgToMap(org), nil
}

// GetOrganizations resolves the organizations query.
func (r *Resolver) GetOrganizations(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Parse pagination args
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method
	result, err := r.OrgService.ListOrganizationsPaginated(ctx, op.ID, page, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list organizations")
	}

	// Convert to GraphQL response format
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, org := range result.Items {
		items = append(items, r.orgToMap(org))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.paginationToMap(result.Pagination),
	}, nil
}

// GetMyMemberships resolves the myMemberships query.
func (r *Resolver) GetMyMemberships(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Parse pagination args
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method
	result, err := r.MemberService.ListOperatorMembershipsPaginated(ctx, op.ID, page, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list memberships")
	}

	// Convert to GraphQL response format
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, m := range result.Items {
		items = append(items, r.membershipToMap(m))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.paginationToMap(result.Pagination),
	}, nil
}

// GetOrganizationMembers resolves the organizationMembers query.
func (r *Resolver) GetOrganizationMembers(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is a member
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	// Parse pagination args
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method (filters to active members)
	result, err := r.MemberService.ListMembersPaginated(ctx, orgID, page, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list members")
	}

	// Convert to GraphQL response format
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, m := range result.Items {
		items = append(items, r.membershipToMap(m))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.paginationToMap(result.Pagination),
	}, nil
}

// GetOrganizationInvitations resolves the organizationInvitations query.
func (r *Resolver) GetOrganizationInvitations(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, ok := p.Args["organizationId"].(string)
	if !ok || orgID == "" {
		return nil, r.Presenter.BadRequestError("organizationId is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is a member
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	// Parse pagination args
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method (filters to pending invitations when status is nil)
	result, err := r.InvitationService.ListInvitationsByOrganizationPaginated(ctx, orgID, page, limit, nil)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list invitations")
	}

	// Convert to GraphQL response format
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, inv := range result.Items {
		items = append(items, r.invitationToMap(inv))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.paginationToMap(result.Pagination),
	}, nil
}

// GetMyInvitations resolves the myInvitations query.
func (r *Resolver) GetMyInvitations(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Validate operator has an email set
	if op.Email == "" {
		// Return empty array if no email - user needs to set their email first
		return []map[string]interface{}{}, nil
	}

	invitations, err := r.InvitationService.ListPendingInvitationsForEmail(ctx, op.Email)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list invitations")
	}

	result := make([]map[string]interface{}, 0, len(invitations))
	for _, inv := range invitations {
		result = append(result, r.invitationToMap(inv))
	}

	return result, nil
}

// GetInvitationByToken resolves the invitation query.
func (r *Resolver) GetInvitationByToken(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	token, ok := p.Args["token"].(string)
	if !ok || token == "" {
		return nil, r.Presenter.BadRequestError("token is required")
	}

	invitation, err := r.InvitationService.GetInvitationByToken(ctx, token)
	if err != nil {
		return nil, r.Presenter.NotFoundError("invitation not found")
	}

	return r.invitationToMap(invitation), nil
}

// ============================================================
// Organization Mutations
// ============================================================

// CreateOrganization resolves the createOrganization mutation.
func (r *Resolver) CreateOrganization(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	name, ok := p.Args["name"].(string)
	if !ok || name == "" {
		return nil, r.Presenter.BadRequestError("name is required")
	}

	maxMembers, _ := p.Args["maxMembers"].(int)
	if maxMembers <= 0 {
		maxMembers = 100
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Note: CreateOrganization creates the org AND the creator's membership as super_admin
	// The membership is created inside the transaction, so we get it from there
	org, err := r.OrgService.CreateOrganization(ctx, op.ID, name, maxMembers)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to create organization")
	}

	// Get the creator's membership
	membership, err := r.MemberService.GetMembership(ctx, op.ID, org.ID.String())
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get membership")
	}

	// Publish organization created event
	if r.Hub != nil {
		r.Hub.PublishOrganizationEvent(org.ID.String(), map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "created",
			"organizationId": org.ID.String(),
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	return map[string]interface{}{
		"organization": r.orgToMap(org),
		"membership":   r.membershipToMap(membership),
	}, nil
}

// UpdateOrganization resolves the updateOrganization mutation.
func (r *Resolver) UpdateOrganization(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	id, _ := p.Args["id"].(string)
	name, _ := p.Args["name"].(string)
	maxMembers, _ := p.Args["maxMembers"].(int)
	isActive, _ := p.Args["isActive"].(bool)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator has admin rights
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, id); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to update organization")
	}

	// Prepare pointer arguments for optional fields
	var namePtr, maxMembersPtr, isActivePtr *string
	if name != "" {
		namePtr = &name
	}
	var maxMembersVal int
	if maxMembers > 0 {
		maxMembersVal = maxMembers
		maxMembersPtr = &maxMembersVal
	}
	var isActiveVal bool
	if p.Args["isActive"] != nil {
		isActiveVal = isActive
		isActivePtr = &isActiveVal
	}

	org, err := r.OrgService.UpdateOrganization(ctx, id, namePtr, maxMembersPtr, isActivePtr)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update organization")
	}

	// Publish organization updated event
	if r.Hub != nil {
		eventType := "updated"
		if isActivePtr != nil {
			if *isActivePtr {
				eventType = "activated"
			} else {
				eventType = "deactivated"
			}
		}
		r.Hub.PublishOrganizationEvent(id, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           eventType,
			"organizationId": id,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	return r.orgToMap(org), nil
}

// DeleteOrganization resolves the deleteOrganization mutation.
func (r *Resolver) DeleteOrganization(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	id, ok := p.Args["id"].(string)
	if !ok || id == "" {
		return nil, r.Presenter.BadRequestError("organization ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is a super_admin of this organization
	membership, err := r.MemberService.GetMembership(ctx, op.ID, id)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("must be a super_admin to delete organization")
	}

	if !membership.Role.CanDeleteOrganization() {
		return nil, r.Presenter.ForbiddenError("must be a super_admin to delete organization")
	}

	err = r.OrgService.DeleteOrganization(ctx, id)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to delete organization")
	}

	// Publish organization deleted event
	if r.Hub != nil {
		r.Hub.PublishOrganizationEvent(id, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "deleted",
			"organizationId": id,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	return true, nil
}

// InviteMember resolves the inviteMember mutation.
func (r *Resolver) InviteMember(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, _ := p.Args["organizationId"].(string)
	email, _ := p.Args["email"].(string)
	roleStr, _ := p.Args["role"].(string)
	notes, _ := p.Args["notes"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator can manage members
	if err := r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to invite members")
	}

	// Convert role string to OrganizationRole
	role := orgdomain.OrganizationRole(roleStr)

	invitation, err := r.InvitationService.CreateInvitation(ctx, orgID, op.ID, email, role, notes)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to create invitation")
	}

	// Publish member invited event
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "member_invited",
			"organizationId": orgID,
			"memberId":       invitation.ID.String(),
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
			"data": map[string]interface{}{
				"email": email,
				"role":  role,
			},
		})
	}

	return r.invitationToMap(invitation), nil
}

// RemoveMember resolves the removeMember mutation.
func (r *Resolver) RemoveMember(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, _ := p.Args["organizationId"].(string)
	memberID, _ := p.Args["memberId"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Get the actor's membership to check if trying to remove self
	actorMembership, err := r.MemberService.GetMembership(ctx, op.ID, orgID)
	if err != nil {
		// Not a member - can't remove anyone
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	// Prevent self-removal - compare actor's membership ID with target
	if actorMembership.ID.String() == memberID {
		return nil, r.Presenter.BadRequestError("cannot remove yourself from the organization")
	}

	// Prevent removing the last super_admin
	canRemove, err := r.canRemoveMember(ctx, orgID, memberID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to check permissions")
	}
	if !canRemove {
		return nil, r.Presenter.BadRequestError("cannot remove the last super_admin")
	}

	// Check if operator can manage members
	if err := r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to remove members")
	}

	err = r.MemberService.RemoveMember(ctx, orgID, memberID, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to remove member")
	}

	// Publish member removed event
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "member_removed",
			"organizationId": orgID,
			"memberId":       memberID,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	return true, nil
}

// canRemoveMember checks if removing the target member would leave the org without a super_admin.
func (r *Resolver) canRemoveMember(ctx context.Context, orgID, memberID string) (bool, error) {
	memberships, err := r.MemberService.ListMembers(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Find the target membership to check its role
	var targetRole orgdomain.OrganizationRole
	var targetLifecycle orgdomain.MemberLifecycle
	var memberFound bool
	for _, m := range memberships {
		if m.ID.String() == memberID {
			targetRole = m.Role
			targetLifecycle = m.Lifecycle
			memberFound = true
			break
		}
	}

	// If member doesn't exist, don't allow removal
	if !memberFound {
		return false, nil
	}

	// If target is not a super_admin or not active, removal is always safe
	if targetRole != orgdomain.RoleSuperAdmin || targetLifecycle != orgdomain.MemberLifecycleActive {
		return true, nil
	}

	// Count other active super_admins
	superAdminCount := 0
	for _, m := range memberships {
		if m.Role == orgdomain.RoleSuperAdmin && m.ID.String() != memberID && m.Lifecycle == orgdomain.MemberLifecycleActive {
			superAdminCount++
		}
	}

	return superAdminCount > 0, nil
}

// UpdateMemberRole resolves the updateMemberRole mutation.
func (r *Resolver) UpdateMemberRole(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, _ := p.Args["organizationId"].(string)
	memberID, _ := p.Args["memberId"].(string)
	roleStr, _ := p.Args["role"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Prevent self-role change
	memberships, err := r.MemberService.ListMembers(ctx, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list members")
	}

	// Find the operator's own membership in this org
	var ownMembership *orgdomain.Membership
	for _, m := range memberships {
		if m.OperatorID.String() == op.ID {
			ownMembership = m
			break
		}
	}

	if ownMembership != nil && ownMembership.ID.String() == memberID {
		return nil, r.Presenter.BadRequestError("cannot change your own role")
	}

	// Check if operator can manage members
	if err := r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to update member role")
	}

	// Prevent demoting the last super_admin
	if roleStr == string(orgdomain.RoleAdmin) || roleStr == string(orgdomain.RoleOperator) || roleStr == string(orgdomain.RoleViewer) {
		// Count super_admins before this change
		canChange, err := r.canDemoteSuperAdmin(ctx, orgID, memberID)
		if err != nil {
			return nil, r.Presenter.InternalError("failed to check permissions")
		}
		if !canChange {
			return nil, r.Presenter.BadRequestError("cannot demote the last super_admin")
		}
	}

	// Convert role string to OrganizationRole
	role := orgdomain.OrganizationRole(roleStr)

	// Get the old role before updating for the event
	var oldRole string
	for _, m := range memberships {
		if m.ID.String() == memberID {
			oldRole = string(m.Role)
			break
		}
	}

	membership, err := r.MemberService.UpdateMemberRole(ctx, orgID, memberID, op.ID, role)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update member role")
	}

	// Publish role changed event
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "role_changed",
			"organizationId": orgID,
			"memberId":       memberID,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
			"data": map[string]interface{}{
				"oldRole": oldRole,
				"newRole": role,
			},
		})
	}

	return r.membershipToMap(membership), nil
}

// canDemoteSuperAdmin checks if demoting the target member would leave the org without a super_admin.
func (r *Resolver) canDemoteSuperAdmin(ctx context.Context, orgID, memberID string) (bool, error) {
	memberships, err := r.MemberService.ListMembers(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Count super_admins excluding the member being demoted
	superAdminCount := 0
	for _, m := range memberships {
		if m.Role == orgdomain.RoleSuperAdmin && m.ID.String() != memberID && m.Lifecycle == orgdomain.MemberLifecycleActive {
			superAdminCount++
		}
	}

	return superAdminCount > 0, nil
}

// AcceptInvitation resolves the acceptInvitation mutation.
func (r *Resolver) AcceptInvitation(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	token, _ := p.Args["token"].(string)
	notes, _ := p.Args["notes"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Validate operator has an email set
	if op.Email == "" {
		return nil, r.Presenter.BadRequestError("operator email not set, please update your profile first")
	}

	// AcceptInvitation now returns the membership directly within the transaction
	membership, err := r.InvitationService.AcceptInvitation(ctx, token, op.ID, op.Email, notes)
	if err != nil {
		// Map service errors to proper GraphQL errors
		switch err {
		case orgdomain.ErrInvitationNotFound, orgdomain.ErrAlreadyResponded:
			return nil, r.Presenter.NotFoundError(err.Error())
		case orgdomain.ErrExpired:
			return nil, r.Presenter.BadRequestError("invitation has expired")
		case orgdomain.ErrEmailMismatch:
			return nil, r.Presenter.ForbiddenError("this invitation was sent to a different email address")
		default:
			return nil, r.Presenter.InternalError("failed to accept invitation: " + err.Error())
		}
	}

	// Publish member joined event
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(membership.OrganizationID.String(), map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "member_joined",
			"organizationId": membership.OrganizationID.String(),
			"memberId":       membership.ID.String(),
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	// Membership is returned directly from the transaction - no race condition
	return r.membershipToMap(membership), nil
}

// RejectInvitation resolves the rejectInvitation mutation.
func (r *Resolver) RejectInvitation(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	token, _ := p.Args["token"].(string)
	notes, _ := p.Args["notes"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Validate operator has an email set
	if op.Email == "" {
		return nil, r.Presenter.BadRequestError("operator email not set, please update your profile first")
	}

	err := r.InvitationService.RejectInvitation(ctx, token, op.ID, op.Email, notes)
	if err != nil {
		// Map service errors to proper GraphQL errors
		switch err {
		case orgdomain.ErrInvitationNotFound, orgdomain.ErrAlreadyResponded:
			return nil, r.Presenter.NotFoundError(err.Error())
		case orgdomain.ErrExpired:
			return nil, r.Presenter.BadRequestError("invitation has expired")
		case orgdomain.ErrEmailMismatch:
			return nil, r.Presenter.ForbiddenError("this invitation was sent to a different email address")
		default:
			return nil, r.Presenter.InternalError("failed to reject invitation: " + err.Error())
		}
	}

	return true, nil
}

// CancelInvitation resolves the cancelInvitation mutation.
func (r *Resolver) CancelInvitation(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	id, ok := p.Args["id"].(string)
	if !ok || id == "" {
		return nil, r.Presenter.BadRequestError("invitation ID is required")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Get the invitation to check org membership
	inv, err := r.InvitationService.GetInvitationByID(ctx, id)
	if err != nil {
		return nil, r.Presenter.NotFoundError("invitation not found")
	}

	// Check if operator can manage members in this org
	if err := r.MemberService.CheckCanManageMembers(ctx, op.ID, inv.OrganizationID.String()); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to cancel invitation")
	}

	err = r.InvitationService.CancelInvitation(ctx, id)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to cancel invitation")
	}

	return true, nil
}

// TransferDevice resolves the transferDevice mutation.
func (r *Resolver) TransferDevice(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	imei, _ := p.Args["imei"].(string)
	sourceOrgID, _ := p.Args["sourceOrganizationId"].(string)
	targetOrgID, _ := p.Args["targetOrganizationId"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check membership in source org
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, sourceOrgID); err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of source organization")
	}

	// Check membership in target org
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, targetOrgID); err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of target organization")
	}

	err := r.DeviceService.TransferDevice(ctx, imei, sourceOrgID, targetOrgID, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to transfer device")
	}

	return map[string]interface{}{
		"success":              true,
		"deviceId":             imei,
		"sourceOrganizationId": sourceOrgID,
		"targetOrganizationId": targetOrgID,
	}, nil
}

// TransferOwnership resolves the transferOwnership mutation.
func (r *Resolver) TransferOwnership(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, _ := p.Args["organizationId"].(string)
	memberID, _ := p.Args["memberId"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is super_admin of this org
	membership, err := r.MemberService.GetMembership(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	if membership.Role != orgdomain.RoleSuperAdmin {
		return nil, r.Presenter.ForbiddenError("only super_admin can transfer ownership")
	}

	// Transfer ownership
	err = r.MemberService.TransferOwnership(ctx, orgID, op.ID, memberID)
	if err != nil {
		if err == orgdomain.ErrMemberNotFound {
			return nil, r.Presenter.NotFoundError("member not found")
		}
		return nil, r.Presenter.InternalError("failed to transfer ownership")
	}

	// Publish ownership transferred event
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "ownership_transferred",
			"organizationId": orgID,
			"memberId":       memberID,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	return true, nil
}

// SuspendMember resolves the suspendMember mutation.
func (r *Resolver) SuspendMember(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, _ := p.Args["organizationId"].(string)
	memberID, _ := p.Args["memberId"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator can manage members
	if err := r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to suspend member")
	}

	err := r.MemberService.SuspendMember(ctx, orgID, memberID, op.ID)
	if err != nil {
		if err == orgdomain.ErrMemberNotFound {
			return nil, r.Presenter.NotFoundError("member not found")
		}
		if err.Error() == "cannot suspend yourself" {
			return nil, r.Presenter.BadRequestError("cannot suspend yourself")
		}
		if err.Error() == "member is already suspended" {
			return nil, r.Presenter.BadRequestError("member is already suspended")
		}
		if err == orgapp.ErrCannotModifyHigherRole {
			return nil, r.Presenter.ForbiddenError("cannot suspend member with equal or higher role")
		}
		return nil, r.Presenter.InternalError("failed to suspend member")
	}

	// Publish member suspended event
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "member_suspended",
			"organizationId": orgID,
			"memberId":       memberID,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	return true, nil
}

// ReinstateMember resolves the reinstateMember mutation.
func (r *Resolver) ReinstateMember(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, _ := p.Args["organizationId"].(string)
	memberID, _ := p.Args["memberId"].(string)

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator can manage members
	if err := r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to reinstate member")
	}

	err := r.MemberService.ReinstateMember(ctx, orgID, memberID, op.ID)
	if err != nil {
		if err == orgdomain.ErrMemberNotFound {
			return nil, r.Presenter.NotFoundError("member not found")
		}
		if err.Error() == "cannot reinstate yourself" {
			return nil, r.Presenter.BadRequestError("cannot reinstate yourself")
		}
		if err.Error() == "member is not suspended" {
			return nil, r.Presenter.BadRequestError("member is not suspended")
		}
		if err == orgapp.ErrCannotModifyHigherRole {
			return nil, r.Presenter.ForbiddenError("cannot reinstate member with equal or higher role")
		}
		return nil, r.Presenter.InternalError("failed to reinstate member")
	}

	// Publish member reinstated event
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "member_reinstated",
			"organizationId": orgID,
			"memberId":       memberID,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	return true, nil
}

// ============================================================
// Helper Functions
// ============================================================

func (r *Resolver) orgToMap(org *orgdomain.Organization) map[string]interface{} {
	result := map[string]interface{}{
		"id":          org.ID.String(),
		"name":        org.Name,
		"lifecycle":   org.Lifecycle,
		"maxMembers":  org.MaxMembers,
		"memberCount": org.MemberCount,
		"createdBy":   org.CreatedBy.String(),
		"createdAt":   formatTime(org.CreatedAt),
		"updatedAt":   formatTime(org.UpdatedAt),
	}

	if !org.DeletedAt.IsZero() {
		result["deletedAt"] = formatTime(org.DeletedAt)
	}

	return result
}

func (r *Resolver) membershipToMap(m *orgdomain.Membership) map[string]interface{} {
	result := map[string]interface{}{
		"id":             m.ID.String(),
		"organizationId": m.OrganizationID.String(),
		"operatorId":     m.OperatorID.String(),
		"role":           m.Role,
		"lifecycle":      m.Lifecycle,
		"invitedAt":      formatTime(m.InvitedAt),
	}

	if !m.JoinedAt.IsZero() {
		result["joinedAt"] = formatTime(m.JoinedAt)
	}

	if !m.RemovedAt.IsZero() {
		result["removedAt"] = formatTime(m.RemovedAt)
	}

	if !m.SuspendedAt.IsZero() {
		result["suspendedAt"] = formatTime(m.SuspendedAt)
	}

	return result
}

func (r *Resolver) invitationToMap(inv *orgdomain.Invitation) map[string]interface{} {
	result := map[string]interface{}{
		"id":               inv.ID.String(),
		"organizationId":    inv.OrganizationID.String(),
		"organizationName": inv.OrganizationName,
		"email":            inv.Email,
		"role":             inv.Role,
		"status":           inv.Status,
		"token":            inv.Token,
		"inviterId":        inv.InviterID.String(),
		"inviterName":      inv.InviterName,
		"createdAt":        formatTime(inv.CreatedAt),
		"expiresAt":        formatTime(inv.ExpiresAt),
	}

	if inv.InviteeID.Valid {
		result["inviteeId"] = inv.InviteeID.String
	}

	if inv.InviteeNotes.Valid {
		result["inviteeNotes"] = inv.InviteeNotes.String
	}

	if inv.InviterNotes.Valid {
		result["inviterNotes"] = inv.InviterNotes.String
	}

	if !inv.RespondedAt.IsZero() {
		result["respondedAt"] = formatTime(inv.RespondedAt)
	}

	return result
}

func (r *Resolver) paginationToMap(p orgapp.Pagination) map[string]interface{} {
	return map[string]interface{}{
		"page":       p.Page,
		"limit":      p.Limit,
		"total":      p.Total,
		"totalPages": p.TotalPages,
		"hasMore":    p.HasMore,
	}
}

func formatTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(time.RFC3339)
}
