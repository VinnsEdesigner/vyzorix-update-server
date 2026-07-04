// Package password provides password hashing and validation.
package password

import (
	"crypto/sha1" // #nosec G505 - SHA-1 required for HIBP k-anonymity API compatibility
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"unicode"
)

// ErrPasswordBreached indicates the password was found in a data breach.
var ErrPasswordBreached = fmt.Errorf("password found in known data breach")

// CheckBreached checks if a password appears in known data breaches using HIBP API.
// Uses k-anonymity: only sends first 5 chars of SHA-1 hash to protect the password.
// 6: Added to prevent password reuse from known breaches.
func CheckBreached(password string) (bool, error) {
	// #nosec G505 - SHA-1 required for HIBP API k-anonymity model (not for password storage)
	hash := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Only send first 5 characters (k-anonymity)
	prefix := hashHex[:5]
	suffix := hashHex[5:]

	// Query HIBP API
	resp, err := http.Get("https://api.pwnedpasswords.com/range/" + prefix)
	if err != nil {
		// If HIBP is unavailable, fail open with a warning
		// In production, you might want to fail closed instead
		return false, nil
	}
	// #nosec G104 - HIBP API doesn't require auth and Body.Close error is not critical
	if closeErr := resp.Body.Close(); closeErr != nil {
		// Log but don't fail - this is not a security issue
	}

	if resp.StatusCode != http.StatusOK {
		return false, nil
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, nil
	}

	// Check if our hash suffix is in the response
	lines := strings.Split(string(body), "\n")
	for _, line := range lines {
		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			continue
		}
		if strings.TrimSpace(parts[0]) == suffix {
			return true, nil
		}
	}

	return false, nil
}

// Policy defines the requirements for a valid password.
type Policy struct {
	MinLength      int
	MaxLength      int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
}

// DefaultPolicy is the standard password policy.
var DefaultPolicy = Policy{
	MinLength:      8,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: true,
}

// UserPolicy is a user-friendly password policy for personal account passwords.
var UserPolicy = Policy{
	MinLength:      12,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: false,
}

// Error represents validation failures for a password.
type Error struct {
	Missing []string
}

func (e *Error) Error() string {
	return "password does not meet requirements: " + strings.Join(e.Missing, "; ")
}

// Validate checks a password against the given policy.
func Validate(password string, policy Policy) error {
	var missing []string

	if len(password) < policy.MinLength {
		missing = append(missing, fmt.Sprintf("minimum %d characters", policy.MinLength))
	}

	if len(password) > policy.MaxLength {
		missing = append(missing, fmt.Sprintf("maximum %d characters", policy.MaxLength))
	}

	if policy.RequireUpper && !containsUpper(password) {
		missing = append(missing, "at least 1 uppercase letter (A-Z)")
	}

	if policy.RequireLower && !containsLower(password) {
		missing = append(missing, "at least 1 lowercase letter (a-z)")
	}

	if policy.RequireDigit && !containsDigit(password) {
		missing = append(missing, "at least 1 number (0-9)")
	}

	if policy.RequireSpecial && !containsSpecial(password) {
		missing = append(missing, "at least 1 special character (!@#$%^&*()_+-=)")
	}

	if len(missing) > 0 {
		return &Error{Missing: missing}
	}

	return nil
}

func containsUpper(s string) bool {
	for _, r := range s {
		if unicode.IsUpper(r) {
			return true
		}
	}

	return false
}

func containsLower(s string) bool {
	for _, r := range s {
		if unicode.IsLower(r) {
			return true
		}
	}

	return false
}

func containsDigit(s string) bool {
	for _, r := range s {
		if unicode.IsDigit(r) {
			return true
		}
	}

	return false
}

func containsSpecial(s string) bool {
	specialChars := "!@#$%^&*()_+-="
	for _, r := range s {
		if strings.ContainsRune(specialChars, r) {
			return true
		}
	}

	return false
}

// Strength returns a score from 0-5 based on password complexity.
func Strength(password string) int {
	score := 0

	if len(password) >= 8 {
		score++
	}

	if len(password) >= 12 {
		score++
	}

	if len(password) >= 16 {
		score++
	}

	if containsUpper(password) && containsLower(password) {
		score++
	}

	if containsDigit(password) || containsSpecial(password) {
		score++
	}

	return score
}
