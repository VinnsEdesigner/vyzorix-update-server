package password_reset

// ForgotPasswordRequest is the payload for initiating a password reset.
type ForgotPasswordRequest struct {
	Email string `json:"email"`
}

// ResetPasswordRequest is the payload for completing a password reset.
type ResetPasswordRequest struct {
	Token       string `json:"token"`
	NewPassword string `json:"newPassword"`
}

// ResendPasswordResetRequest is the payload for requesting a new reset email.
type ResendPasswordResetRequest struct {
	Email string `json:"email"`
}
