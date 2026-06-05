package models

import "time"

// OperatorRole represents the role of an operator in the system.
type OperatorRole string

const (
	RoleViewer    OperatorRole = "viewer"
	RoleOperator  OperatorRole = "operator"
	RoleSuperAdmin OperatorRole = "super_admin"
)

// Operator represents a human operator who can access the dashboard.
type Operator struct {
	ID            string       `json:"id"`
	Email         string       `json:"email"`
	Name          string       `json:"name"`
	PasswordHash  string       `json:"-"` // Never exposed via JSON
	Role          OperatorRole `json:"role"`
	GoogleID      string       `json:"googleId,omitempty"`
	EmailVerified bool         `json:"emailVerified,omitempty"`
	CreatedAt     time.Time    `json:"createdAt"`
	UpdatedAt     time.Time    `json:"updatedAt"`
}

// OperatorResponse is the safe JSON representation returned to clients.
type OperatorResponse struct {
	ID            string       `json:"id"`
	Email         string       `json:"email"`
	Name          string       `json:"name"`
	Role          OperatorRole `json:"role"`
	EmailVerified bool         `json:"emailVerified,omitempty"`
	CreatedAt     int64        `json:"createdAt"`
}

// ToResponse converts an Operator to its safe JSON representation.
func (o *Operator) ToResponse() OperatorResponse {
	return OperatorResponse{
		ID:            o.ID,
		Email:         o.Email,
		Name:          o.Name,
		Role:          o.Role,
		EmailVerified: o.EmailVerified,
		CreatedAt:     o.CreatedAt.UnixMilli(),
	}
}

// Session represents an active operator session.
type Session struct {
	ID         string    `json:"id"`
	OperatorID string    `json:"operatorId"`
	TokenHash  string    `json:"-"` // never exposed via JSON
	ExpiresAt  time.Time `json:"expiresAt"`
	CreatedAt  time.Time `json:"createdAt"`
	UserAgent  string    `json:"userAgent,omitempty"`
	IPAddress  string    `json:"ipAddress,omitempty"`
}

// LoginRequest is the payload for email/password login.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest is the payload for operator self-registration.
// Only allowed when no operators exist in the system (bootstrap phase).
type OperatorRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// AuthResponse is returned on successful login or registration.
type AuthResponse struct {
	Token     string            `json:"token"`
	ExpiresAt int64             `json:"expiresAt"`
	Operator  OperatorResponse  `json:"operator"`
}

// GoogleOAuthURLRequest asks the server for the Google OAuth authorization URL.
type GoogleOAuthURLRequest struct{}

// GoogleOAuthURLResponse returns the URL to redirect the browser to.
type GoogleOAuthURLResponse struct {
	URL string `json:"url"`
}

// GoogleOAuthCallbackRequest is the internal payload after Google validates the callback.
type GoogleOAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// ErrorResponse is the standard error envelope for auth endpoints.
type AuthErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// UpdateNameRequest is the payload for updating the operator's display name.
type UpdateNameRequest struct {
	Name string `json:"name"`
}

// VerifyEmailRequest is the payload for verifying an email address.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ResendVerificationRequest is the payload for requesting a new verification email.
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// ForgotPasswordRequest is the payload for requesting a password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest is the payload for resetting a password.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// MessageResponse is a simple message response for certain auth operations.
type MessageResponse struct {
	Message string `json:"message"`
}

// EmailVerifiedResponse is returned after successful email verification.
type EmailVerifiedResponse struct {
	Verified bool   `json:"verified"`
	Email    string `json:"email,omitempty"`
}
