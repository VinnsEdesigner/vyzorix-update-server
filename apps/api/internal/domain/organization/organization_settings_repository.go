package organization

import "context"

// OrganizationSettingsRepository defines the interface for organization settings data access.
type OrganizationSettingsRepository interface {
	// Create creates new organization settings.
	Create(ctx context.Context, settings *OrganizationSettings) error

	// FindByID retrieves settings by ID.
	FindByID(ctx context.Context, id string) (*OrganizationSettings, error)

	// FindByOrganizationID retrieves settings by organization ID.
	FindByOrganizationID(ctx context.Context, orgID string) (*OrganizationSettings, error)

	// Update updates organization settings.
	Update(ctx context.Context, settings *OrganizationSettings) error

	// Delete deletes organization settings.
	Delete(ctx context.Context, id string) error

	// DeleteByOrganizationID deletes settings by organization ID.
	DeleteByOrganizationID(ctx context.Context, orgID string) error
}
