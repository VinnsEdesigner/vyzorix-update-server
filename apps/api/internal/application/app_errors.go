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
