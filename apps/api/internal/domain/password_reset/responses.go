package password_reset

// ResendPasswordResetResponse indicates a reset email was sent.
type ResendPasswordResetResponse struct {
	Message string `json:"message"`
}

// PasswordResetResponse indicates the password was reset successfully.
type PasswordResetResponse struct {
	Message string `json:"message"`
}
