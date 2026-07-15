package organization

import (
	"context"
)

// OrganizationRepository defines the interface for organization data access.
type OrganizationRepository interface {
	// Create creates a new organization.
	Create(ctx context.Context, org *Organization) error

	// FindByID retrieves an organization by ID.
	FindByID(ctx context.Context, id string) (*Organization, error)

	// FindByOrganizationID retrieves an organization by OrganizationID value object.
	FindByOrganizationID(ctx context.Context, id OrganizationID) (*Organization, error)

	// FindByName finds an organization by name for a specific operator.
	FindByName(ctx context.Context, operatorID, name string) (*Organization, error)

	// FindByNameAndOperatorID finds an organization by name for a specific operator.
	FindByNameAndOperatorID(ctx context.Context, operatorID OperatorID, name string) (*Organization, error)

	// Update updates an existing organization.
	Update(ctx context.Context, org *Organization) error

	// SoftDelete soft-deletes an organization.
	SoftDelete(ctx context.Context, id string) error

	// SoftDeleteByID soft-deletes an organization by OrganizationID.
	SoftDeleteByID(ctx context.Context, id OrganizationID) error

	// ListByOperator lists all organizations for an operator (via memberships).
	ListByOperator(ctx context.Context, operatorID string) ([]*Organization, error)

	// ListByOperatorID lists all organizations for an OperatorID.
	ListByOperatorID(ctx context.Context, operatorID OperatorID) ([]*Organization, error)

	// ListActive lists all active (non-archived) organizations.
	ListActive(ctx context.Context) ([]*Organization, error)

	// CountActiveMembers counts the number of active members in an organization.
	CountActiveMembers(ctx context.Context, orgID string) (int, error)

	// CountActiveMembersByID counts the number of active members by OrganizationID.
	CountActiveMembersByID(ctx context.Context, orgID OrganizationID) (int, error)
}
