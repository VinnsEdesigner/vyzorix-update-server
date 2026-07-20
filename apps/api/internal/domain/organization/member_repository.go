package organization

import (
	"context"
)

// MemberRepository defines the interface for organization member data access.
type MemberRepository interface {
	// Create creates a new membership.
	Create(ctx context.Context, member *OrganizationMember) error

	// FindByID retrieves a member by ID.
	FindByID(ctx context.Context, id string) (*OrganizationMember, error)

	// FindByMemberID retrieves a member by MemberID value object.
	FindByMemberID(ctx context.Context, id MemberID) (*OrganizationMember, error)

	// FindByOperatorAndOrg finds a membership by operator ID and organization ID.
	FindByOperatorAndOrg(ctx context.Context, operatorID, orgID string) (*OrganizationMember, error)

	// FindByOperatorAndOrgID finds a membership by OperatorID and OrganizationID value objects.
	FindByOperatorAndOrgID(ctx context.Context, operatorID OperatorID, orgID OrganizationID) (*OrganizationMember, error)

	// FindByOrganization lists all members of an organization.
	FindByOrganization(ctx context.Context, orgID string) ([]*OrganizationMember, error)

	// FindByOrganizationID lists all members by OrganizationID value object.
	FindByOrganizationID(ctx context.Context, orgID OrganizationID) ([]*OrganizationMember, error)

	// FindActiveByOrganization lists all active members of an organization.
	FindActiveByOrganization(ctx context.Context, orgID string) ([]*OrganizationMember, error)

	// FindActiveByOrganizationPaginated lists active members with pagination.
	FindActiveByOrganizationPaginated(ctx context.Context, orgID string, limit, offset int) ([]*OrganizationMember, int, error)

	// FindActiveByOrganizationID lists all active members by OrganizationID.
	FindActiveByOrganizationID(ctx context.Context, orgID OrganizationID) ([]*OrganizationMember, error)

	// Update updates an existing membership.
	Update(ctx context.Context, member *OrganizationMember) error

	// SoftDelete soft-deletes a membership (marks as removed).
	SoftDelete(ctx context.Context, id string) error

	// SoftDeleteByMemberID soft-deletes a membership by MemberID.
	SoftDeleteByMemberID(ctx context.Context, id MemberID) error

	// SoftDeleteByOperator soft-deletes all memberships for an operator (used during operator deletion).
	SoftDeleteByOperator(ctx context.Context, operatorID string) error

	// SoftDeleteByOperatorID soft-deletes all memberships for an OperatorID.
	SoftDeleteByOperatorID(ctx context.Context, operatorID OperatorID) error

	// SoftDeleteByOrganization soft-deletes all memberships for an organization (used during org deletion).
	SoftDeleteByOrganization(ctx context.Context, orgID string) error

	// CountByOrganization counts members in an organization.
	CountByOrganization(ctx context.Context, orgID string) (int, error)

	// CountByOrganizationID counts members by OrganizationID.
	CountByOrganizationID(ctx context.Context, orgID OrganizationID) (int, error)

	// CountActiveByOrganization counts active (non-removed) members in an organization.
	CountActiveByOrganization(ctx context.Context, orgID string) (int, error)

	// CountActiveByOrganizationID counts active members by OrganizationID.
	CountActiveByOrganizationID(ctx context.Context, orgID OrganizationID) (int, error)

	// CountSuperAdminsByOrganization counts super_admin members in an organization.
	CountSuperAdminsByOrganization(ctx context.Context, orgID string) (int, error)

	// CountSuperAdminsByOrganizationID counts super_admin members by OrganizationID.
	CountSuperAdminsByOrganizationID(ctx context.Context, orgID OrganizationID) (int, error)

	// ListByOperator lists all memberships for an operator.
	ListByOperator(ctx context.Context, operatorID string) ([]*OrganizationMember, error)

	// ListByOperatorPaginated lists memberships for an operator with pagination.
	ListByOperatorPaginated(ctx context.Context, operatorID string, limit, offset int) ([]*OrganizationMember, int, error)

	// ListByOperatorID lists all memberships by OperatorID.
	ListByOperatorID(ctx context.Context, operatorID OperatorID) ([]*OrganizationMember, error)
}
