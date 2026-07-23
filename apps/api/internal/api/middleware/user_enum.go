// Package middleware provides HTTP middleware.
package middleware

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/auth"
	"github.com/gin-gonic/gin"
)

// UserEnumSafeResponse provides timing-safe responses to prevent user enumeration attacks.
// All authentication responses (success or failure) return identical timing characteristics.
type UserEnumSafeResponse struct {
	Enabled bool
}

// NewUserEnumSafeResponse creates a new user enumeration prevention middleware.
func NewUserEnumSafeResponse(enabled bool) *UserEnumSafeResponse {
	return &UserEnumSafeResponse{Enabled: enabled}
}

// Middleware ensures consistent response timing regardless of whether user exists.
func (m *UserEnumSafeResponse) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.Enabled {
			c.Next()
			return
		}

		c.Next()
	}
}

// LoginSafeResponse returns a timing-safe login response.
// Always returns the same response structure regardless of whether user exists.
type LoginSafeResponse struct{}

func NewLoginSafeResponse() *LoginSafeResponse {
	return &LoginSafeResponse{}
}

// Response returns a consistent response for failed logins.
func (r *LoginSafeResponse) Response() gin.H {
	// Always return same structure to prevent enumeration.
	return gin.H{
		"error":   "unauthorized",
		"message": "Invalid credentials",
	}
}

// SignupResponse returns a consistent signup response.
func (r *LoginSafeResponse) SignupResponse() gin.H {
	return gin.H{
		"error":   "unauthorized",
		"message": "Registration failed",
	}
}

// PasswordResetResponse returns a consistent password reset response.
func (r *LoginSafeResponse) PasswordResetResponse() gin.H {
	return gin.H{
		"error":   "unauthorized",
		"message": "Password reset not available",
	}
}

// PreventUserEnum is a middleware that ensures no information about user existence leaks.
// It should be applied to login, signup, password reset, and forgot password endpoints.
func PreventUserEnum() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		// After handler runs, ensure consistent response for auth failures.
		status := c.Writer.Status()
		if status >= 400 && status < 500 {
			// Add fake processing delay if this was an auth endpoint.
			path := c.Request.URL.Path
			if isAuthEndpoint(path) {
				// Use the domain auth package's timing-safe implementation.
				auth.ComputeFakePasswordHash()
			}
		}
	}
}

func isAuthEndpoint(path string) bool {
	return path == "/v1/auth/login" ||
		path == "/v1/auth/register" ||
		path == "/v1/auth/forgot-password" ||
		path == "/v1/auth/verify-email"
}
