package organization

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/invitation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/google/uuid"
)

const (
	// MaxActiveOrgsPerOperator is the maximum number of active organizations an operator can have.
	MaxActiveOrgsPerOperator = 2

	// DefaultOrgMaxMembers is the default maximum members per organization.
	DefaultOrgMaxMembers = 100
)

var (
	ErrMaxOrgsReached = errors.New("maximum 2 active organizations allowed")
)

// OrganizationService handles organization operations.
type OrganizationService struct {
	orgRepo         organization.OrganizationRepository
	memberRepo      organization.MemberRepository
	invitationRepo  invitation.Repository
	txManager       transaction.TxManager
	logger          *slog.Logger
}

// NewOrganizationService creates a new OrganizationService.
func NewOrganizationService(
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	invitationRepo invitation.Repository,
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
		txManager:      txManager,
		logger:          logger,
	}
}

// CreateOrganization creates a new organization with the creator as super_admin.
func (s *OrganizationService) CreateOrganization(ctx context.Context, operatorID, name string, maxMembers int) (*organization.Organization, error) {
	// Validate input
	if name == "" {
		return nil, errors.New("organization name is required")
	}
	if len(name) < 2 {
		return nil, errors.New("organization name must be at least 2 characters")
	}
	if len(name) > 255 {
		return nil, errors.New("organization name exceeds 255 characters")
	}

	// Check max members
	if maxMembers <= 0 {
		maxMembers = DefaultOrgMaxMembers
	}

	var createdOrg *organization.Organization

	// Use transaction to prevent race condition on max orgs
	err := s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// Count active orgs for this operator
		orgs, err := s.orgRepo.ListByOperator(txCtx, operatorID)
		if err != nil {
			return err
		}

		activeCount := 0
		for _, org := range orgs {
			if !org.IsDeleted() {
				activeCount++
			}
		}

		if activeCount >= MaxActiveOrgsPerOperator {
			return ErrMaxOrgsReached
		}

		// Check if org with same name already exists for this operator
		existing, err := s.orgRepo.FindByName(txCtx, operatorID, name)
		if err != nil && !errors.Is(err, organization.ErrNotFound) {
			return err
		}
		if existing != nil {
			return organization.ErrOrganizationExists
		}

		// Create organization using constructor
		orgID := uuid.New().String()
		createdOrg = organization.NewOrganization(orgID, name, operatorID)
		createdOrg.MaxMembers = maxMembers

		if err := s.orgRepo.Create(txCtx, createdOrg); err != nil {
			return err
		}

		// Create membership for creator as super_admin using constructor
		member := organization.NewMember(
			uuid.New().String(),
			createdOrg.ID,
			operatorID,
			organization.RoleSuperAdmin,
		)

		if err := s.memberRepo.Create(txCtx, member); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return createdOrg, nil
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

	// Check if deleted
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
		org.IsActive = *isActive
	}

	org.UpdatedAt = time.Now()

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, err
	}

	return org, nil
}

// DeleteOrganization soft-deletes an organization and cancels all pending invitations.
func (s *OrganizationService) DeleteOrganization(ctx context.Context, orgID string) error {
	// Use transaction to ensure atomic deletion
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// First, expire all pending invitations
		if err := s.invitationRepo.ExpireByOrganization(txCtx, orgID); err != nil {
			s.logger.Error("failed to expire invitations for org", "org_id", orgID, "error", err)
			// Continue anyway - this is a best-effort cleanup
		}

		// Soft delete the organization
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

	// Filter out deleted organizations
	result := make([]*organization.Organization, 0, len(orgs))
	for _, org := range orgs {
		if !org.IsDeleted() {
			result = append(result, org)
		}
	}

	return result, nil
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

	// Get member count
	count, err := s.orgRepo.CountActiveMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	org.MemberCount = count
	return org, nil
}
