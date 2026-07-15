package organization

import (
	"context"
)

// InvitationRepository defines the interface for invitation data access.
type InvitationRepository interface {
	// Create creates a new invitation.
	Create(ctx context.Context, invitation *Invitation) error

	// FindByID retrieves an invitation by ID.
	FindByID(ctx context.Context, id string) (*Invitation, error)

	// FindByToken retrieves an invitation by token.
	FindByToken(ctx context.Context, token string) (*Invitation, error)

	// FindByEmail retrieves all invitations for an email across all organizations.
	FindByEmail(ctx context.Context, email string) ([]*Invitation, error)

	// FindPendingByEmail retrieves all pending invitations for an email.
	FindPendingByEmail(ctx context.Context, email string) ([]*Invitation, error)

	// FindByOrganization retrieves all invitations for an organization.
	FindByOrganization(ctx context.Context, orgID string) ([]*Invitation, error)

	// FindPendingByOrganization retrieves all pending invitations for an organization.
	FindPendingByOrganization(ctx context.Context, orgID string) ([]*Invitation, error)

	// FindByOrganizationAndEmail retrieves invitations by organization and email.
	FindByOrganizationAndEmail(ctx context.Context, orgID, email string) ([]*Invitation, error)

	// FindPendingByOrganizationAndEmail retrieves pending invitations by org and email.
	FindPendingByOrganizationAndEmail(ctx context.Context, orgID, email string) ([]*Invitation, error)

	// Update updates an existing invitation.
	Update(ctx context.Context, invitation *Invitation) error

	// SoftDelete soft-deletes an invitation.
	SoftDelete(ctx context.Context, id string) error

	// SoftDeleteByToken soft-deletes an invitation by token.
	SoftDeleteByToken(ctx context.Context, token string) error

	// SoftDeleteExpired soft-deletes all expired invitations.
	SoftDeleteExpired(ctx context.Context) error

	// SoftDeleteByOrganization soft-deletes all invitations for an organization.
	SoftDeleteByOrganization(ctx context.Context, orgID string) error

	// SoftDeleteByInvitedBy soft-deletes all invitations sent by an operator.
	SoftDeleteByInvitedBy(ctx context.Context, operatorID string) error

	// CountByOrganization counts invitations for an organization.
	CountByOrganization(ctx context.Context, orgID string) (int, error)

	// CountPendingByOrganization counts pending invitations for an organization.
	CountPendingByOrganization(ctx context.Context, orgID string) (int, error)

	// ExpireOldThan soft-deletes invitations older than the given duration.
	ExpireOldThan(ctx context.Context, duration string) error
}
