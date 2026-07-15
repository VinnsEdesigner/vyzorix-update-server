package application

import "errors"

var (
	// Auth errors.
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrAccountLocked     = errors.New("account locked")
	ErrMFARequired      = errors.New("mfa required")
	ErrInvalidMFACode   = errors.New("invalid mfa code")
	ErrTokenExpired     = errors.New("token expired")
	ErrTokenUsed       = errors.New("token already used")
	ErrEmailNotVerified = errors.New("email not verified")
	ErrEmailAlreadyVerified = errors.New("email already verified")
	ErrEmailExists      = errors.New("email already exists")
	ErrUserExists      = errors.New("user already exists")
	ErrPasswordBreached = errors.New("password found in data breach, please choose another")
	ErrPasswordPolicy = errors.New("password policy violation")
	ErrOAuthFailed     = errors.New("oauth authentication failed")
	ErrOperatorNotFound = errors.New("not_found")

	// Organization errors (multi-tenant).
	ErrNoOrganization        = errors.New("no organization found, please create or join one")
	ErrOrgSelectionRequired  = errors.New("organization selection required")
	ErrInvalidOrganization  = errors.New("invalid organization")
	ErrOrganizationInactive  = errors.New("organization is inactive")
	ErrMembershipInactive    = errors.New("membership is not active")
	ErrNotOrgMember          = errors.New("not a member of this organization")
	ErrCannotDeleteLastSuperAdmin = errors.New("cannot delete the last super_admin of an organization")

	// Device errors.
	ErrDeviceNotFound   = errors.New("device not found")
	ErrDeviceHijack     = errors.New("device registration conflict")
	ErrDeviceExists     = errors.New("device already exists")
	ErrCommandNotFound  = errors.New("command not found")
	ErrCommandFailed    = errors.New("command failed")
	ErrDeviceOnline     = errors.New("device must be offline to transfer")

	// General errors.
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrRateLimited  = errors.New("rate limited")
	ErrForbiddenIP  = errors.New("forbidden IP address")
	ErrClientNotFound = errors.New("client not found")
)
	ErrEmailExists      = errors.New("email already exists")
	ErrUserExists      = errors.New("user already exists")
	ErrPasswordBreached = errors.New("password found in data breach, please choose another")
	ErrPasswordPolicy = errors.New("password policy violation")
	ErrOAuthFailed     = errors.New("oauth authentication failed")
	ErrOperatorNotFound = errors.New("not_found")

	// Device errors.
	ErrDeviceNotFound   = errors.New("device not found")
	ErrDeviceHijack     = errors.New("device registration conflict")
	ErrDeviceExists     = errors.New("device already exists")
	ErrCommandNotFound  = errors.New("command not found")
	ErrCommandFailed    = errors.New("command failed")

	// General errors.
	ErrInvalidInput  = errors.New("invalid input")
	ErrInternal     = errors.New("internal error")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden    = errors.New("forbidden")
	ErrRateLimited  = errors.New("rate limited")
	ErrForbiddenIP  = errors.New("forbidden IP address")
	ErrClientNotFound = errors.New("client not found")
)
