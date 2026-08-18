package operator

import (
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
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

// OperatorSettings holds all settings for an operator.
type OperatorSettings struct {
	Notifications *NotificationSettings `json:"notifications"`
	Client        ClientSettings        `json:"client"`
	Thresholds    Thresholds            `json:"thresholds"`
}

// ClientSettings control dashboard behavior.
type ClientSettings struct {
	ServerURL            string `json:"serverUrl"`
	DeviceID             string `json:"deviceId"`
	RequestTimeoutMs     int    `json:"requestTimeoutMs"`
	LogBufferLimit       int    `json:"logBufferLimit"`
	SignalHistoryLimit   int    `json:"signalHistoryLimit"`
	AutoReconnect        bool   `json:"autoReconnect"`
	StrictHmac           bool   `json:"strictHmac"`
	NotificationsEnabled bool   `json:"notificationsEnabled"`
}

// SecuritySettings holds security-related settings per operator.
type SecuritySettings struct {
	MaxConcurrentSessions int  `json:"maxConcurrentSessions"` // 0 = unlimited.
	PasswordMinAgeDays    int  `json:"passwordMinAgeDays"`    // 0 = no minimum.
	PasswordMaxAgeDays    int  `json:"passwordMaxAgeDays"`    // 0 = no expiry.
	PasswordHistoryCount  int  `json:"passwordHistoryCount"`  // remember N passwords.
	SessionPinRequired    bool `json:"sessionPinRequired"`    // require PIN for sensitive ops.
}

// Operator represents a system operator (user).
type Operator struct {
	CreatedAt          time.Time
	UpdatedAt          time.Time
	MFAEnabledAt       *time.Time
	PasswordHash       string
	Email              string
	ID                 string
	MFASecret          string
	MFASecretMAC       string
	Name               string
	GitHubID           string
	FCMToken           string `json:"fcmToken,omitempty"`
	GoogleID           string
	LastOrganizationID string
	BackupCodes        []string
	Memberships        []*organization.OrganizationMember
	ClientSettings     ClientSettings   `json:"client"`
	Thresholds         Thresholds       `json:"thresholds"`
	SecuritySettings   SecuritySettings `json:"security"`
	MFARequired        bool
	EmailVerified      bool
	MFAEnabled         bool
}

// DefaultSecuritySettings returns default security settings.
func DefaultSecuritySettings() SecuritySettings {
	return SecuritySettings{
		MaxConcurrentSessions: 3,
		PasswordMinAgeDays:    0,
		PasswordMaxAgeDays:    90,
		PasswordHistoryCount:  5,
		SessionPinRequired:    false,
	}
}

// DefaultClientSettings returns default client settings.
func DefaultClientSettings() *ClientSettings {
	return &ClientSettings{
		ServerURL:            "",
		DeviceID:             "",
		RequestTimeoutMs:     8000,
		AutoReconnect:        true,
		StrictHmac:           false,
		LogBufferLimit:       500,
		SignalHistoryLimit:   100,
		NotificationsEnabled: true,
	}
}

// DefaultThresholds returns default thresholds.
func DefaultThresholds() Thresholds {
	return Thresholds{
		RiskWarn:    70,
		RiskCrit:    90,
		ThermalWarn: 75,
		ThermalCrit: 85,
		BufferWarn:  80,
		BufferCrit:  95,
	}
}

// IsSuperAdminIn returns true if the operator is a super admin in the specified organization.
func (o *Operator) IsSuperAdminIn(orgID string) bool {
	m := o.GetMembership(orgID)
	if m == nil {
		return false
	}
	return m.Role.IsSuperAdmin() && m.IsActive()
}

// IsAdminIn returns true if the operator is an admin or super admin in the specified organization.
func (o *Operator) IsAdminIn(orgID string) bool {
	m := o.GetMembership(orgID)
	if m == nil {
		return false
	}
	return m.Role.IsAdmin() && m.IsActive()
}

// GetMembership returns the operator's membership in the specified organization.
func (o *Operator) GetMembership(orgID string) *organization.OrganizationMember {
	for _, m := range o.Memberships {
		if m.OrganizationID == orgID && m.IsActive() {
			return m
		}
	}
	return nil
}

// CanManageOperatorsIn returns true if the operator can manage other operators in the specified organization.
func (o *Operator) CanManageOperatorsIn(orgID string) bool {
	m := o.GetMembership(orgID)
	if m == nil {
		return false
	}
	return m.Role.IsAdmin()
}

// CanManageDevicesIn returns true if the operator can manage devices in the specified organization.
func (o *Operator) CanManageDevicesIn(orgID string) bool {
	m := o.GetMembership(orgID)
	if m == nil {
		return false
	}
	return m.Role.CanManageDevices()
}

// CanManageAPIKeysIn returns true if the operator can manage API keys in the specified organization.
func (o *Operator) CanManageAPIKeysIn(orgID string) bool {
	m := o.GetMembership(orgID)
	if m == nil {
		return false
	}
	return m.Role.CanManageAPIKeys()
}

// CanViewLogsIn returns true if the operator can view logs in the specified organization.
func (o *Operator) CanViewLogsIn(orgID string) bool {
	return o.GetMembership(orgID) != nil
}

// CanManageSettingsIn returns true if the operator can manage settings in the specified organization.
func (o *Operator) CanManageSettingsIn(orgID string) bool {
	m := o.GetMembership(orgID)
	if m == nil {
		return false
	}
	return m.Role.Level() >= organization.LevelAdmin
}

// IsSuperAdmin returns true if the operator has super_admin role in any organization.
func (o *Operator) IsSuperAdmin() bool {
	for _, m := range o.Memberships {
		if m.Role.IsSuperAdmin() && m.IsActive() {
			return true
		}
	}
	return false
}

// IsAdmin returns true if the operator has admin or super_admin role in any organization.
func (o *Operator) IsAdmin() bool {
	for _, m := range o.Memberships {
		if m.Role.IsAdmin() && m.IsActive() {
			return true
		}
	}
	return false
}

// IsOperator returns true if the operator has operator role in any organization.
func (o *Operator) IsOperator() bool {
	for _, m := range o.Memberships {
		if m.Role.Level() == organization.LevelOperator && m.IsActive() {
			return true
		}
	}
	return false
}

// IsViewer returns true if the operator has viewer role in any organization.
func (o *Operator) IsViewer() bool {
	for _, m := range o.Memberships {
		if m.Role.Level() == organization.LevelViewer && m.IsActive() {
			return true
		}
	}
	return false
}

// HasPermission returns true if the operator has the given permission in any
// organization. It evaluates the scoped permission engine (role defaults merged
// with custom scoped grants); the legacy coarse Permission string is mapped to
// the scoped action+scope form.
func (o *Operator) HasPermission(perm Permission) bool {
	action, scope := scopedFromLegacy(perm)
	for _, m := range o.Memberships {
		if !m.IsActive() {
			continue
		}
		defaults := permission.DefaultScopesForRole(string(m.Role))
		// Custom per-resource scopes are unioned on top of role defaults by the
		// authorization layer's Evaluator (wired grants repo); this defaults-only
		// check covers the common coarse case.
		if defaults.Grants(permission.Action(action), scope) {
			return true
		}
	}
	return false
}

// HasAnyPermission returns true if the operator has any of the given permissions.
func (o *Operator) HasAnyPermission(perms ...Permission) bool {
	for _, perm := range perms {
		if o.HasPermission(perm) {
			return true
		}
	}
	return false
}

// HasAllPermissions returns true if the operator has all of the given permissions.
func (o *Operator) HasAllPermissions(perms ...Permission) bool {
	for _, perm := range perms {
		if !o.HasPermission(perm) {
			return false
		}
	}
	return true
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
