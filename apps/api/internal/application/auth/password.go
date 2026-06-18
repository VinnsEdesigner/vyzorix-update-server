package auth

import (
	"errors"
)

// PasswordHasher defines the interface for password hashing.
type PasswordHasher interface {
	// Hash creates a hash of the password.
	Hash(password string) (string, error)
	// Verify checks if the password matches the hash.
	Verify(password, hash string) error
}

var (
	ErrInvalidPassword = errors.New("invalid password")
)
