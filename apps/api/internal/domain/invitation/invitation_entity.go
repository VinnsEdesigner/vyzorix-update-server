package invitation

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"time"
)

// InvitationStatus represents the status of an invitation.
type InvitationStatus string

const (
	InvitationStatusPending   InvitationStatus = "pending"
	InvitationStatusApproved InvitationStatus = "approved"
	InvitationStatusRejected InvitationStatus = "rejected"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// InvitationStatusPtr returns a pointer to the given InvitationStatus.
func InvitationStatusPtr(s InvitationStatus) *InvitationStatus {
	return &s
}

// Invitation represents an invitation to join an organization.
type Invitation struct {
	ID             string
	OrganizationID string
	Email         string
	Role          InvitationRole
	Status        InvitationStatus
	Token         string
	InviterNotes  *string
	InviteeNotes  *string
	InvitedBy     string
	InvitedAt     time.Time
	RespondedAt   *time.Time
	ResponderID   *string
	ExpiresAt     time.Time

	// Populated fields (joined from other tables)
	OrganizationName string
	InviterName     string
}

// InvitationRole represents the role that will be assigned when an invitation is accepted.
type InvitationRole string

const (
	InvitationRoleAdmin    InvitationRole = "admin"
	InvitationRoleOperator InvitationRole = "operator"
	InvitationRoleViewer  InvitationRole = "viewer"
)

// IsValid returns true if the invitation role is valid.
func (r InvitationRole) IsValid() bool {
	switch r {
	case InvitationRoleAdmin, InvitationRoleOperator, InvitationRoleViewer:
		return true
	default:
		return false
	}
}

// ToOrgRole converts an invitation role to an organization role.
func (r InvitationRole) ToOrgRole() string {
	switch r {
	case InvitationRoleAdmin:
		return "admin"
	case InvitationRoleOperator:
		return "operator"
	case InvitationRoleViewer:
		return "viewer"
	default:
		return "viewer"
	}
}

// IsExpired returns true if the invitation has expired.
func (i *Invitation) IsExpired() bool {
	return time.Now().After(i.ExpiresAt)
}

// IsPending returns true if the invitation is pending.
func (i *Invitation) IsPending() bool {
	return i.Status == InvitationStatusPending
}

// CanBeAccepted returns true if the invitation can be accepted.
func (i *Invitation) CanBeAccepted() bool {
	return i.Status == InvitationStatusPending && !i.IsExpired()
}

// CreateInvitationRequest represents a request to create an invitation.
type CreateInvitationRequest struct {
	OrganizationID string
	Email         string
	Role          InvitationRole
	InviterNotes  string
}

// Validate validates the create invitation request.
func (r *CreateInvitationRequest) Validate() error {
	if r.OrganizationID == "" {
		return errors.New("organization ID is required")
	}
	if r.Email == "" {
		return errors.New("email is required")
	}
	if !r.Role.IsValid() {
		return errors.New("invalid role")
	}
	return nil
}

// AcceptInvitationRequest represents a request to accept an invitation.
type AcceptInvitationRequest struct {
	Token       string
	InviteeNotes string
}

// RespondInvitationRequest represents a request to respond to an invitation.
type RespondInvitationRequest struct {
	Token       string
	Notes       string
}


// GenerateSecureToken generates a secure random token for invitation links.
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// ConstantTimeCompare performs a constant-time comparison of two strings.
func ConstantTimeCompare(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}
