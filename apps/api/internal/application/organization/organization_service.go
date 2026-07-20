package organization

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/transaction"
	"github.com/google/uuid"
)

// Pagination constants
const (
	DefaultPageSize = 50
	MaxPageSize     = 100
)

// Organization limits
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
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
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
	Pagination Pagination                          `json:"pagination"`
}

// InvitationListResponse represents a paginated list of invitations.
type InvitationListResponse struct {
	Items      []*organization.Invitation `json:"items"`
	Pagination Pagination                `json:"pagination"`
}

// MembershipListResponse represents a paginated list of memberships.
type MembershipListResponse struct {
	Items      []*organization.OrganizationMember `json:"items"`
	Pagination Pagination                          `json:"pagination"`
}

// OrganizationService handles organization operations.
type OrganizationService struct {
	orgRepo         organization.OrganizationRepository
	memberRepo      organization.MemberRepository
	invitationRepo  organization.InvitationRepository
	operatorRepo    operator.Repository
	txManager       transaction.TxManager
	logger          *slog.Logger
}

// NewOrganizationService creates a new OrganizationService.
func NewOrganizationService(
	orgRepo organization.OrganizationRepository,
	memberRepo organization.MemberRepository,
	invitationRepo organization.InvitationRepository,
	operatorRepo operator.Repository,
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
		txManager:      txManager,
		logger:          logger,
	}
}

// CreateOrganization creates a new organization with the creator as a member with the specified role.
func (s *OrganizationService) CreateOrganization(ctx context.Context, operatorID, name, description string, maxMembers int, role string) (*organization.Organization, error) {
	// Validate input
	if name == "" {
		name = "personal" // Default name
	}
	if len(name) < 2 {
		return nil, errors.New("organization name must be at least 2 characters")
	}
	if len(name) > 255 {
		return nil, errors.New("organization name exceeds 255 characters")
	}

	// Validate role
	var memberRole organization.OrganizationRole
	switch role {
	case "super_admin":
		memberRole = organization.RoleSuperAdmin
	case "admin":
		memberRole = organization.RoleAdmin
	default:
		return nil, errors.New("role must be 'super_admin' or 'admin'")
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
		createdOrg.Description = description

		if err := s.orgRepo.Create(txCtx, createdOrg); err != nil {
			return err
		}

		// Create membership for creator with the specified role
		member := organization.NewMember(
			uuid.New().String(),
			createdOrg.ID,
			operatorID,
			memberRole,
		)

		if err := s.memberRepo.Create(txCtx, member); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Update operator LastOrganizationID for auto-selection on next login
	if s.operatorRepo != nil {
		op, opErr := s.operatorRepo.FindByID(ctx, operatorID)
		if opErr == nil && op.LastOrganizationID != createdOrg.ID {
			op.LastOrganizationID = createdOrg.ID
			_ = s.operatorRepo.Update(ctx, op)
		}
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

// DeleteOrganization soft-deletes an organization and all its memberships, cancels all pending invitations.
func (s *OrganizationService) DeleteOrganization(ctx context.Context, orgID string) error {
	// Use transaction to ensure atomic deletion
	return s.txManager.WithTx(ctx, func(txCtx context.Context) error {
		// First, soft-delete all memberships (cascade)
		if err := s.memberRepo.SoftDeleteByOrganization(txCtx, orgID); err != nil {
			s.logger.Error("failed to remove members from org", "org_id", orgID, "error", err)
			// Continue anyway - this is a best-effort cleanup
		}

		// Expire all pending invitations
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

// ListOrganizationsPaginated lists organizations with pagination.
func (s *OrganizationService) ListOrganizationsPaginated(ctx context.Context, operatorID string, page, limit int) (*OrganizationListResponse, error) {
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

	// Get member count
	count, err := s.orgRepo.CountActiveMembers(ctx, orgID)
	if err != nil {
		return nil, err
	}

	org.MemberCount = count
	return org, nil
}
