package dto

import "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"

// LoginRequest represents a login request.
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// OrganizationInfo represents an organization with the operator's role in it.
type OrganizationInfo struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Role string `json:"role"`
}

// LoginResponse represents a login response.
type LoginResponse struct {
	OperatorID            string             `json:"operator_id"`
	Email                 string             `json:"email"`
	Name                  string             `json:"name"`
	MFAEnabled            bool               `json:"mfa_enabled"`
	NeedsOrganization     bool               `json:"needs_organization"`
	Organizations         []OrganizationInfo `json:"organizations,omitempty"`
	LastOrganizationID    string            `json:"last_organization_id,omitempty"`
	SelectedOrganization  *OrganizationInfo `json:"selected_organization,omitempty"`
}

// LoginWithTokensResponse represents a login response with API tokens.
type LoginWithTokensResponse struct {
	OperatorID            string             `json:"operator_id"`
	Email                 string             `json:"email"`
	Name                  string             `json:"name"`
	MFAEnabled            bool               `json:"mfa_enabled"`
	NeedsOrganization     bool               `json:"needs_organization"`
	Organizations         []OrganizationInfo `json:"organizations,omitempty"`
	LastOrganizationID    string            `json:"last_organization_id,omitempty"`
	SelectedOrganization  *OrganizationInfo `json:"selected_organization,omitempty"`
	AccessToken          string             `json:"access_token"`
	RefreshToken         string             `json:"refresh_token"`
	ExpiresAt            int64              `json:"expires_at"`
	SessionID            string             `json:"session_id"`
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
	Email      string `json:"email"`
	Name       string `json:"name"`
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

// RefreshTokenResponse represents a token refresh response.
type RefreshTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	SessionID    string `json:"session_id"`
	ExpiresAt    int64  `json:"expires_at"`
}

// LogoutRequest represents a logout request.
type LogoutRequest struct {
	AllDevices bool `json:"all_devices"`
}

// OperatorResponse represents an operator in responses.
type OperatorResponse struct {
	Thresholds    *Thresholds     `json:"thresholds,omitempty"`
	Client        *ClientSettings `json:"client,omitempty"`
	ID            string          `json:"id"`
	Email         string          `json:"email"`
	Name          string          `json:"name"`
	NeedsOrganization     bool             `json:"needs_organization"`
	Organizations         []OrganizationInfo `json:"organizations,omitempty"`
	LastOrganizationID    string           `json:"last_organization_id,omitempty"`
	SelectedOrganization  *OrganizationInfo `json:"selected_organization,omitempty"`
	CreatedAt     string          `json:"created_at"`
	MFAEnabled    bool            `json:"mfa_enabled"`
	EmailVerified bool            `json:"email_verified"`
}

// GoogleOAuthCallbackRequest represents a Google OAuth callback request.
type GoogleOAuthCallbackRequest struct {
	Code  string `json:"code"`
	State string `json:"state"`
}

// UpdateOperatorRequest represents an operator update request.
type UpdateOperatorRequest struct {
	Name string `json:"name"`
}

// MFAStatusResponse represents MFA status.
type MFAStatusResponse struct {
	BackupCodes []string `json:"backup_codes,omitempty"`
	Enabled     bool     `json:"enabled"`
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

// RoleOperator is the role for regular operators.
const RoleOperator = "operator"

// OperatorRegisterRequest represents a registration request (alias for RegisterRequest).
type OperatorRegisterRequest = RegisterRequest

// AuthResponse represents an authentication response.
type AuthResponse struct {
	Token    string           `json:"token"`
	Operator OperatorResponse `json:"operator"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// UpdateNameRequest represents a name update request.
type UpdateNameRequest struct {
	Name string `json:"name"`
}

// Thresholds represents threshold settings (alias to domain).
type Thresholds = operator.Thresholds

// ClientSettings represents client settings (alias to domain).
type ClientSettings = operator.ClientSettings

// UpdateSettingsRequest represents a settings update request.
type UpdateSettingsRequest struct {
	Thresholds *Thresholds     `json:"thresholds,omitempty"`
	Client     *ClientSettings `json:"client,omitempty"`
	Name       string          `json:"name,omitempty"`
	Reset      bool            `json:"reset,omitempty"`
}
