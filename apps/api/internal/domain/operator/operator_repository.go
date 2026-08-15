package operator

import (
	"context"
	"time"
)

// Repository defines the interface for operator data access.
type Repository interface {
	// FindByID retrieves an operator by ID.
	FindByID(ctx context.Context, id string) (*Operator, error)

	// FindByEmail retrieves an operator by email.
	FindByEmail(ctx context.Context, email string) (*Operator, error)

	// FindByGoogleID retrieves an operator by Google ID.
	FindByGoogleID(ctx context.Context, googleID string) (*Operator, error)

	// FindByGitHubID retrieves an operator by GitHub ID.
	FindByGitHubID(ctx context.Context, githubID string) (*Operator, error)

	// Create creates a new operator.
	Create(ctx context.Context, op *Operator) error

	// Update updates an existing operator.
	Update(ctx context.Context, op *Operator) error

	// Delete deletes an operator.
	Delete(ctx context.Context, id string) error

	// Count returns the total number of operators.
	Count(ctx context.Context) (int, error)

	// List returns a paginated list of operators.
	List(ctx context.Context, limit, offset int) ([]*Operator, int, error)

	// UpdatePassword updates the password hash for an operator.
	UpdatePassword(ctx context.Context, id, passwordHash string) error

	// UpdateMFA updates MFA settings for an operator.
	// When enabled is true, mfaEnabledAt should be set to the current time.
	// When enabled is false, mfaEnabledAt is ignored (will be set to nil).
	UpdateMFA(ctx context.Context, id, secret, secretMAC string, enabled bool, mfaEnabledAt *time.Time) error

	// UpdateOperatorMFA updates the MFA secret, MAC, and backup codes for an operator.
	UpdateOperatorMFA(ctx context.Context, operatorID, mfaSecret, mfaSecretMAC string, backupCodes []string) error

	// VerifyEmail marks an operator's email as verified.
	VerifyEmail(ctx context.Context, id string) error

	// UpdateEmailVerified updates the email verified status for an operator.
	UpdateEmailVerified(ctx context.Context, id string, verified bool) error

	// UpdateGoogleID updates the Google ID for an operator.
	UpdateGoogleID(ctx context.Context, id, googleID string) error

	// UpdateGitHubID updates the GitHub ID for an operator.
	UpdateGitHubID(ctx context.Context, id, githubID string) error

	// UpdateName updates the display name for an operator.
	UpdateName(ctx context.Context, id, name string) error

	// UpdateThresholds updates the alert thresholds for an operator.
	UpdateThresholds(ctx context.Context, id string, th Thresholds) error

	// GetThresholds retrieves the alert thresholds for an operator.
	GetThresholds(ctx context.Context, id string) (Thresholds, error)

	// UpdateClientSettings updates the client preferences for an operator.
	UpdateClientSettings(ctx context.Context, id string, cs ClientSettings) error

	// ResetSettings resets all settings to defaults for an operator.
	ResetSettings(ctx context.Context, id string) error

	// GetEmailVerified returns whether an operator has verified their email.
	GetEmailVerified(ctx context.Context, id string) (bool, error)

	// DisableMFA disables MFA for an operator by clearing the MFA secret and backup codes.
	DisableMFA(ctx context.Context, id string) error

	// GetSetting retrieves a setting value by key.
	GetSetting(ctx context.Context, key string) (string, error)

	// SetSetting updates or inserts a setting value.
	SetSetting(ctx context.Context, key, value string) error

	// GetEnforceHMAC returns whether HMAC enforcement is enabled.
	GetEnforceHMAC(ctx context.Context) (bool, error)

	// SetEnforceHMAC updates the HMAC enforcement setting.
	SetEnforceHMAC(ctx context.Context, enforce bool) error

	// GetHMACWindowSeconds returns the HMAC timestamp window in seconds.
	GetHMACWindowSeconds(ctx context.Context) (int, error)

	// SetHMACWindowSeconds updates the HMAC timestamp window.
	SetHMACWindowSeconds(ctx context.Context, seconds int) error

	// GetOperatorSettings retrieves all settings for an operator.
	GetOperatorSettings(ctx context.Context, operatorID string) (*OperatorSettings, error)

	// GetNotifications retrieves notification settings for an operator.
	GetNotifications(ctx context.Context, operatorID string) (*NotificationSettings, error)

	// UpdateNotifications updates notification settings for an operator.
	UpdateNotifications(ctx context.Context, operatorID string, settings *NotificationSettings) error

	// RotateWebhookSecret generates a new webhook secret for an operator.
	RotateWebhookSecret(ctx context.Context, operatorID string) (string, error)

	// UpdateFCMToken updates the FCM token for an operator.
	UpdateFCMToken(ctx context.Context, operatorID, fcmToken string) error
}
