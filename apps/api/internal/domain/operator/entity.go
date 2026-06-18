package operator

import (
	"errors"
	"strings"
	"time"
)

// ErrNotFound is returned when an operator is not found.
var ErrNotFound = errors.New("operator not found")

// Role represents the role of an operator.
type Role string

const (
	// RoleSuperAdmin is the highest privilege role.
	RoleSuperAdmin Role = "super_admin"
	// RoleAdmin is an administrative role.
	RoleAdmin Role = "admin"
	// RoleOperator is a standard operator role.
	RoleOperator Role = "operator"
)

// Thresholds define alert levels for device telemetry.
type Thresholds struct {
	RiskWarn    int `json:"riskWarn"`
	RiskCrit   int `json:"riskCrit"`
	ThermalWarn int `json:"thermalWarn"`
	ThermalCrit int `json:"thermalCrit"`
	BufferWarn  int `json:"bufferWarn"`
	BufferCrit  int `json:"bufferCrit"`
}

// ClientSettings holds operator preferences that control dashboard behavior.
type ClientSettings struct {
	StrictHmac           bool `json:"strictHmac"`
	AutoReconnect       bool `json:"autoReconnect"`
	NotificationsEnabled bool `json:"notificationsEnabled"`
}

// Operator represents a system operator (user).
type Operator struct {
	ID           string
	Email        string
	Name         string
	PasswordHash string
	Role         Role

	// OAuth fields (optional - one or both may be set).
	GoogleID string
	GitHubID string

	// MFA fields.
	MFASecret    string
	MFAEnabled   bool
	BackupCodes  []string

	// Email verification.
	EmailVerified bool

	// Settings.
	Thresholds    Thresholds    `json:"thresholds"`
	ClientSettings ClientSettings `json:"client"`

	// Timestamps.
	CreatedAt time.Time
	UpdatedAt time.Time
}

// IsSuperAdmin returns true if the operator is a super admin.
func (o *Operator) IsSuperAdmin() bool {
	return o.Role == RoleSuperAdmin
}

// IsAdmin returns true if the operator is an admin or super admin.
func (o *Operator) IsAdmin() bool {
	return o.Role == RoleSuperAdmin || o.Role == RoleAdmin
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
