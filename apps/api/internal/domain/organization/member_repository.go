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

	// FindByOperatorAndOrg finds a membership by operator ID and organization ID.
	FindByOperatorAndOrg(ctx context.Context, operatorID, orgID string) (*OrganizationMember, error)

	// FindByOrganization lists all members of an organization.
	FindByOrganization(ctx context.Context, orgID string) ([]*OrganizationMember, error)

	// Update updates an existing membership.
	Update(ctx context.Context, member *OrganizationMember) error

	// SoftDelete soft-deletes a membership (marks as removed).
	SoftDelete(ctx context.Context, id string) error

	// CountByOrganization counts members in an organization.
	CountByOrganization(ctx context.Context, orgID string) (int, error)

	// CountActiveByOrganization counts active (non-removed) members in an organization.
	CountActiveByOrganization(ctx context.Context, orgID string) (int, error)

	// CountSuperAdminsByOrganization counts super_admin members in an organization.
	CountSuperAdminsByOrganization(ctx context.Context, orgID string) (int, error)

	// ListByOperator lists all memberships for an operator.
	ListByOperator(ctx context.Context, operatorID string) ([]*OrganizationMember, error)
}
