package models

import "time"

// OperatorRole represents the role of an operator in the system.
type OperatorRole string

// RoleViewer is the default read-only role.
const RoleViewer OperatorRole = "viewer"

// RoleOperator can perform standard operations.
const RoleOperator OperatorRole = "operator"

// RoleSuperAdmin has full system access.
const RoleSuperAdmin OperatorRole = "super_admin"

// Thresholds define alert levels for device telemetry.
type Thresholds struct {
	RiskWarn    int `json:"riskWarn" db:"risk_warn"`
	RiskCrit    int `json:"riskCrit" db:"risk_crit"`
	ThermalWarn int `json:"thermalWarn" db:"thermal_warn"`
	ThermalCrit int `json:"thermalCrit" db:"thermal_crit"`
	BufferWarn  int `json:"bufferWarn" db:"buffer_warn"`
	BufferCrit  int `json:"bufferCrit" db:"buffer_crit"`
}

// ClientSettings holds operator preferences that control dashboard behavior.
type ClientSettings struct {
	StrictHmac           bool `json:"strictHmac" db:"strict_hmac"`
	AutoReconnect        bool `json:"autoReconnect" db:"auto_reconnect"`
	NotificationsEnabled bool `json:"notificationsEnabled" db:"notifications_enabled"`
}

// Operator represents a human operator who can access the dashboard.
type Operator struct {
	CreatedAt     time.Time      `json:"createdAt"`
	UpdatedAt     time.Time      `json:"updatedAt"`
	ID            string         `json:"id"`
	Email         string         `json:"email"`
	Name          string         `json:"name"`
	PasswordHash  string         `json:"-"`
	Role          OperatorRole   `json:"role"`
	GoogleID      string         `json:"googleId,omitempty"`
	Thresholds    Thresholds     `json:"thresholds,omitempty" db:"-"`
	Client        ClientSettings `json:"client,omitempty" db:"-"`
	EmailVerified bool           `json:"emailVerified,omitempty"`
}

// OperatorResponse is the safe JSON representation returned to clients.
type OperatorResponse struct {
	Thresholds    *Thresholds     `json:"thresholds,omitempty"`
	Client        *ClientSettings `json:"client,omitempty"`
	ID            string          `json:"id"`
	Email         string          `json:"email"`
	Name          string          `json:"name"`
	Role          OperatorRole    `json:"role"`
	CreatedAt     int64           `json:"createdAt"`
	EmailVerified bool            `json:"emailVerified,omitempty"`
}

// ToResponse converts an Operator to its safe JSON representation.
func (o *Operator) ToResponse() OperatorResponse {
	return OperatorResponse{
		ID:            o.ID,
		Email:         o.Email,
		Name:          o.Name,
		Role:          o.Role,
		EmailVerified: o.EmailVerified,
		Thresholds:    &o.Thresholds,
		Client:        &o.Client,
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

// OperatorRegisterRequest is the payload for operator self-registration.
// Only allowed when no operators exist in the system (bootstrap phase).
type OperatorRegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

// AuthResponse is returned on successful login or registration.
type AuthResponse struct {
	Token     string           `json:"token"`
	Operator  OperatorResponse `json:"operator"`
	ExpiresAt int64            `json:"expiresAt"`
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

// AuthErrorResponse is the standard error envelope for auth endpoints.
type AuthErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// UpdateNameRequest is the payload for updating the operator's display name.
type UpdateNameRequest struct {
	Name *string `json:"name,omitempty"`
}

// UpdateSettingsRequest is the payload for updating operator settings (name, thresholds, and client preferences).
type UpdateSettingsRequest struct {
	Name       *string         `json:"name,omitempty"`
	Thresholds *Thresholds     `json:"thresholds,omitempty"`
	Client     *ClientSettings `json:"client,omitempty"`
	Reset      bool            `json:"reset,omitempty"` // Reset all settings to defaults
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
	Email     string `json:"email,omitempty"`
	Verified  bool   `json:"verified"`
	AutoLogin bool   `json:"autoLogin,omitempty"`
}

// VerificationPollResponse is the response for polling verification status.
type VerificationPollResponse struct {
	Status string `json:"status"` // "waiting", "success", "expired", "invalid"
	Email  string `json:"email,omitempty"`
}

// CancelVerificationRequest is the payload for canceling pending verification.
type CancelVerificationRequest struct {
	Email string `json:"email"`
}

// ResendPasswordResetRequest is the payload for requesting a password reset resend.
type ResendPasswordResetRequest struct {
	Email string `json:"email"`
}

// ResendPasswordResetResponse is the response for password reset resend requests.
type ResendPasswordResetResponse struct {
	Success     bool   `json:"success"`
	Message     string `json:"message"`
	RetryAfter  int    `json:"retry_after,omitempty"`  // Seconds until next resend allowed
	LockedUntil int64  `json:"locked_until,omitempty"` // Unix timestamp when lockout ends
}

// PasswordResetResendTracker tracks resend attempts for rate limiting.
type PasswordResetResendTracker struct {
	ID           string     `json:"id"`
	EmailHash    string     `json:"-"` // never exposed via JSON
	ResendCount  int        `json:"resend_count"`
	LastResendAt time.Time  `json:"last_resend_at"`
	LockoutUntil *time.Time `json:"lockout_until,omitempty"` // Unix timestamp when lockout ends
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
