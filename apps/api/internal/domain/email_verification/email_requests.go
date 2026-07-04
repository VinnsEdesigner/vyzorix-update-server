package email_verification

// VerifyEmailRequest is the payload for verifying an email address.
type VerifyEmailRequest struct {
	Token string `json:"token"`
}

// ResendVerificationRequest is the payload for requesting a new verification email.
type ResendVerificationRequest struct {
	Email string `json:"email"`
}

// CancelVerificationRequest is the payload for cancelling email verification.
type CancelVerificationRequest struct {
	Token string `json:"token"`
}
