package organization

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"
)

// InvitationStatus represents the status of an invitation.
type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusApproved InvitationStatus = "approved"
	InvitationStatusRejected InvitationStatus = "rejected"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// Invitation represents an invitation to join an organization.
type Invitation struct {
	// Lifecycle tracks the invitation lifecycle state
	Lifecycle InvitationLifecycle

	// Core fields
	ID             string
	OrganizationID string
	Email          string
	Role           OrganizationRole
	Status         InvitationStatus
	Token          string

	// Notes
	InviterNotes string // Notes from inviter
	InviteeNotes string // Notes from invitee on accept/reject

	// References
	InvitedBy   string // OperatorID of inviter
	RespondedBy string // OperatorID of responder (when responded)

	// Timestamps
	InvitedAt  time.Time
	RespondedAt *time.Time // When invitation was responded to
	ExpiresAt  time.Time  // When invitation expires

	// Populated fields
	OrganizationName string
	InviterName     string
	InviterEmail    string
}

// InvitationLifecycle represents the lifecycle state of an invitation.
type InvitationLifecycle string

const (
	// InvitationLifecyclePending represents a pending invitation.
	InvitationLifecyclePending InvitationLifecycle = "pending"
	// InvitationLifecycleAccepted represents an accepted invitation.
	InvitationLifecycleAccepted InvitationLifecycle = "accepted"
	// InvitationLifecycleRejected represents a rejected invitation.
	InvitationLifecycleRejected InvitationLifecycle = "rejected"
	// InvitationLifecycleExpired represents an expired invitation.
	InvitationLifecycleExpired InvitationLifecycle = "expired"
)

// ErrInvalidInvitationTransition is returned when an invitation lifecycle transition is not allowed.
var ErrInvalidInvitationTransition = errors.New("invalid invitation lifecycle transition")

// InvitationLifecycleTransitions defines valid state transitions for invitation lifecycle.
var InvitationLifecycleTransitions = map[InvitationLifecycle]map[InvitationLifecycle]bool{
	InvitationLifecyclePending: {
		InvitationLifecycleAccepted: true,
		InvitationLifecycleRejected: true,
		InvitationLifecycleExpired: true,
	},
	InvitationLifecycleAccepted: {}, // No transitions allowed from accepted
	InvitationLifecycleRejected: {}, // No transitions allowed from rejected
	InvitationLifecycleExpired: {}, // No transitions allowed from expired
}

// CanTransitionTo returns true if the lifecycle can transition to the target state.
func (l InvitationLifecycle) CanTransitionTo(target InvitationLifecycle) bool {
	allowed, exists := InvitationLifecycleTransitions[l]
	if !exists {
		return false
	}
	return allowed[target]
}

// TransitionTo transitions the lifecycle to a new state.
// Returns ErrInvalidInvitationTransition if the transition is not allowed.
func (l *InvitationLifecycle) TransitionTo(target InvitationLifecycle) error {
	if !l.CanTransitionTo(target) {
		return fmt.Errorf("%w: cannot transition from %s to %s", ErrInvalidInvitationTransition, *l, target)
	}
	*l = target
	return nil
}

// IsPending returns true if the invitation is pending.
func (l InvitationLifecycle) IsPending() bool {
	return l == InvitationLifecyclePending
}

// IsAccepted returns true if the invitation was accepted.
func (l InvitationLifecycle) IsAccepted() bool {
	return l == InvitationLifecycleAccepted
}

// IsRejected returns true if the invitation was rejected.
func (l InvitationLifecycle) IsRejected() bool {
	return l == InvitationLifecycleRejected
}

// IsExpired returns true if the invitation has expired.
func (l InvitationLifecycle) IsExpired() bool {
	return l == InvitationLifecycleExpired
}

// GenerateToken generates a secure random token for the invitation.
func GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// NewInvitation creates a new Invitation with pending lifecycle.
func NewInvitation(id, organizationID, email string, role OrganizationRole, invitedBy string) (*Invitation, error) {
	token, err := GenerateToken()
	if err != nil {
		return nil, err
	}

	// Default expiry: 7 days
	expiresAt := time.Now().Add(7 * 24 * time.Hour)

	return &Invitation{
		ID:             id,
		OrganizationID: organizationID,
		Email:          email,
		Role:           role,
		Status:         InvitationStatusPending,
		Token:          token,
		InvitedBy:      invitedBy,
		InvitedAt:      time.Now(),
		ExpiresAt:      expiresAt,
		Lifecycle:      InvitationLifecyclePending,
	}, nil
}

// Accept transitions the invitation to accepted state.
func (i *Invitation) Accept(respondedBy string, notes string) error {
	if err := i.Lifecycle.TransitionTo(InvitationLifecycleAccepted); err != nil {
		return err
	}
	now := time.Now()
	i.RespondedAt = &now
	i.RespondedBy = respondedBy
	i.InviteeNotes = notes
	i.Status = InvitationStatusApproved
	return nil
}

// Reject transitions the invitation to rejected state.
func (i *Invitation) Reject(respondedBy string, notes string) error {
	if err := i.Lifecycle.TransitionTo(InvitationLifecycleRejected); err != nil {
		return err
	}
	now := time.Now()
	i.RespondedAt = &now
	i.RespondedBy = respondedBy
	i.InviteeNotes = notes
	i.Status = InvitationStatusRejected
	return nil
}

// Expire transitions the invitation to expired state.
func (i *Invitation) Expire() error {
	if err := i.Lifecycle.TransitionTo(InvitationLifecycleExpired); err != nil {
		return err
	}
	i.Status = InvitationStatusExpired
	return nil
}

// IsPending returns true if the invitation is pending.
func (i *Invitation) IsPending() bool {
	return i.Lifecycle.IsPending()
}

// IsAccepted returns true if the invitation was accepted.
func (i *Invitation) IsAccepted() bool {
	return i.Lifecycle.IsAccepted()
}

// IsRejected returns true if the invitation was rejected.
func (i *Invitation) IsRejected() bool {
	return i.Lifecycle.IsRejected()
}

// IsExpired returns true if the invitation has expired.
func (i *Invitation) IsExpired() bool {
	return i.Lifecycle.IsExpired() || time.Now().After(i.ExpiresAt)
}

// IsValid returns true if the invitation has all required fields.
func (i *Invitation) IsValid() bool {
	return i.ID != "" && i.OrganizationID != "" && i.Email != "" && i.Token != ""
}

// HasResponded returns true if the invitation has been responded to.
func (i *Invitation) HasResponded() bool {
	return i.RespondedAt != nil
}

// CreateInvitationRequest represents a request to create an invitation.
type CreateInvitationRequest struct {
	OrganizationID string
	Email          string
	Role           OrganizationRole
	InviterNotes   string
}

// Validate validates the create invitation request.
func (r *CreateInvitationRequest) Validate() error {
	if r.OrganizationID == "" {
		return errors.New("organization ID is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if r.Role == "" {
		return errors.New("role is required")
	}
	// Cannot invite super_admin role
	if r.Role == RoleSuperAdmin {
		return ErrCannotInviteSuperAdmin
	}
	return nil
}

// RespondToInvitationRequest represents a request to respond to an invitation.
type RespondToInvitationRequest struct {
	Token string
	Email string
	Notes string
}

// Validate validates the respond to invitation request.
func (r *RespondToInvitationRequest) Validate() error {
	if r.Token == "" {
		return errors.New("token is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	return nil
}


