package dto

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	OperatorID string `json:"operator_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
	MFAEnabled bool   `json:"mfa_enabled"`
}

// RegisterRequest represents a registration request.
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Role     string `json:"role,omitempty"`
}

// RegisterResponse represents a registration response.
type RegisterResponse struct {
	OperatorID string `json:"operator_id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
}

// ForgotPasswordRequest represents a forgot password request.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest represents a password reset request.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"new_password"`
}

// ChangePasswordRequest represents a password change request.
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

// RefreshTokenRequest represents a token refresh request.
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	AllDevices bool `json:"all_devices"`
}

// OperatorResponse represents an operator in responses.
type OperatorResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	MFAEnabled    bool   `json:"mfa_enabled"`
	EmailVerified  bool   `json:"email_verified"`
	CreatedAt      string `json:"created_at"`
}

// UpdateOperatorRequest represents an operator update request.
type UpdateOperatorRequest struct {
	Name string `json:"name"`
}

// MFAStatusResponse represents MFA status.
type MFAStatusResponse struct {
	Enabled   bool     `json:"enabled"`
	BackupCodes []string `json:"backup_codes,omitempty"`
}

// MFAEnrollRequest represents MFA enrollment request.
type MFAEnrollRequest struct {
	Code string `json:"code"`
}

// MFAVerifyRequest represents MFA verification request.
type MFAVerifyRequest struct {
	Code string `json:"code"`
}

// OAuthCallbackRequest represents OAuth callback.
type OAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// OperatorListResponse represents an operator in list responses (admin view).
type OperatorListResponse struct {
	ID            string `json:"id"`
	Email         string `json:"email"`
	Name          string `json:"name"`
	Role          string `json:"role"`
	MFAEnabled    bool   `json:"mfa_enabled"`
	EmailVerified bool   `json:"email_verified"`
	CreatedAt     int64  `json:"created_at"`
}
