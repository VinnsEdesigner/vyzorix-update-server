package operator

import (
	"strings"
	"time"
)

// Thresholds define alert levels for device telemetry.
type Thresholds struct {
	RiskWarn    int `json:"riskWarn"`
	RiskCrit    int `json:"riskCrit"`
	ThermalWarn int `json:"thermalWarn"`
	ThermalCrit int `json:"thermalCrit"`
	BufferWarn  int `json:"bufferWarn"`
	BufferCrit  int `json:"bufferCrit"`
}

// ClientSettings holds operator preferences that control dashboard behavior.
type ClientSettings struct {
	ServerURL           string `json:"serverUrl"`
	DeviceID            string `json:"deviceId"`
	RequestTimeoutMs    int    `json:"requestTimeoutMs"`
	AutoReconnect       bool   `json:"autoReconnect"`
	StrictHmac          bool   `json:"strictHmac"`
	LogBufferLimit      int    `json:"logBufferLimit"`
	SignalHistoryLimit  int    `json:"signalHistoryLimit"`
	NotificationsEnabled bool   `json:"notificationsEnabled"`
}

// OperatorSettings represents all settings for an operator.
type OperatorSettings struct {
	Client        ClientSettings      `json:"client"`
	Thresholds    Thresholds          `json:"thresholds"`
	Notifications *NotificationSettings `json:"notifications"`
}

// DefaultClientSettings returns default client settings.
func DefaultClientSettings() *ClientSettings {
	return &ClientSettings{
		ServerURL:           "",
		DeviceID:            "",
		RequestTimeoutMs:    8000,
		AutoReconnect:       true,
		StrictHmac:          false,
		LogBufferLimit:      500,
		SignalHistoryLimit:  240,
		NotificationsEnabled: true,
	}
}

// Operator represents a system operator (user).
type Operator struct {
	CreatedAt      time.Time
	UpdatedAt      time.Time
	GitHubID       string
	PasswordHash   string
	Role           OperatorRole
	GoogleID       string
	ID             string
	MFASecret      string
	Name           string
	Email          string
	BackupCodes    []string
	Thresholds     Thresholds     `json:"thresholds"`
	ClientSettings ClientSettings `json:"client"`
	MFAEnabled     bool
	MFARequired    bool           // HIGH-6: Forces MFA for this operator
	EmailVerified  bool
	FCMToken       string `json:"fcmToken,omitempty"` // FCM token for push notifications
}

// IsSuperAdmin returns true if the operator is a super admin.
func (o *Operator) IsSuperAdmin() bool {
	return o.Role == RoleSuperAdmin
}

// IsAdmin returns true if the operator is an admin or super admin.
func (o *Operator) IsAdmin() bool {
	return o.Role == RoleSuperAdmin
}

// CanManageOperators returns true if the operator can manage other operators.
func (o *Operator) CanManageOperators() bool {
	return o.IsAdmin()
}

// CanManageDevices returns true if the operator can manage devices.
func (o *Operator) CanManageDevices() bool {
	return o.IsAdmin()
}

// HasMFA returns true if MFA is enabled for this operator.
func (o *Operator) HasMFA() bool {
	return o.MFAEnabled && o.MFASecret != ""
}

// UsesOAuth returns true if this operator uses OAuth (Google or GitHub).
func (o *Operator) UsesOAuth() bool {
	return o.GoogleID != "" || o.GitHubID != ""
}

// HasPassword returns true if this operator has a password set.
func (o *Operator) HasPassword() bool {
	return o.PasswordHash != ""
}

// ValidateEmail validates the email format.
func (o *Operator) ValidateEmail() bool {
	return strings.Contains(o.Email, "@") && strings.Contains(o.Email, ".")
}

// IsValid returns true if the operator has all required fields.
func (o *Operator) IsValid() bool {
	return o.ID != "" && o.Email != "" && o.Name != ""
}
