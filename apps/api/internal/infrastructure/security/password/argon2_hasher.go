package password

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"fmt"

	"golang.org/x/crypto/argon2"
)

// ErrHashFailed is returned when hashing fails.
var ErrHashFailed = errors.New("hash failed")

// ErrInvalidHash = errors.New("invalid hash")

// Argon2idParams holds Argon2id configuration following OWASP 2023 recommendations.
type Argon2idParams struct {
	Memory      uint32
	Iterations  uint32
	Parallelism uint8
	SaltLength  uint32
	KeyLength   uint32
}

// DefaultArgon2idParams returns the default Argon2id parameters.
var DefaultArgon2idParams = Argon2idParams{
	Memory:      64 * 1024, // 64 MB
	Iterations:  3,
	Parallelism: 4,
	SaltLength:  16,
	KeyLength:   32,
}

// Argon2idHasher implements password hashing using Argon2id.
type Argon2idHasher struct {
	params Argon2idParams
}

// NewArgon2idHasher creates a new Argon2idHasher with default parameters.
func NewArgon2idHasher() *Argon2idHasher {
	return &Argon2idHasher{params: DefaultArgon2idParams}
}

// Hash creates an Argon2id hash of the password.
func (h *Argon2idHasher) Hash(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	salt, err := generateSecureSalt(h.params.SaltLength)
	if err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		h.params.Memory,
		h.params.Iterations,
		h.params.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// Verify checks if the password matches the Argon2id hash using constant-time comparison.
func (h *Argon2idHasher) Verify(password, hash string) error {
	if password == "" || hash == "" {
		return fmt.Errorf("password or hash cannot be empty")
	}

	salt, expectedKey, err := decodeArgon2Hash(hash)
	if err != nil {
		return fmt.Errorf("invalid hash format: %w", err)
	}

	computedKey := argon2.IDKey(
		[]byte(password),
		salt,
		h.params.Iterations,
		h.params.Memory,
		h.params.Parallelism,
		h.params.KeyLength,
	)

	if subtle.ConstantTimeCompare(computedKey, expectedKey) != 1 {
		return fmt.Errorf("password mismatch")
	}

	return nil
}

// generateSecureSalt generates cryptographically secure random bytes for salt.
func generateSecureSalt(length uint32) ([]byte, error) {
	salt := make([]byte, length)

	_, err := rand.Read(salt)
	if err != nil {
		return nil, fmt.Errorf("failed to generate secure salt: %w", err)
	}

	return salt, nil
}

// decodeArgon2Hash parses an argon2id hash string and extracts salt and key.
func decodeArgon2Hash(hash string) ([]byte, []byte, error) {
	if len(hash) < 11 || hash[:11] != "$argon2id$v" {
		return nil, nil, fmt.Errorf("invalid hash prefix")
	}

	var dollarPositions []int

	for i, c := range hash {
		if c == '$' {
			dollarPositions = append(dollarPositions, i)
		}
	}

	if len(dollarPositions) < 5 {
		return nil, nil, fmt.Errorf("invalid hash format: missing $ delimiters")
	}

	saltStart := dollarPositions[3] + 1
	saltEnd := dollarPositions[4] - 1
	hashStart := dollarPositions[4] + 1

	if saltEnd >= len(hash) || hashStart >= len(hash) {
		return nil, nil, fmt.Errorf("invalid hash format: out of bounds")
	}

	salt, err := base64.RawStdEncoding.DecodeString(hash[saltStart : saltEnd+1])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid salt encoding: %w", err)
	}

	key, err := base64.RawStdEncoding.DecodeString(hash[hashStart:])
	if err != nil {
		return nil, nil, fmt.Errorf("invalid hash encoding: %w", err)
	}

	return salt, key, nil
}

// HashSecret creates an Argon2id hash of a secret for secure storage.
// Uses OWASP 2023 recommended parameters for strong protection.
func HashSecret(secret string) (string, error) {
	if secret == "" {
		return "", fmt.Errorf("secret cannot be empty")
	}

	salt := make([]byte, 16)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(secret),
		salt,
		DefaultArgon2idParams.Iterations,
		DefaultArgon2idParams.Memory,
		DefaultArgon2idParams.Parallelism,
		DefaultArgon2idParams.KeyLength,
	)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		DefaultArgon2idParams.Memory,
		DefaultArgon2idParams.Iterations,
		DefaultArgon2idParams.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifySecret checks if a secret matches its Argon2id hash using constant-time comparison.
func VerifySecret(secret, hash string) error {
	if secret == "" || hash == "" {
		return fmt.Errorf("secret or hash cannot be empty")
	}

	// Check if it's the legacy SHA512-based hash format (no $argon2id$ prefix)
	if len(hash) > 0 && hash[0] != '$' {
		return errors.New("invalid hash format")
	}

	salt, expectedKey, err := decodeArgon2Hash(hash)
	if err != nil {
		return fmt.Errorf("invalid hash format: %w", err)
	}

	computedKey := argon2.IDKey(
		[]byte(secret),
		salt,
		DefaultArgon2idParams.Iterations,
		DefaultArgon2idParams.Memory,
		DefaultArgon2idParams.Parallelism,
		DefaultArgon2idParams.KeyLength,
	)

	if subtle.ConstantTimeCompare(computedKey, expectedKey) != 1 {
		return fmt.Errorf("secret mismatch")
	}

	return nil
}

// HashPassword creates an Argon2id hash of a password for operator login.
// Uses OWASP 2023 recommended parameters for strong protection against GPU/ASIC attacks.
func HashPassword(password string) (string, error) {
	if password == "" {
		return "", fmt.Errorf("password cannot be empty")
	}

	salt := make([]byte, DefaultArgon2idParams.SaltLength)
	if _, err := rand.Read(salt); err != nil {
		return "", fmt.Errorf("failed to generate salt: %w", err)
	}

	hash := argon2.IDKey(
		[]byte(password),
		salt,
		DefaultArgon2idParams.Iterations,
		DefaultArgon2idParams.Memory,
		DefaultArgon2idParams.Parallelism,
		DefaultArgon2idParams.KeyLength,
	)

	// Format: $argon2id$v=19$m=65536,t=3,p=4$<salt>$<hash>
	encodedSalt := base64.RawStdEncoding.EncodeToString(salt)
	encodedHash := base64.RawStdEncoding.EncodeToString(hash)

	return fmt.Sprintf("$argon2id$v=%d$m=%d,t=%d,p=%d$%s$%s",
		argon2.Version,
		DefaultArgon2idParams.Memory,
		DefaultArgon2idParams.Iterations,
		DefaultArgon2idParams.Parallelism,
		encodedSalt,
		encodedHash,
	), nil
}

// VerifyPassword checks if a password matches its Argon2id hash using constant-time comparison.
func VerifyPassword(password, hash string) error {
	if password == "" || hash == "" {
		return fmt.Errorf("password or hash cannot be empty")
	}

	salt, expectedKey, err := decodeArgon2Hash(hash)
	if err != nil {
		return fmt.Errorf("invalid hash format: %w", err)
	}

	computedKey := argon2.IDKey(
		[]byte(password),
		salt,
		DefaultArgon2idParams.Iterations,
		DefaultArgon2idParams.Memory,
		DefaultArgon2idParams.Parallelism,
		DefaultArgon2idParams.KeyLength,
	)

	if subtle.ConstantTimeCompare(computedKey, expectedKey) != 1 {
		return fmt.Errorf("password mismatch")
	}

	return nil
}
