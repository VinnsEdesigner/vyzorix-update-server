// Package password provides password hashing and validation.
package password

import (
	"crypto/sha1" // #nosec G505 - SHA-1 required for HIBP k-anonymity API compatibility
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"
)

// ErrPasswordBreached indicates the password was found in a data breach.
var ErrPasswordBreached = fmt.Errorf("password found in known data breach")

// CheckBreached checks if a password appears in known data breaches using HIBP API.
// Uses k-anonymity: only sends first 5 chars of SHA-1 hash to protect the password.
func CheckBreached(password string) (bool, error) {
	if len(password) < 1 {
		return false, fmt.Errorf("password cannot be empty")
	}

	// #nosec G505 - SHA-1 required for HIBP API k-anonymity model (not for password storage)
	hash := sha1.Sum([]byte(password))
	hashHex := strings.ToUpper(hex.EncodeToString(hash[:]))

	// Only send first 5 characters (k-anonymity)
	prefix := hashHex[:5]
	suffix := hashHex[5:]

	// Create HTTP client with timeout for security
	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Query HIBP API
	resp, err := client.Get("https://api.pwnedpasswords.com/range/" + prefix)
	if err != nil {
		// Fail closed: if we cannot verify, treat as potentially breached for security
		return true, fmt.Errorf("unable to verify password against breach database: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		// Fail closed: any non-200 response means we couldn't verify
		return true, fmt.Errorf("HIBP API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1024*1024)) // Limit response to 1MB
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
	MaxLength     int
	MaxConsecutive int
	RequireUpper   bool
	RequireLower   bool
	RequireDigit   bool
	RequireSpecial bool
	DisallowCommon bool
}

// DefaultPolicy is the standard password policy for operators.
var DefaultPolicy = Policy{
	MinLength:      12, // Increased from 8
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: true,
	MaxConsecutive: 3, // Prevent "aaaaaa" type passwords
	DisallowCommon: true,
}

// UserPolicy is a user-friendly password policy for personal account passwords.
var UserPolicy = Policy{
	MinLength:      12,
	MaxLength:      128,
	RequireUpper:   true,
	RequireLower:   true,
	RequireDigit:   true,
	RequireSpecial: false,
	MaxConsecutive: 4,
	DisallowCommon: true,
}

// CommonPasswords is a small set of commonly used passwords to reject.
// In production, use a larger list or external service.
var CommonPasswords = map[string]bool{
	"password":     true,
	"password123":  true,
	"password1":    true,
	"12345678":     true,
	"123456789":    true,
	"qwerty":       true,
	"qwerty123":    true,
	"admin":        true,
	"admin123":     true,
	"letmein":      true,
	"welcome":      true,
	"welcome123":   true,
	"monkey":       true,
	"dragon":       true,
	"master":       true,
	"login":        true,
	"abc123":       true,
	"starwars":     true,
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

	if policy.MaxConsecutive > 0 && containsConsecutive(password, policy.MaxConsecutive) {
		missing = append(missing, fmt.Sprintf("no more than %d consecutive identical characters", policy.MaxConsecutive))
	}

	if policy.DisallowCommon && isCommonPassword(password) {
		missing = append(missing, "password is too common, choose a stronger password")
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

// containsConsecutive checks for repeated characters (e.g., "aaa").
func containsConsecutive(s string, maxConsecutive int) bool {
	if len(s) < maxConsecutive {
		return false
	}
	count := 1
	for i := 1; i < len(s); i++ {
		if s[i] == s[i-1] {
			count++
			if count >= maxConsecutive {
				return true
			}
		} else {
			count = 1
		}
	}
	return false
}

// isCommonPassword checks if a password is in the common passwords list.
func isCommonPassword(password string) bool {
	lower := strings.ToLower(password)
	return CommonPasswords[lower]
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
