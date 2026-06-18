// Package storage provides SQLite database operations.
package storage

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// Argon2id parameters following OWASP 2023 recommendations.
// These values provide strong security against GPU/ASIC attacks.
var argon2Params = &argon2ParamsConfig{
	memory:      64 * 1024, // 64 MB
	iterations:  3,
	parallelism: 4,
	saltLength:  16,
	keyLength:   32,
}

type argon2ParamsConfig struct {
	memory      uint32
	iterations  uint32
	parallelism uint8
	saltLength  uint32
	keyLength   uint32
}

// ErrInvalidHash is returned when the hash format is invalid.
var ErrInvalidHash = errors.New("hash is not a valid argon2id hash")

// ErrHashMismatch is returned when the password doesn't match the hash.
var ErrHashMismatch = errors.New("password does not match hash")

// HashPassword creates an Argon2id hash of the password.
// The hash is deterministic for the same password - salt is prepended to output.
// Uses cryptographically secure random salt generation.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	salt, err := generateSecureSalt(argon2Params.saltLength)
	if err != nil {
		// crypto/rand failure is critical - return error rather than using insecure fallback
		return "", fmt.Errorf("failed to generate secure salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Params.iterations,
		argon2Params.memory,
		argon2Params.parallelism,
		argon2Params.keyLength,
	)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		argon2Params.memory,
		argon2Params.iterations,
		argon2Params.parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifyPassword checks if the password matches the argon2id hash.
// Uses constant-time comparison to prevent timing attacks.
func VerifyPassword(password, hash string) error {
	if password == "" || hash == "" {
		return ErrHashMismatch
	}

	salt, key, err := decodeArgon2Hash(hash)
	if err != nil {
		return err
	}

	expectedKey := argon2.IDKey(
		[]byte(password),
		salt,
		argon2Params.iterations,
		argon2Params.memory,
		argon2Params.parallelism,
		argon2Params.keyLength,
	)

	// Constant-time comparison
	if subtle.ConstantTimeCompare(key, expectedKey) != 1 {
		return ErrHashMismatch
	}

	return nil
}

// HashSecret creates an Argon2id hash of a secret (e.g., command secrets).
// Uses same parameters as password hashing.
func HashSecret(secret string) (string, error) {
	return HashPassword(secret)
}

// VerifySecret verifies a secret against its Argon2id hash.
func VerifySecret(secret, hash string) error {
	return VerifyPassword(secret, hash)
}

// decodeArgon2Hash parses an argon2id hash string and extracts salt and key.
func decodeArgon2Hash(hash string) ([]byte, []byte, error) {
	// Expected format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	// Example: $argon2id$v=19$m=65536,t=3,p=4$54kXwGAEJq6plTG/lD+5ow$frEN2YulgDEeCpKSHVXXlfjO6yNnZCjiQQ7pPZGAvuQ
	//
	// Structure (5 $ total at positions 0,9,14,30,53):
	//   $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	//   ^^^^^^^^   ^^^   ^^^^^^^^^^^^^   ^^^^^   ^^^^^
	//   dollar[0] dollar[1] dollar[2]    dollar[3] dollar[4]
	//
	// Salt is between dollar[3] and dollar[4]
	// Hash is after dollar[4]

	// Check prefix - must be "$argon2id$v" (11 chars)
	if len(hash) < 11 || hash[:11] != "$argon2id$v" {
		return nil, nil, ErrInvalidHash
	}

	// Find all $ positions
	var dollarPositions []int
	for i, c := range hash {
		if c == '$' {
			dollarPositions = append(dollarPositions, i)
		}
	}

	// We need at least 5 $ positions (format: $algo$v=x$m=y,t=z,p=n$<salt>$<hash>)
	if len(dollarPositions) < 5 {
		return nil, nil, ErrInvalidHash
	}

	// dollarPositions[3] = position of $ before salt
	// dollarPositions[4] = position of $ after salt (before hash)
	saltStart := dollarPositions[3] + 1
	saltEnd := dollarPositions[4] - 1
	hashStart := dollarPositions[4] + 1

	if saltEnd >= len(hash) || hashStart >= len(hash) {
		return nil, nil, ErrInvalidHash
	}

	// Extract salt and hash
	saltEncoded := hash[saltStart : saltEnd+1]
	hashEncoded := hash[hashStart:]

	salt, err := base64.RawStdEncoding.DecodeString(saltEncoded)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: invalid salt encoding: %v", ErrInvalidHash, err)
	}

	key, err := base64.RawStdEncoding.DecodeString(hashEncoded)
	if err != nil {
		return nil, nil, fmt.Errorf("%w: invalid hash encoding: %v", ErrInvalidHash, err)
	}

	return salt, key, nil
}

// generateSecureSalt generates cryptographically secure random bytes for salt.
// Uses crypto/rand which is guaranteed to be available in production environments.
func generateSecureSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)
	_, err := rand.Read(salt)
	if err != nil {
		// In production, crypto/rand NEVER fails. If it does, the system is in a critical
		// state and we should fail explicitly rather than fall back to insecure random.
		// The error is returned so the caller can handle it appropriately.
		return nil, err
	}
	return salt, nil
}