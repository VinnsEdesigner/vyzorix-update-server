package invitation

import (
	"context"
)

// InvitationFilter represents filters for listing invitations.
type InvitationFilter struct {
	Status *InvitationStatus
}

// Repository defines the interface for invitation data access.
type Repository interface {
	// Create creates a new invitation.
	Create(ctx context.Context, invite *Invitation) error

	// FindByID retrieves an invitation by ID.
	FindByID(ctx context.Context, id string) (*Invitation, error)

	// FindByToken retrieves an invitation by its secure token.
	FindByToken(ctx context.Context, token string) (*Invitation, error)

	// FindPendingByEmail finds all pending invitations for an email.
	FindPendingByEmail(ctx context.Context, email string) ([]*Invitation, error)

	// FindPendingByEmailAndOrg finds a pending invitation for an email and organization.
	FindPendingByEmailAndOrg(ctx context.Context, email, orgID string) (*Invitation, error)

	// Update updates an invitation.
	Update(ctx context.Context, invite *Invitation) error

	// ListByOrganization lists all invitations for an organization.
	ListByOrganization(ctx context.Context, orgID string, filter *InvitationFilter) ([]*Invitation, error)

	// ListByInviter lists all invitations sent by an operator.
	ListByInviter(ctx context.Context, inviterID string) ([]*Invitation, error)

	// Delete deletes an invitation.
	Delete(ctx context.Context, id string) error

	// ExpireByOrganization expires all pending invitations for an organization.
	ExpireByOrganization(ctx context.Context, orgID string) error
}
