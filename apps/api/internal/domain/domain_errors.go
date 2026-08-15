package domain

import "errors"

var (
	// ErrNotFound is returned when an entity is not found.
	ErrNotFound = errors.New("entity not found")

	// ErrAlreadyExists is returned when trying to create a duplicate entity.
	ErrAlreadyExists = errors.New("entity already exists")

	// ErrInvalidInput is returned when input validation fails.
	ErrInvalidInput = errors.New("invalid input")

	// ErrUnauthorized is returned when authentication fails.
	ErrUnauthorized = errors.New("unauthorized")

	// ErrForbidden is returned when the user doesn't have permission.
	ErrForbidden = errors.New("forbidden")

	// ErrInternal is returned for internal errors.
	ErrInternal = errors.New("internal error")

	// ErrNotImplemented is returned when a method is not implemented.
	ErrNotImplemented = errors.New("not implemented")

	// ErrExpired is returned when a token or session has expired.
	ErrExpired = errors.New("expired")

	// ErrInvalidToken is returned when a token is invalid.
	ErrInvalidToken = errors.New("invalid token")
)
