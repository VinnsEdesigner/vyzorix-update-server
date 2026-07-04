package operator

// EmailNotifications holds email notification preferences.
type EmailNotifications struct {
	ThresholdBreach    bool `json:"thresholdBreach"`
	DeviceOffline      bool `json:"deviceOffline"`
	DeviceOnline       bool `json:"deviceOnline"`
	UpdateAvailable    bool `json:"updateAvailable"`
	CommandFailed      bool `json:"commandFailed"`
	RegistrationRequest bool `json:"registrationRequest"`
}

// PushNotifications holds push notification preferences.
type PushNotifications struct {
	ThresholdBreach    bool `json:"thresholdBreach"`
	DeviceOffline      bool `json:"deviceOffline"`
	DeviceOnline       bool `json:"deviceOnline"`
	UpdateAvailable    bool `json:"updateAvailable"`
	CommandFailed      bool `json:"commandFailed"`
	RegistrationRequest bool `json:"registrationRequest"`
}

// WebhookNotifications holds webhook notification configuration.
type WebhookNotifications struct {
	Enabled bool   `json:"enabled"`
	URL     string `json:"url"`
	Secret  string `json:"secret,omitempty"`
	Types   []string `json:"types"`
}

// NotificationSettings holds all notification preferences for an operator.
type NotificationSettings struct {
	Enabled   bool                  `json:"enabled"`
	Channels  []string              `json:"channels"`
	Email     EmailNotifications    `json:"email"`
	Push      PushNotifications     `json:"push"`
	Webhook   WebhookNotifications  `json:"webhook"`
}

// DefaultNotificationSettings returns default notification settings.
func DefaultNotificationSettings() *NotificationSettings {
	return &NotificationSettings{
		Enabled:  true,
		Channels: []string{"email"},
		Email: EmailNotifications{
			ThresholdBreach:     true,
			DeviceOffline:       true,
			DeviceOnline:        false,
			UpdateAvailable:     false,
			CommandFailed:       true,
			RegistrationRequest: false,
		},
		Push: PushNotifications{
			ThresholdBreach:     true,
			DeviceOffline:       true,
			DeviceOnline:        false,
			UpdateAvailable:     false,
			CommandFailed:       false,
			RegistrationRequest: false,
		},
		Webhook: WebhookNotifications{
			Enabled: false,
			URL:     "",
			Secret:  "",
			Types:   []string{},
		},
	}
}

// NotificationType constants.
const (
	NotificationTypeThresholdBreach     = "threshold_breach"
	NotificationTypeDeviceOffline      = "device_offline"
	NotificationTypeDeviceOnline       = "device_online"
	NotificationTypeUpdateAvailable    = "update_available"
	NotificationTypeCommandFailed      = "command_failed"
	NotificationTypeRegistrationRequest = "registration_request"
)

// ValidNotificationTypes returns all valid notification type strings.
func ValidNotificationTypes() []string {
	return []string{
		NotificationTypeThresholdBreach,
		NotificationTypeDeviceOffline,
		NotificationTypeDeviceOnline,
		NotificationTypeUpdateAvailable,
		NotificationTypeCommandFailed,
		NotificationTypeRegistrationRequest,
	}
}

// IsValidNotificationType checks if a notification type is valid.
func IsValidNotificationType(t string) bool {
	for _, valid := range ValidNotificationTypes() {
		if t == valid {
			return true
		}
	}
	return false
}
