package organization

import "errors"

var (
	// ErrNotFound is returned when an organization is not found.
	ErrNotFound = errors.New("organization not found")

	// ErrMemberNotFound is returned when a member is not found.
	ErrMemberNotFound = errors.New("member not found")

	// ErrInvitationNotFound is returned when an invitation is not found.
	ErrInvitationNotFound = errors.New("invitation not found")

	// ErrOrganizationExists is returned when an organization with the same name already exists.
	ErrOrganizationExists = errors.New("organization with this name already exists")

	// ErrMemberExists is returned when a member already exists in the organization.
	ErrMemberExists = errors.New("member already exists in organization")

	// ErrInvitationExists is returned when an invitation already exists for this email.
	ErrInvitationExists = errors.New("invitation already exists for this email")

	// ErrMaxMembersReached is returned when the organization has reached its member limit.
	ErrMaxMembersReached = errors.New("organization has reached its member limit")

	// ErrCannotRemoveLastSuperAdmin is returned when trying to remove the last super_admin.
	ErrCannotRemoveLastSuperAdmin = errors.New("cannot remove the last super_admin")

	// ErrInvalidRole is returned when an invalid role is provided.
	ErrInvalidRole = errors.New("invalid role")

	// ErrForbidden is returned when the operator doesn't have permission.
	ErrForbidden = errors.New("forbidden")

	// ErrInvitationExpired is returned when an invitation has expired.
	ErrInvitationExpired = errors.New("invitation has expired")

	// ErrInvitationAlreadyResponded is returned when an invitation has already been responded to.
	ErrInvitationAlreadyResponded = errors.New("invitation has already been responded to")

	// ErrEmailMismatch is returned when the invitee email doesn't match.
	ErrEmailMismatch = errors.New("email does not match invitation")

	// ErrCannotInviteSuperAdmin is returned when trying to invite a super_admin.
	ErrCannotInviteSuperAdmin = errors.New("cannot invite super_admin role")

	// ErrOperatorNotInOrganization is returned when operator is not a member of the organization.
	ErrOperatorNotInOrganization = errors.New("operator is not a member of this organization")

	// ErrAlreadyExists is a generic already exists error.
	ErrAlreadyExists = errors.New("resource already exists")

	// ErrAlreadyResponded is returned when an invitation has already been responded to.
	ErrAlreadyResponded = errors.New("invitation has already been responded to")

	// ErrExpired is returned when an invitation has expired.
	ErrExpired = errors.New("resource has expired")

	// ErrCannotDeleteLastSuperAdmin is returned when trying to delete the last super_admin of an org.
	ErrCannotDeleteLastSuperAdmin = errors.New("cannot delete the last super_admin of an organization")

	// ErrTransferOwnershipNotAllowed is returned when transfer ownership is not allowed.
	ErrTransferOwnershipNotAllowed = errors.New("transfer ownership not allowed")

	// ErrDeviceTransferNotAllowed is returned when device transfer is not allowed.
	ErrDeviceTransferNotAllowed = errors.New("device transfer not allowed")

	// ErrDeviceOnline is returned when a device is still online.
	ErrDeviceOnline = errors.New("device must be offline to transfer")
)
