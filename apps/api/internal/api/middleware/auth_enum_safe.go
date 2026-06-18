// Package middleware provides HTTP middleware for the Vyzorix API.
package middleware

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// AuthEnumSafe provides safe authentication responses to prevent user enumeration.
type AuthEnumSafe struct {
	// ResponseDelay is the constant delay to add to auth responses.
	ResponseDelay time.Duration
}

// NewAuthEnumSafe creates a new safe auth response handler.
func NewAuthEnumSafe() *AuthEnumSafe {
	return &AuthEnumSafe{
		ResponseDelay: 100 * time.Millisecond,
	}
}

// SafeErrorResponse returns a generic error message.
func (a *AuthEnumSafe) SafeErrorResponse(c *gin.Context, status int, message string) {
	// Add constant delay to prevent timing attacks.
	time.Sleep(a.ResponseDelay)
	c.JSON(status, gin.H{
		"error":   message,
		"code":    "AUTH_ERROR",
		"success": false,
	})
}

// SafeLoginError returns a uniform login error.
func (a *AuthEnumSafe) SafeLoginError(c *gin.Context) {
	a.SafeErrorResponse(c, http.StatusUnauthorized, "Invalid credentials")
}

// SafeSignupError returns a uniform signup error.
func (a *AuthEnumSafe) SafeSignupError(c *gin.Context) {
	a.SafeErrorResponse(c, http.StatusBadRequest, "Invalid request")
}

// SafeResetError returns a uniform password reset error.
func (a *AuthEnumSafe) SafeResetError(c *gin.Context) {
	a.SafeErrorResponse(c, http.StatusBadRequest, "Invalid request")
}

// ConstantTimeValidate performs constant-time validation.
func ConstantTimeValidate(expected, actual string) bool {
	if len(expected) != len(actual) {
		// Still do comparison to maintain timing.
		subtle.ConstantTimeCompare([]byte(expected), []byte(actual))
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expected), []byte(actual)) == 1
}

// SafeUserLookup simulates a user lookup with constant time.
func (a *AuthEnumSafe) SafeUserLookup(email string) (bool, string) {
	// This should always take the same time regardless of whether.
	// the user exists or not.
	time.Sleep(a.ResponseDelay)

	// Generate a realistic-looking Argon2id format fake hash for non-existent users.
	fakeSalt := make([]byte, 16)
	fakeHash := make([]byte, 32)
	if _, err := rand.Read(fakeSalt); err != nil {
		// Fallback - should never happen.
		for i := range fakeSalt {
			fakeSalt[i] = byte(i)
		}
	}
	if _, err := rand.Read(fakeHash); err != nil {
		// Fallback - should never happen.
		for i := range fakeHash {
			fakeHash[i] = byte(i)
		}
	}
	fakeHashStr := "$argon2id$v=19$m=65536,t=3,p=4$" +
		base64.RawStdEncoding.EncodeToString(fakeSalt) + "$" +
		base64.RawStdEncoding.EncodeToString(fakeHash)

	return false, fakeHashStr
}
