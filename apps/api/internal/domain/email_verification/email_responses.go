package email_verification

// EmailVerifiedResponse indicates the email was verified successfully.
type EmailVerifiedResponse struct {
	Message string `json:"message"`
}

// VerificationPollResponse is returned when polling for email verification status.
type VerificationPollResponse struct {
	Message   string `json:"message,omitempty"`
	ExpiresIn int64  `json:"expiresIn,omitempty"`
	Verified  bool   `json:"verified"`
}

// PollVerificationStatus represents the possible statuses for email verification polling.
type PollVerificationStatus string

const (
	PollStatusInvalid     PollVerificationStatus = "invalid"
	PollStatusExpired     PollVerificationStatus = "expired"
	PollStatusEmailFailed PollVerificationStatus = "email_failed"
	PollStatusWaiting     PollVerificationStatus = "waiting"
	PollStatusSuccess     PollVerificationStatus = "success"
)

// PollVerificationResult contains all information about a verification poll result.
type PollVerificationResult struct {
	Status     PollVerificationStatus `json:"status"`
	Email      string                 `json:"email,omitempty"`
	EmailError string                 `json:"emailError,omitempty"`
}

// ResendVerificationResponse indicates a verification email was sent.
type ResendVerificationResponse struct {
	Message string `json:"message"`
}
