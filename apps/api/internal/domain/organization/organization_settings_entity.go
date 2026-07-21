package organization

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// generateID generates a unique ID with a prefix.
func generateID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return prefix + hex.EncodeToString(bytes)
}

// OrganizationSettings represents settings for an organization.
type OrganizationSettings struct {
	ID                     string         `json:"id"`
	OrganizationID         string         `json:"organizationId"`
	Timezone              string         `json:"timezone"`
	DateFormat            string         `json:"dateFormat"`
	AlertCooldownMinutes  int            `json:"alertCooldownMinutes"`
	DefaultThresholds     *Thresholds    `json:"defaultThresholds,omitempty"`
	CreatedAt             time.Time      `json:"createdAt"`
	UpdatedAt             time.Time      `json:"updatedAt"`
}

// Thresholds represents threshold values for alerts.
type Thresholds struct {
	RiskWarn    int `json:"riskWarn"`
	RiskCrit    int `json:"riskCrit"`
	ThermalWarn int `json:"thermalWarn"`
	ThermalCrit int `json:"thermalCrit"`
	BufferWarn  int `json:"bufferWarn"`
	BufferCrit  int `json:"bufferCrit"`
}

// DefaultThresholds returns default threshold values.
func DefaultThresholds() *Thresholds {
	return &Thresholds{
		RiskWarn:    70,
		RiskCrit:    90,
		ThermalWarn: 75,
		ThermalCrit: 85,
		BufferWarn:  30,
		BufferCrit:  10,
	}
}

// NewOrganizationSettings creates a new OrganizationSettings with defaults.
func NewOrganizationSettings(organizationID string) *OrganizationSettings {
	now := time.Now()
	return &OrganizationSettings{
		ID:                    generateID("org_set_"),
		OrganizationID:        organizationID,
		Timezone:             "UTC",
		DateFormat:           "YYYY-MM-DD",
		AlertCooldownMinutes: 15,
		DefaultThresholds:     DefaultThresholds(),
		CreatedAt:            now,
		UpdatedAt:            now,
	}
}

// IsValid validates the organization settings.
func (s *OrganizationSettings) IsValid() bool {
	return s.OrganizationID != ""
}

// UpdateThresholds updates the default thresholds.
func (s *OrganizationSettings) UpdateThresholds(t *Thresholds) error {
	if err := t.Validate(); err != nil {
		return err
	}
	s.DefaultThresholds = t
	s.UpdatedAt = time.Now()
	return nil
}

// Validate validates the thresholds.
func (t *Thresholds) Validate() error {
	if t.RiskWarn <= 0 || t.RiskCrit <= 0 {
		return ErrInvalidThreshold
	}
	if t.ThermalWarn <= 0 || t.ThermalCrit <= 0 {
		return ErrInvalidThreshold
	}
	if t.BufferWarn <= 0 || t.BufferCrit <= 0 {
		return ErrInvalidThreshold
	}
	// Ensure warn values are less than crit values
	if t.RiskWarn >= t.RiskCrit {
		return ErrInvalidThreshold
	}
	if t.ThermalWarn >= t.ThermalCrit {
		return ErrInvalidThreshold
	}
	if t.BufferCrit >= t.BufferWarn {
		return ErrInvalidThreshold
	}
	return nil
}

// UpdateOrganizationSettingsRequest represents a request to update organization settings.
type UpdateOrganizationSettingsRequest struct {
	Timezone             *string    `json:"timezone,omitempty"`
	DateFormat           *string    `json:"dateFormat,omitempty"`
	AlertCooldownMinutes *int       `json:"alertCooldownMinutes,omitempty"`
	DefaultThresholds    *Thresholds `json:"defaultThresholds,omitempty"`
}

// UpdateThresholdsRequest represents a request to update only thresholds.
type UpdateThresholdsRequest struct {
	RiskWarn    *int `json:"riskWarn,omitempty"`
	RiskCrit    *int `json:"riskCrit,omitempty"`
	ThermalWarn *int `json:"thermalWarn,omitempty"`
	ThermalCrit *int `json:"thermalCrit,omitempty"`
	BufferWarn  *int `json:"bufferWarn,omitempty"`
	BufferCrit  *int `json:"bufferCrit,omitempty"`
}

// ToThresholds converts to Thresholds.
func (r *UpdateThresholdsRequest) ToThresholds() *Thresholds {
	if r == nil {
		return nil
	}
	t := &Thresholds{}
	if r.RiskWarn != nil {
		t.RiskWarn = *r.RiskWarn
	}
	if r.RiskCrit != nil {
		t.RiskCrit = *r.RiskCrit
	}
	if r.ThermalWarn != nil {
		t.ThermalWarn = *r.ThermalWarn
	}
	if r.ThermalCrit != nil {
		t.ThermalCrit = *r.ThermalCrit
	}
	if r.BufferWarn != nil {
		t.BufferWarn = *r.BufferWarn
	}
	if r.BufferCrit != nil {
		t.BufferCrit = *r.BufferCrit
	}
	return t
}
