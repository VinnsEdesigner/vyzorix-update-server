package invitation

import "errors"

var (
	// ErrNotFound is returned when an invitation is not found.
	ErrNotFound = errors.New("invitation not found")

	// ErrTokenInvalid is returned when the invitation token is invalid.
	ErrTokenInvalid = errors.New("invalid invitation token")

	// ErrExpired is returned when the invitation has expired.
	ErrExpired = errors.New("invitation has expired")

	// ErrAlreadyExists is returned when an invitation already exists for this email and organization.
	ErrAlreadyExists = errors.New("invitation already exists for this email and organization")

	// ErrAlreadyResponded is returned when the invitation has already been responded to.
	ErrAlreadyResponded = errors.New("invitation has already been responded to")

	// ErrEmailMismatch is returned when the invitee email doesn't match the invitation.
	ErrEmailMismatch = errors.New("email does not match invitation")

	// ErrForbidden is returned when the operator doesn't have permission.
	ErrForbidden = errors.New("forbidden")

	// ErrInvalidRole is returned when an invalid role is provided.
	ErrInvalidRole = errors.New("invalid role")
)
