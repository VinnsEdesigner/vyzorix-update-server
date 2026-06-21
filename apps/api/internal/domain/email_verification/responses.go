package email_verification

// EmailVerifiedResponse indicates the email was verified successfully.
type EmailVerifiedResponse struct {
	Message string `json:"message"`
}

// VerificationPollResponse is returned when polling for email verification status.
type VerificationPollResponse struct {
	Verified  bool   `json:"verified"`
	Message   string `json:"message,omitempty"`
	ExpiresIn int64  `json:"expiresIn,omitempty"` // Seconds remaining
}

// ResendVerificationResponse indicates a verification email was sent.
type ResendVerificationResponse struct {
	Message string `json:"message"`
}
