package organization

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/google/uuid"
)

// Pagination constants.
const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

// Organization limits.
const (
	// MaxActiveOrgsPerOperator is the maximum number of active organizations an operator can have.
	MaxActiveOrgsPerOperator = 2
	// DefaultOrgMaxMembers is the default maximum members per organization.
	DefaultOrgMaxMembers = 100
)

var (
	ErrMaxOrgsReached = errors.New("maximum 2 active organizations allowed")
)

// Pagination represents pagination metadata.
type Pagination struct {
	Page       int  `json:"page"`
	Limit      int  `json:"limit"`
	Total      int  `json:"total"`
	TotalPages int  `json:"totalPages"`
	HasMore    bool `json:"hasMore"`
}

// OrganizationListResponse represents a paginated list of organizations.
type OrganizationListResponse struct {
	Items      []*organization.Organization `json:"items"`
	Pagination Pagination                   `json:"pagination"`
}

// MemberListResponse represents a paginated list of members.
type MemberListResponse struct {
	Items      []*organization.OrganizationMember `json:"items"`
	Pagination Pagination                         `json:"pagination"`
}

// InvitationListResponse represents a paginated list of invitations.
type InvitationListResponse struct {
	Items      []*organization.Invitation `json:"items"`
	Pagination Pagination                 `json:"pagination"`
}

// MembershipListResponse represents a paginated list of memberships.
type MembershipListResponse struct {
	Items      []*organization.OrganizationMember `json:"items"`
	Pagination Pagination                         `json:"pagination"`
}

// OrganizationService handles organization operations.
type OrganizationService struct {
	orgRepo        organization.OrganizationRepository
	memberRepo     organization.MemberRepository
	invitationRepo organization.InvitationRepository
	operatorRepo   operator.Repository
	sessionRepo    session.Repository
	deviceRepo     device.Repository
	telemetryRepo  telemetry.Repository
	commandRepo    command.Repository
	txManager      transaction.TxManager
	logger         *slog.Logger
}

// NewOrganizationService creates a new OrganizationService.
func NewOrganizationService(
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	invitationRepo organization.InvitationRepository,
	operatorRepo operator.Repository,
	sessionRepo session.Repository,
	deviceRepo device.Repository,
	telemetryRepo telemetry.Repository,
	commandRepo command.Repository,
	txManager transaction.TxManager,
	logger *slog.Logger,
) *OrganizationService {
	if logger == nil {
		logger = slog.Default()
	}
	return &OrganizationService{
		orgRepo:        orgRepo,
		memberRepo:     memberRepo,
		invitationRepo: invitationRepo,
		operatorRepo:   operatorRepo,
		sessionRepo:    sessionRepo,
		deviceRepo:     deviceRepo,
		telemetryRepo:  telemetryRepo,
		commandRepo:    commandRepo,
		txManager:      txManager,
		logger:         logger,
	}
}

// CreateOrganization creates a new organization with the creator as a member with the specified role.
func (s *OrganizationService) CreateOrganization(ctx context.Context, operatorID, name, description string, maxMembers int, role string) (*organization.Organization, error) {
	name, maxMembers, memberRole, err := s.validateAndPrepareOrg(name, maxMembers, role)
	if err != nil {
		return nil, err
	}

	createdOrg, err := s.createOrgWithTx(ctx, operatorID, name, description, maxMembers, memberRole)
	if err != nil {
		return nil, err
	}

	s.updateOperatorLastOrg(ctx, operatorID, createdOrg.ID)

	return createdOrg, nil
}

// validateAndPrepareOrg validates inputs and prepares organization creation parameters.
func (s *OrganizationService) validateAndPrepareOrg(name string, maxMembers int, role string) (string, int, organization.OrganizationRole, error) {
	if name == "" {
		name = "personal"
	}
	if len(name) < 2 {
		return name, maxMembers, organization.RoleAdmin, errors.New("organization name must be at least 2 characters")
	}
	if len(name) > 255 {
		return name, maxMembers, organization.RoleAdmin, errors.New("organization name exceeds 255 characters")
	}

	var memberRole organization.OrganizationRole
	switch role {
	case "super_admin":
		memberRole = organization.RoleSuperAdmin
	case "admin":
		memberRole = organization.RoleAdmin
	default:
		return name, maxMembers, memberRole, errors.New("role must be 'super_admin' or 'admin'")
	}

	if maxMembers <= 0 {
		maxMembers = DefaultOrgMaxMembers
	}

	return name, maxMembers, memberRole, nil
}

// createOrgWithTx creates the organization within a transaction.
func (s *OrganizationService) createOrgWithTx(ctx context.Context, operatorID, name, description string, maxMembers int, memberRole organization.OrganizationRole) (*organization.Organization, error) {
	var createdOrg *organization.Organization

	txErr := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		activeCount, countErr := s.countActiveOrgs(txCtx, operatorID)
		if countErr != nil {
			return countErr
		}
		if activeCount >= MaxActiveOrgsPerOperator {
			return ErrMaxOrgsReached
		}

		if err := s.checkOrgNameUnique(txCtx, operatorID, name); err != nil {
			return err
		}

		var buildErr error
		createdOrg, buildErr = s.buildAndCreateOrg(txCtx, operatorID, name, description, maxMembers, memberRole)
		return buildErr
	})

	return createdOrg, txErr
}

// countActiveOrgs counts active organizations for an operator.
func (s *OrganizationService) countActiveOrgs(ctx context.Context, operatorID string) (int, error) {
	orgs, err := s.orgRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, org := range orgs {
		if !org.IsDeleted() {
			count++
		}
	}
	return count, nil
}

// checkOrgNameUnique checks if an organization name is unique for the operator.
func (s *OrganizationService) checkOrgNameUnique(ctx context.Context, operatorID, name string) error {
	existing, err := s.orgRepo.FindByName(ctx, operatorID, name)
	if err != nil && !errors.Is(err, organization.ErrNotFound) {
		return err
	}
	if existing != nil {
		return organization.ErrOrganizationExists
	}
	return nil
}

// buildAndCreateOrg builds and creates the organization with its initial member.
func (s *OrganizationService) buildAndCreateOrg(ctx context.Context, operatorID, name, description string, maxMembers int, memberRole organization.OrganizationRole) (*organization.Organization, error) {
	orgID := uuid.New().String()
	org := organization.NewOrganization(orgID, name, operatorID)
	org.MaxMembers = maxMembers
	org.Description = description

	if err := s.orgRepo.Create(ctx, org); err != nil {
		return nil, err
	}

	member := organization.NewMember(uuid.New().String(), org.ID, operatorID, memberRole)
	if err := s.memberRepo.Create(ctx, member); err != nil {
		return nil, err
	}

	return org, nil
}

// updateOperatorLastOrg updates the operator's LastOrganizationID.
func (s *OrganizationService) updateOperatorLastOrg(ctx context.Context, operatorID, orgID string) {
	if s.operatorRepo == nil {
		return
	}
	op, opErr := s.operatorRepo.FindByID(ctx, operatorID)
	if opErr != nil || op.LastOrganizationID == orgID {
		return
	}
	op.LastOrganizationID = orgID
	if err := s.operatorRepo.Update(ctx, op); err != nil {
		s.logger.Warn("failed to update operator LastOrganizationID", "operatorID", operatorID, "error", err)
	}
}

// GetOrganization retrieves an organization by ID.
func (s *OrganizationService) GetOrganization(ctx context.Context, orgID string) (*organization.Organization, error) {
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			return nil, organization.ErrNotFound
		}
		return nil, err
	}

	// Check if deleted.
	if org.IsDeleted() {
		return nil, organization.ErrNotFound
	}

	return org, nil
}

// UpdateOrganization updates an organization.
func (s *OrganizationService) UpdateOrganization(ctx context.Context, orgID string, name *string, maxMembers *int, isActive *bool) (*organization.Organization, error) {
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			return nil, organization.ErrNotFound
		}
		return nil, err
	}

	if name != nil {
		if len(*name) < 2 {
			return nil, errors.New("organization name must be at least 2 characters")
		}
		if len(*name) > 255 {
			return nil, errors.New("organization name exceeds 255 characters")
		}
		org.Name = *name
	}

	if maxMembers != nil {
		if *maxMembers < 0 {
			return nil, errors.New("max members cannot be negative")
		}
		org.MaxMembers = *maxMembers
	}

	if isActive != nil {
		if *isActive {
			if err := org.Activate(); err != nil {
				return nil, err
			}
		} else {
			if err := org.Deactivate(); err != nil {
				return nil, err
			}
		}
	}

	org.UpdatedAt = time.Now()

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// DeleteOrganization soft-deletes an organization and all its resources.
func (s *OrganizationService) DeleteOrganization(ctx context.Context, orgID string) error {
	// Get all members before deleting them so we can revoke their sessions.
	members, err := s.memberRepo.FindByOrganization(ctx, orgID)
	if err != nil {
		s.logger.Error("failed to get members for session revocation", "org_id", orgID, "error", err)
		// Continue anyway - we still want to delete the org.
	}

	// Get all devices for this organization before soft-deleting them.
	var deviceIDs []string
	if s.deviceRepo != nil {
		devices, err := s.deviceRepo.ListByOrganization(ctx, orgID)
		if err != nil {
			s.logger.Error("failed to get devices for org", "org_id", orgID, "error", err)
		} else {
			for _, d := range devices {
				deviceIDs = append(deviceIDs, d.ID)
			}
		}
	}

	// Use transaction to ensure atomic deletion.
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Revoke sessions for all members first.
		for _, member := range members {
			if err := s.sessionRepo.RevokeAllOperatorSessions(txCtx, member.OperatorID); err != nil {
				s.logger.Error("failed to revoke sessions for member",
					"org_id", orgID,
					"operator_id", member.OperatorID,
					"error", err)
				// Continue anyway - best effort cleanup.
			}
		}

		// Delete telemetry for all devices in this org.
		if len(deviceIDs) > 0 && s.telemetryRepo != nil {
			deleted, err := s.telemetryRepo.DeleteByDeviceIDs(txCtx, deviceIDs)
			if err != nil {
				s.logger.Error("failed to delete telemetry for org",
					"org_id", orgID,
					"count", deleted,
					"error", err)
				// Continue anyway - best effort cleanup.
			} else {
				s.logger.Info("deleted telemetry for org", "org_id", orgID, "count", deleted)
			}
		}

		// Delete commands for all devices in this org.
		if len(deviceIDs) > 0 && s.commandRepo != nil {
			deleted, err := s.commandRepo.DeleteByDeviceIDs(txCtx, deviceIDs)
			if err != nil {
				s.logger.Error("failed to delete commands for org",
					"org_id", orgID,
					"count", deleted,
					"error", err)
				// Continue anyway - best effort cleanup.
			} else {
				s.logger.Info("deleted commands for org", "org_id", orgID, "count", deleted)
			}
		}

		// Soft-delete all devices in this org.
		if len(deviceIDs) > 0 && s.deviceRepo != nil {
			now := time.Now()
			deregisteredAt := now.UnixMilli()
			deletionScheduledAt := now.Add(30 * 24 * time.Hour).UnixMilli()
			deleted, err := s.deviceRepo.SoftDeleteByOrganization(txCtx, orgID, deregisteredAt, deletionScheduledAt)
			if err != nil {
				s.logger.Error("failed to soft-delete devices for org",
					"org_id", orgID,
					"count", deleted,
					"error", err)
				// Continue anyway - best effort cleanup.
			} else {
				s.logger.Info("soft-deleted devices for org", "org_id", orgID, "count", deleted)
			}
		}

		// Soft-delete all memberships (cascade).
		if err := s.memberRepo.SoftDeleteByOrganization(txCtx, orgID); err != nil {
			s.logger.Error("failed to remove members from org", "org_id", orgID, "error", err)
			// Continue anyway - this is a best-effort cleanup.
		}

		// Expire all pending invitations.
		if err := s.invitationRepo.ExpireByOrganization(txCtx, orgID); err != nil {
			s.logger.Error("failed to expire invitations for org", "org_id", orgID, "error", err)
			// Continue anyway - this is a best-effort cleanup.
		}

		// Soft delete the organization.
		if err := s.orgRepo.SoftDelete(txCtx, orgID); err != nil {
			if errors.Is(err, organization.ErrNotFound) {
				return organization.ErrNotFound
			}
			return err
		}

		s.logger.Info("organization deleted", "org_id", orgID)
		return nil
	})
}

// ListOrganizations lists all organizations for an operator.
func (s *OrganizationService) ListOrganizations(ctx context.Context, operatorID string) ([]*organization.Organization, error) {
	orgs, err := s.orgRepo.ListByOperator(ctx, operatorID)
	if err != nil {
		return nil, err
	}

	// Filter out deleted organizations.
	result := make([]*organization.Organization, 0, len(orgs))
	for _, org := range orgs {
		if !org.IsDeleted() {
			result = append(result, org)
		}
	}

	return result, nil
}

// ListOrganizationsPaginated lists organizations with pagination.
func (s *OrganizationService) ListOrganizationsPaginated(ctx context.Context, operatorID string, page, limit int) (*OrganizationListResponse, error) {
	// Apply defaults and limits.
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

	// Use repository paginated method - does LIMIT/OFFSET at DB level.
	orgs, total, err := s.orgRepo.ListByOperatorPaginated(ctx, operatorID, limit, offset)
	if err != nil {
		return nil, err
	}

	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}

	return &OrganizationListResponse{
		Items: orgs,
		Pagination: Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			HasMore:    page < totalPages,
		},
	}, nil
}

// GetOrganizationWithMembers retrieves an organization with its member count.
func (s *OrganizationService) GetOrganizationWithMembers(ctx context.Context, orgID string) (*organization.Organization, error) {
	org, err := s.orgRepo.FindByID(ctx, orgID)
	if err != nil {
		if errors.Is(err, organization.ErrNotFound) {
			return nil, organization.ErrNotFound
		}
		return nil, err
	}

	// Get member count.
	count, err := s.orgRepo.CountActiveMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	org.MemberCount = count
	return org, nil
}
