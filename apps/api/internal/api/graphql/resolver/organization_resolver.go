// Package resolver provides GraphQL resolver implementations.
package resolver

import (
	"context"
	"errors"
	"fmt"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	orgapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/organization"
	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/graphql-go/graphql"
	"time"
)

// ============================================================.
// Organization Queries.
// ============================================================.

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

	// Check if operator is a member of the organization.
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

	// Parse pagination args.
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method.
	result, err := r.OrgService.ListOrganizationsPaginated(ctx, op.ID, page, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list organizations")
	}

	// Convert to GraphQL response format.
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, org := range result.Items {
		items = append(items, r.orgToMap(org))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.orgPaginationToMap(result.Pagination),
	}, nil
}

// GetMyMemberships resolves the myMemberships query.
func (r *Resolver) GetMyMemberships(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Parse pagination args.
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method.
	result, err := r.MemberService.ListOperatorMembershipsPaginated(ctx, op.ID, page, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list memberships")
	}

	// Convert to GraphQL response format.
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, m := range result.Items {
		items = append(items, r.membershipToMap(m))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.orgPaginationToMap(result.Pagination),
	}, nil
}

// GetOrganizationMembers resolves the organizationMembers query.
func (r *Resolver) GetOrganizationMembers(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is a member.
	err = r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	// Parse pagination args.
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method (filters to active members).
	result, err := r.MemberService.ListMembersPaginated(ctx, orgID, page, limit)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list members")
	}

	// Convert to GraphQL response format.
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, m := range result.Items {
		items = append(items, r.membershipToMap(m))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.orgPaginationToMap(result.Pagination),
	}, nil
}

// GetOrganizationInvitations resolves the organizationInvitations query.
func (r *Resolver) GetOrganizationInvitations(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is a member.
	err = r.MemberService.CheckCanManageOrganization(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	// Parse pagination args.
	page := 1
	if p, ok := p.Args["page"].(int); ok && p > 0 {
		page = p
	}
	limit := 50
	if l, ok := p.Args["limit"].(int); ok && l > 0 {
		limit = l
	}

	// Use paginated service method (filters to pending invitations when status is nil).
	result, err := r.InvitationService.ListInvitationsByOrganizationPaginated(ctx, orgID, page, limit, nil)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list invitations")
	}

	// Convert to GraphQL response format.
	items := make([]map[string]interface{}, 0, len(result.Items))
	for _, inv := range result.Items {
		items = append(items, r.invitationToMap(inv))
	}

	return map[string]interface{}{
		"items":      items,
		"pagination": r.orgPaginationToMap(result.Pagination),
	}, nil
}

// GetMyInvitations resolves the myInvitations query.
func (r *Resolver) GetMyInvitations(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Validate operator has an email set.
	if op.Email == "" {
		// Return empty array if no email - user needs to set their email first.
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

// ============================================================.
// Organization Mutations.
// ============================================================.

// CreateOrganization resolves the createOrganization mutation.
func (r *Resolver) CreateOrganization(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	name, ok := p.Args["name"].(string)
	if !ok || name == "" {
		return nil, r.Presenter.BadRequestError("name is required")
	}

	maxMembers, ok := p.Args["maxMembers"].(int)
	if !ok || maxMembers <= 0 {
		maxMembers = 100
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Note: CreateOrganization creates the org AND the creator's membership as super_admin.
	// The membership is created inside the transaction, so we get it from there.
	org, err := r.OrgService.CreateOrganization(ctx, op.ID, name, "", maxMembers, "super_admin")
	if err != nil {
		if errors.Is(err, orgapp.ErrMaxOrgsReached) {
			return nil, r.Presenter.BadRequestError("maximum 2 active organizations allowed")
		}
		if errors.Is(err, orgdomain.ErrOrganizationExists) {
			return nil, r.Presenter.BadRequestError("organization with this name already exists")
		}
		return nil, r.Presenter.InternalError("failed to create organization")
	}

	// Get the creator's membership.
	membership, err := r.MemberService.GetMembership(ctx, op.ID, org.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to get membership")
	}

	// Publish organization created event.
	if r.Hub != nil {
		r.Hub.PublishOrganizationEvent(org.ID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "created",
			"organizationId": org.ID,
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

	
	id, ok := p.Args["id"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("id must be a string")
	}
	name, ok := p.Args["name"].(string)
	if !ok {
		name = "" // name is optional.
	}
	maxMembers, ok := p.Args["maxMembers"].(int)
	if !ok || maxMembers <= 0 {
		maxMembers = 0 // 0 means don't update.
	}
	isActive, ok := p.Args["isActive"].(bool)
	if !ok {
		isActive = false // false means don't update.
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator has admin rights.
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, id); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to update organization")
	}

	// Prepare pointer arguments for optional fields.
	var namePtr *string
	if name != "" {
		namePtr = &name
	}
	var maxMembersPtr *int
	if maxMembers > 0 {
		maxMembersPtr = &maxMembers
	}
	var isActivePtr *bool
	if p.Args["isActive"] != nil {
		isActivePtr = &isActive
	}

	org, err := r.OrgService.UpdateOrganization(ctx, id, namePtr, maxMembersPtr, isActivePtr)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update organization")
	}

	// Publish organization updated event.
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

	// Check if operator is a super_admin of this organization.
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

	// Publish organization deleted event.
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

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}
	email, ok := p.Args["email"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("email must be a string")
	}
	roleStr, ok := p.Args["role"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("role must be a string")
	}
	notes, ok := p.Args["notes"].(string)
	if !ok {
		notes = "" // optional field.
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator can manage members.
	err = r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to invite members")
	}

	// Convert role string to OrganizationRole.
	role := orgdomain.OrganizationRole(roleStr)

	invitation, err := r.InvitationService.CreateInvitation(ctx, orgID, op.ID, email, role, notes)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to create invitation")
	}

	// Publish member invited event.
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "member_invited",
			"organizationId": orgID,
			"memberId":       invitation.ID,
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

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}
	memberID, ok := p.Args["memberId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("memberId must be a string")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Get the actor's membership to check if trying to remove self.
	actorMembership, err := r.MemberService.GetMembership(ctx, op.ID, orgID)
	if err != nil {
		// Not a member - can't remove anyone.
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	// Prevent self-removal - compare actor's membership ID with target.
	if actorMembership.ID == memberID {
		return nil, r.Presenter.BadRequestError("cannot remove yourself from the organization")
	}

	// Prevent removing the last super_admin.
	canRemove, err := r.canRemoveMember(ctx, orgID, memberID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to check permissions")
	}
	if !canRemove {
		return nil, r.Presenter.BadRequestError("cannot remove the last super_admin")
	}

	// Check if operator can manage members.
	if err = r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to remove members")
	}

	err = r.MemberService.RemoveMember(ctx, orgID, memberID, op.ID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to remove member")
	}

	// Publish member removed event.
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

	// Find the target membership to check its role.
	var targetRole orgdomain.OrganizationRole
	var targetLifecycle orgdomain.MemberLifecycle
	var memberFound bool
	for _, m := range memberships {
		if m.ID == memberID {
			targetRole = m.Role
			targetLifecycle = m.Lifecycle
			memberFound = true
			break
		}
	}

	// If member doesn't exist, don't allow removal.
	if !memberFound {
		return false, nil
	}

	// If target is not a super_admin or not active, removal is always safe.
	if targetRole != orgdomain.RoleSuperAdmin || targetLifecycle != orgdomain.MemberLifecycleActive {
		return true, nil
	}

	// Count other active super_admins.
	superAdminCount := 0
	for _, m := range memberships {
		if m.Role == orgdomain.RoleSuperAdmin && m.ID != memberID && m.Lifecycle == orgdomain.MemberLifecycleActive {
			superAdminCount++
		}
	}

	return superAdminCount > 0, nil
}

// UpdateMemberRole resolves the updateMemberRole mutation.
func (r *Resolver) UpdateMemberRole(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}
	memberID, ok := p.Args["memberId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("memberId must be a string")
	}
	roleStr, ok := p.Args["role"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("role must be a string")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	memberships, err := r.MemberService.ListMembers(ctx, orgID)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to list members")
	}

	if r.isSelfRoleChange(memberships, memberID, op.ID) {
		return nil, r.Presenter.BadRequestError("cannot change your own role")
	}

	if err = r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID); err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to update member role")
	}

	if err = r.checkDemotionSafety(ctx, orgID, memberID, roleStr); err != nil {
		return nil, err
	}

	role := orgdomain.OrganizationRole(roleStr)
	oldRole := r.findMemberRoleByID(memberships, memberID)

	membership, err := r.MemberService.UpdateMemberRole(ctx, orgID, memberID, op.ID, role)
	if err != nil {
		return nil, r.Presenter.InternalError("failed to update member role")
	}

	r.publishRoleChangedEvent(orgID, memberID, op.ID, oldRole, string(role))

	return r.membershipToMap(membership), nil
}

// isSelfRoleChange checks if the operator is trying to change their own role.
func (r *Resolver) isSelfRoleChange(memberships []*orgdomain.OrganizationMember, memberID, operatorID string) bool {
	for _, m := range memberships {
		if m.OperatorID == operatorID && m.ID == memberID {
			return true
		}
	}
	return false
}

// checkDemotionSafety checks if demoting the member is safe.
func (r *Resolver) checkDemotionSafety(ctx context.Context, orgID, memberID, roleStr string) error {
	isDemotion := roleStr == string(orgdomain.RoleAdmin) || roleStr == string(orgdomain.RoleOperator) || roleStr == string(orgdomain.RoleViewer)
	if !isDemotion {
		return nil
	}

	canChange, err := r.canDemoteSuperAdmin(ctx, orgID, memberID)
	if err != nil {
		return r.Presenter.InternalError("failed to check permissions")
	}
	if !canChange {
		return r.Presenter.BadRequestError("cannot demote the last super_admin")
	}
	return nil
}

// findMemberRoleByID finds the current role of a member by ID.
func (r *Resolver) findMemberRoleByID(memberships []*orgdomain.OrganizationMember, memberID string) string {
	for _, m := range memberships {
		if m.ID == memberID {
			return string(m.Role)
		}
	}
	return ""
}

// publishRoleChangedEvent publishes a role changed event to the hub.
func (r *Resolver) publishRoleChangedEvent(orgID, memberID, operatorID, oldRole, newRole string) {
	if r.Hub == nil {
		return
	}
	r.Hub.PublishMemberEvent(orgID, map[string]interface{}{
		"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
		"type":           "role_changed",
		"organizationId": orgID,
		"memberId":       memberID,
		"operatorId":     operatorID,
		"timestamp":      time.Now().UTC(),
		"data": map[string]interface{}{
			"oldRole": oldRole,
			"newRole": newRole,
		},
	})
}

// canDemoteSuperAdmin checks if demoting the target member would leave the org without a super_admin.
func (r *Resolver) canDemoteSuperAdmin(ctx context.Context, orgID, memberID string) (bool, error) {
	memberships, err := r.MemberService.ListMembers(ctx, orgID)
	if err != nil {
		return false, err
	}

	// Count super_admins excluding the member being demoted.
	superAdminCount := 0
	for _, m := range memberships {
		if m.Role == orgdomain.RoleSuperAdmin && m.ID != memberID && m.Lifecycle == orgdomain.MemberLifecycleActive {
			superAdminCount++
		}
	}

	return superAdminCount > 0, nil
}

// AcceptInvitation resolves the acceptInvitation mutation.
func (r *Resolver) AcceptInvitation(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	token, ok := p.Args["token"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("token must be a string")
	}
	notes, ok := p.Args["notes"].(string)
	if !ok {
		notes = "" // optional field.
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Validate operator has an email set.
	if op.Email == "" {
		return nil, r.Presenter.BadRequestError("operator email not set, please update your profile first")
	}

	// AcceptInvitation now returns the membership directly within the transaction.
	membership, err := r.InvitationService.AcceptInvitation(ctx, token, op.ID, op.Email, notes)
	if err != nil {
		// Map service errors to proper GraphQL errors.
		switch err {
		case orgdomain.ErrInvitationNotFound, orgdomain.ErrAlreadyResponded:
			return nil, r.Presenter.NotFoundError(err.Error())
		case orgdomain.ErrInvitationExpired:
			return nil, r.Presenter.BadRequestError("invitation has expired")
		case orgdomain.ErrEmailMismatch:
			return nil, r.Presenter.ForbiddenError("this invitation was sent to a different email address")
		default:
			return nil, r.Presenter.InternalError("failed to accept invitation: " + err.Error())
		}
	}

	// Publish member joined event.
	if r.Hub != nil {
		r.Hub.PublishMemberEvent(membership.OrganizationID, map[string]interface{}{
			"id":             fmt.Sprintf("evt-%d", time.Now().UnixNano()),
			"type":           "member_joined",
			"organizationId": membership.OrganizationID,
			"memberId":       membership.ID,
			"operatorId":     op.ID,
			"timestamp":      time.Now().UTC(),
		})
	}

	// Membership is returned directly from the transaction - no race condition.
	return r.membershipToMap(membership), nil
}

// RejectInvitation resolves the rejectInvitation mutation.
func (r *Resolver) RejectInvitation(p graphql.ResolveParams) (interface{}, error) {
	ctx := p.Context

	token, ok := p.Args["token"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("token must be a string")
	}
	notes, ok := p.Args["notes"].(string)
	if !ok {
		notes = "" // optional field.
	}
	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Validate operator has an email set.
	if op.Email == "" {
		return nil, r.Presenter.BadRequestError("operator email not set, please update your profile first")
	}

	err := r.InvitationService.RejectInvitation(ctx, token, op.ID, op.Email, notes)
	if err != nil {
		// Map service errors to proper GraphQL errors.
		switch err {
		case orgdomain.ErrInvitationNotFound, orgdomain.ErrAlreadyResponded:
			return nil, r.Presenter.NotFoundError(err.Error())
		case orgdomain.ErrInvitationExpired:
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

	// Get the invitation to check org membership.
	inv, err := r.InvitationService.GetInvitationByID(ctx, id)
	if err != nil {
		return nil, r.Presenter.NotFoundError("invitation not found")
	}

	// Check if operator can manage members in this org.
	if err = r.MemberService.CheckCanManageMembers(ctx, op.ID, inv.OrganizationID); err != nil {
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

	imei, ok := p.Args["imei"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("imei must be a string")
	}
	sourceOrgID, ok := p.Args["sourceOrganizationId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("sourceOrganizationId must be a string")
	}
	targetOrgID, ok := p.Args["targetOrganizationId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("targetOrganizationId must be a string")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check membership in source org.
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, sourceOrgID); err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of source organization")
	}

	// Check membership in target org.
	if err := r.MemberService.CheckCanManageOrganization(ctx, op.ID, targetOrgID); err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of target organization")
	}

	if err := r.DeviceService.TransferDevice(ctx, imei, sourceOrgID, targetOrgID, op.ID); err != nil {
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

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}
	memberID, ok := p.Args["memberId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("memberId must be a string")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator is super_admin of this org.
	membership, err := r.MemberService.GetMembership(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("not a member of this organization")
	}

	if membership.Role != orgdomain.RoleSuperAdmin {
		return nil, r.Presenter.ForbiddenError("only super_admin can transfer ownership")
	}

	// Transfer ownership.
	err = r.MemberService.TransferOwnership(ctx, orgID, op.ID, memberID)
	if err != nil {
		if err == orgdomain.ErrMemberNotFound {
			return nil, r.Presenter.NotFoundError("member not found")
		}
		return nil, r.Presenter.InternalError("failed to transfer ownership")
	}

	// Publish ownership transferred event.
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

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}
	memberID, ok := p.Args["memberId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("memberId must be a string")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator can manage members.
	err = r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to suspend member")
	}

	err = r.MemberService.SuspendMember(ctx, orgID, memberID, op.ID)
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

	// Publish member suspended event.
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

	orgID, err := r.resolveOrgID(p)
	if err != nil {
		return nil, err
	}
	memberID, ok := p.Args["memberId"].(string)
	if !ok {
		return nil, r.Presenter.BadRequestError("memberId must be a string")
	}

	op, ok := gqlcontext.GetOperator(ctx)
	if !ok || op == nil {
		return nil, r.Presenter.UnauthorizedError()
	}

	// Check if operator can manage members.
	err = r.MemberService.CheckCanManageMembers(ctx, op.ID, orgID)
	if err != nil {
		return nil, r.Presenter.ForbiddenError("insufficient permissions to reinstate member")
	}

	err = r.MemberService.ReinstateMember(ctx, orgID, memberID, op.ID)
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

	// Publish member reinstated event.
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

// ============================================================.
// Helper Functions.
// ============================================================.

func (r *Resolver) orgToMap(org *orgdomain.Organization) map[string]interface{} {
	result := map[string]interface{}{
		"id":          org.ID,
		"name":        org.Name,
		"lifecycle":   string(org.Lifecycle),
		"maxMembers":  org.MaxMembers,
		"memberCount": org.MemberCount,
		"createdBy":   org.CreatedBy,
		"createdAt":   formatTime(org.CreatedAt),
		"updatedAt":   formatTime(org.UpdatedAt),
	}

	if org.DeletedAt != nil && !org.DeletedAt.IsZero() {
		result["deletedAt"] = formatTime(*org.DeletedAt)
	}

	return result
}

func (r *Resolver) membershipToMap(m *orgdomain.OrganizationMember) map[string]interface{} {
	result := map[string]interface{}{
		"id":             m.ID,
		"organizationId": m.OrganizationID,
		"operatorId":     m.OperatorID,
		"role":            string(m.Role),
		"lifecycle":       string(m.Lifecycle),
	}

	if !m.JoinedAt.IsZero() {
		result["joinedAt"] = formatTime(m.JoinedAt)
	}

	if m.RemovedAt != nil && !m.RemovedAt.IsZero() {
		result["removedAt"] = formatTime(*m.RemovedAt)
	}

	if m.SuspendedAt != nil && !m.SuspendedAt.IsZero() {
		result["suspendedAt"] = formatTime(*m.SuspendedAt)
	}

	return result
}

func (r *Resolver) invitationToMap(inv *orgdomain.Invitation) map[string]interface{} {
	result := map[string]interface{}{
		"id":               inv.ID,
		"organizationId":    inv.OrganizationID,
		"organizationName": inv.OrganizationName,
		"email":            inv.Email,
		"role":             string(inv.Role),
		"status":           string(inv.Status),
		"token":            inv.Token,
		"inviterId":        inv.InvitedBy,
		"inviterName":      inv.InviterName,
		"invitedAt":        formatTime(inv.InvitedAt),
		"expiresAt":        formatTime(inv.ExpiresAt),
	}

	if inv.InviteeNotes != "" {
		result["inviteeNotes"] = inv.InviteeNotes
	}

	if inv.InviterNotes != "" {
		result["inviterNotes"] = inv.InviterNotes
	}

	if inv.RespondedAt != nil && !inv.RespondedAt.IsZero() {
		result["respondedAt"] = formatTime(*inv.RespondedAt)
	}

	return result
}

func (r *Resolver) orgPaginationToMap(p orgapp.Pagination) map[string]interface{} {
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
