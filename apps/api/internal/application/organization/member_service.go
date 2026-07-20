package organization

import (
	"context"
	"errors"
	"log/slog"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/google/uuid"
)

var (
	ErrCannotRemoveLastSuperAdmin = errors.New("cannot remove the last super_admin")
	ErrCannotModifyHigherRole     = errors.New("cannot modify member with equal or higher role")
	ErrNotOrgMember               = errors.New("operator is not a member of this organization")
)

// MemberService handles organization member operations.
type MemberService struct {
	memberRepo organization.MemberRepository
	orgRepo    organization.OrganizationRepository
	logger     *slog.Logger
}

// NewMemberService creates a new MemberService.
func NewMemberService(
	memberRepo organization.MemberRepository,
	orgRepo organization.OrganizationRepository,
	logger *slog.Logger,
) *MemberService {
	if logger == nil {
		logger = slog.Default()
	}
	return &MemberService{
		memberRepo: memberRepo,
		orgRepo:    orgRepo,
		logger:     logger,
	}
}

// AddMember adds a member to an organization (called after invitation acceptance).
func (s *MemberService) AddMember(ctx context.Context, orgID, operatorID string, role organization.OrganizationRole, invitedBy string) (*organization.OrganizationMember, error) {
	// Validate role - super_admin cannot be invited
	if role == organization.RoleSuperAdmin {
		return nil, organization.ErrCannotInviteSuperAdmin
	}

	// Check if already a member
	existing, err := s.memberRepo.FindByOperatorAndOrg(ctx, operatorID, orgID)
	if err != nil && !errors.Is(err, organization.ErrMemberNotFound) {
		return nil, err
	}
	if existing != nil && existing.IsActive() {
		return nil, organization.ErrMemberExists
	}

	// Check org member limit
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		return nil, err
	}

	if !org.CanAddMember() {
		return nil, organization.ErrMaxMembersReached
	}

	// Use domain constructor for new member
	member := organization.NewMember(uuid.New().String(), orgID, operatorID, role)
	member.InvitedBy = &invitedBy

	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	s.logger.Info("member added to organization",
		"org_id", orgID,
		"operator_id", operatorID,
		"role", role,
	)

	return member, nil
}

// RemoveMember removes a member from an organization.
func (s *MemberService) RemoveMember(ctx context.Context, orgID, memberID, actorOperatorID string) error {
	// Get the member to remove
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrMemberNotFound
		}
		return err
	}

	// Verify member belongs to this org
	if member.OrganizationID != orgID {
		return organization.ErrMemberNotFound
	}

	// Check if member is active
	if !member.IsActive() {
		return organization.ErrMemberNotFound
	}

	// Get actor's membership to check permissions
	actorMember, err := s.memberRepo.FindByOperatorAndOrg(ctx, actorOperatorID, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrForbidden
		}
		return err
	}

	// Cannot remove self
	if member.OperatorID == actorOperatorID {
		return errors.New("cannot remove yourself from organization")
	}

	// Cannot remove if actor has equal or lower role
	if actorMember.Role.Level() <= member.Role.Level() {
		return ErrCannotModifyHigherRole
	}

	// If removing super_admin, check if it's the last one
	if member.Role == organization.RoleSuperAdmin {
		superAdminCount, err := s.memberRepo.CountSuperAdminsByOrganization(ctx, orgID)
		if err != nil {
			return err
		}
		if superAdminCount <= 1 {
			return ErrCannotRemoveLastSuperAdmin
		}
	}

	// Soft delete the membership
	if err := s.memberRepo.SoftDelete(ctx, memberID); err != nil {
		return err
	}

	s.logger.Info("member removed from organization",
		"org_id", orgID,
		"member_id", memberID,
		"actor_id", actorOperatorID,
	)

	return nil
}

// UpdateMemberRole updates a member's role.
func (s *MemberService) UpdateMemberRole(ctx context.Context, orgID, memberID, actorOperatorID string, newRole organization.OrganizationRole) (*organization.OrganizationMember, error) {
	// Cannot change to super_admin via update
	if newRole == organization.RoleSuperAdmin {
		return nil, organization.ErrCannotInviteSuperAdmin
	}

	// Get the member to update
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return nil, organization.ErrMemberNotFound
		}
		return nil, err
	}

	// Verify member belongs to this org
	if member.OrganizationID != orgID {
		return nil, organization.ErrMemberNotFound
	}

	// Get actor's membership to check permissions
	actorMember, err := s.memberRepo.FindByOperatorAndOrg(ctx, actorOperatorID, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return nil, organization.ErrForbidden
		}
		return nil, err
	}

	// Cannot modify equal or lower role
	if actorMember.Role.Level() <= member.Role.Level() {
		return nil, ErrCannotModifyHigherRole
	}

	// Actor must have higher level than new role
	if actorMember.Role.Level() <= newRole.Level() {
		return nil, ErrCannotModifyHigherRole
	}

	// Update the role
	member.Role = newRole
	if err := s.memberRepo.Update(ctx, member); err != nil {
		return nil, err
	}

	s.logger.Info("member role updated",
		"org_id", orgID,
		"member_id", memberID,
		"new_role", newRole,
		"actor_id", actorOperatorID,
	)

	return member, nil
}

// TransferOwnership transfers super_admin ownership to another member.
func (s *MemberService) TransferOwnership(ctx context.Context, orgID, currentSuperAdminID, newSuperAdminMemberID string) error {
	// Get the current super_admin
	currentAdmin, err := s.memberRepo.FindByOperatorAndOrg(ctx, currentSuperAdminID, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrForbidden
		}
		return err
	}

	// Must be super_admin to transfer
	if currentAdmin.Role != organization.RoleSuperAdmin {
		return organization.ErrForbidden
	}

	// Get the new super_admin member
	newAdmin, err := s.memberRepo.FindByID(ctx, newSuperAdminMemberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrMemberNotFound
		}
		return err
	}

	// Verify new admin belongs to this org
	if newAdmin.OrganizationID != orgID {
		return organization.ErrMemberNotFound
	}

	// New admin must be active
	if !newAdmin.IsActive() {
		return organization.ErrMemberNotFound
	}

	// Transfer ownership: new admin becomes super_admin, current becomes admin
	newAdmin.Role = organization.RoleSuperAdmin
	if err := s.memberRepo.Update(ctx, newAdmin); err != nil {
		return err
	}

	currentAdmin.Role = organization.RoleAdmin
	if err := s.memberRepo.Update(ctx, currentAdmin); err != nil {
		return err
	}

	s.logger.Info("ownership transferred",
		"org_id", orgID,
		"from", currentSuperAdminID,
		"to", newSuperAdminMemberID,
	)

	return nil
}

// GetMember retrieves a member by ID.
func (s *MemberService) GetMember(ctx context.Context, memberID string) (*organization.OrganizationMember, error) {
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return nil, organization.ErrMemberNotFound
		}
		return nil, err
	}

	// Only return if active
	if !member.IsActive() {
		return nil, organization.ErrMemberNotFound
	}

	return member, nil
}

// ListMembers lists all members of an organization.
func (s *MemberService) ListMembers(ctx context.Context, orgID string) ([]*organization.OrganizationMember, error) {
	members, err := s.memberRepo.FindByOrganization(ctx, orgID)
	if err != nil {
		return nil, err
	}

	// Filter to only active members
	result := make([]*organization.OrganizationMember, 0, len(members))
	for _, m := range members {
		if m.IsActive() {
			result = append(result, m)
		}
	}

	return result, nil
}

// ListMembersPaginated lists all active members of an organization with pagination.
func (s *MemberService) ListMembersPaginated(ctx context.Context, orgID string, page, limit int) (*MemberListResponse, error) {
	// Apply defaults and limits
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	offset := (page - 1) * limit

	// Use repository paginated method - does LIMIT/OFFSET at DB level
	members, total, err := s.memberRepo.FindActiveByOrganizationPaginated(ctx, orgID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &MemberListResponse{
		Items: members,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    page < totalPages,
		},
	}, nil
}

// GetMembership retrieves a specific operator's membership in an org.
func (s *MemberService) GetMembership(ctx context.Context, operatorID, orgID string) (*organization.OrganizationMember, error) {
	member, err := s.memberRepo.FindByOperatorAndOrg(ctx, operatorID, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return nil, ErrNotOrgMember
		}
		return nil, err
	}

	if !member.IsActive() {
		return nil, ErrNotOrgMember
	}

	return member, nil
}

// ListOperatorMemberships lists all memberships for an operator across all orgs.
func (s *MemberService) ListOperatorMemberships(ctx context.Context, operatorID string) ([]*organization.OrganizationMember, error) {
	return s.memberRepo.ListByOperator(ctx, operatorID)
}

// ListOperatorMembershipsPaginated lists memberships with pagination.
func (s *MemberService) ListOperatorMembershipsPaginated(ctx context.Context, operatorID string, page, limit int) (*MembershipListResponse, error) {
	// Apply defaults and limits
	if page <= 0 {
		page = 1
	}
	if limit <= 0 {
		limit = DefaultPageSize
	}
	if limit > MaxPageSize {
		limit = MaxPageSize
	}

	offset := (page - 1) * limit

	// Use repository paginated method - does LIMIT/OFFSET at DB level
	members, total, err := s.memberRepo.ListByOperatorPaginated(ctx, operatorID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &MembershipListResponse{
		Items: members,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    page < totalPages,
		},
	}, nil
}

// CheckCanManageMembers checks if an operator can manage other members in an org.
func (s *MemberService) CheckCanManageMembers(ctx context.Context, operatorID, orgID string) error {
	member, err := s.GetMembership(ctx, operatorID, orgID)
	if err != nil {
		return err
	}

	if !member.Role.CanManageMembers() {
		return organization.ErrForbidden
	}

	return nil
}

// CheckCanManageOrganization checks if an operator can manage organization settings.
func (s *MemberService) CheckCanManageOrganization(ctx context.Context, operatorID, orgID string) error {
	member, err := s.GetMembership(ctx, operatorID, orgID)
	if err != nil {
		return err
	}

	if !member.Role.CanManageOrganization() {
		return organization.ErrForbidden
	}

	return nil
}

// CheckMembership checks if an operator is a member of the organization.
func (s *MemberService) CheckMembership(ctx context.Context, operatorID, orgID string) error {
	_, err := s.GetMembership(ctx, operatorID, orgID)
	return err
}

// SuspendMember suspends a member (only active members can be suspended).
func (s *MemberService) SuspendMember(ctx context.Context, orgID, memberID, actorOperatorID string) error {
	// Get the member to suspend
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrMemberNotFound
		}
		return err
	}

	// Verify member belongs to this org
	if member.OrganizationID != orgID {
		return organization.ErrMemberNotFound
	}

	// Cannot suspend self
	if member.OperatorID == actorOperatorID {
		return errors.New("cannot suspend yourself")
	}

	// Get actor's membership to check permissions
	actorMember, err := s.memberRepo.FindByOperatorAndOrg(ctx, actorOperatorID, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrForbidden
		}
		return err
	}

	// Cannot suspend if actor has equal or lower role
	if actorMember.Role.Level() <= member.Role.Level() {
		return ErrCannotModifyHigherRole
	}

	// Cannot suspend already suspended member
	if member.IsSuspended() {
		return errors.New("member is already suspended")
	}

	// Use domain method to suspend
	if err := member.Suspend(); err != nil {
		return err
	}

	if err := s.memberRepo.Update(ctx, member); err != nil {
		return err
	}

	s.logger.Info("member suspended",
		"org_id", orgID,
		"member_id", memberID,
		"actor_id", actorOperatorID,
	)

	return nil
}

// ReinstateMember reinstates a suspended member back to active.
func (s *MemberService) ReinstateMember(ctx context.Context, orgID, memberID, actorOperatorID string) error {
	// Get the member to reinstate
	member, err := s.memberRepo.FindByID(ctx, memberID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrMemberNotFound
		}
		return err
	}

	// Verify member belongs to this org
	if member.OrganizationID != orgID {
		return organization.ErrMemberNotFound
	}

	// Cannot reinstate self
	if member.OperatorID == actorOperatorID {
		return errors.New("cannot reinstate yourself")
	}

	// Get actor's membership to check permissions
	actorMember, err := s.memberRepo.FindByOperatorAndOrg(ctx, actorOperatorID, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrMemberNotFound) {
			return organization.ErrForbidden
		}
		return err
	}

	// Cannot reinstate if actor has equal or lower role than the target
	if actorMember.Role.Level() <= member.Role.Level() {
		return ErrCannotModifyHigherRole
	}

	// Cannot reinstate non-suspended member
	if !member.IsSuspended() {
		return errors.New("member is not suspended")
	}

	// Use domain method to reinstate
	if err := member.Reinstate(); err != nil {
		return err
	}

	if err := s.memberRepo.Update(ctx, member); err != nil {
		return err
	}

	s.logger.Info("member reinstated",
		"org_id", orgID,
		"member_id", memberID,
		"actor_id", actorOperatorID,
	)

	return nil
}
