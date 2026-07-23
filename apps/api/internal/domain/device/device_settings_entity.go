package device

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	orgdomain "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
)

// ErrSettingsNotFound is returned when device settings are not found.
var ErrSettingsNotFound = errors.New("device settings not found")

// ErrInvalidThreshold is returned when threshold values are invalid.
var ErrInvalidThreshold = errors.New("invalid threshold values")

// generateDeviceID generates a unique ID with a prefix.
func generateDeviceID(prefix string) string {
	bytes := make([]byte, 8)
	rand.Read(bytes)
	return prefix + hex.EncodeToString(bytes)
}

// Thresholds represents threshold values for alerts at the device level.
// These override organization-level defaults when set.
type Thresholds struct {
	RiskWarn    int `json:"riskWarn,omitempty"`
	RiskCrit    int `json:"riskCrit,omitempty"`
	ThermalWarn int `json:"thermalWarn,omitempty"`
	ThermalCrit int `json:"thermalCrit,omitempty"`
	BufferWarn  int `json:"bufferWarn,omitempty"`
	BufferCrit  int `json:"bufferCrit,omitempty"`
}

// HasValues returns true if any threshold value is set.
func (t *Thresholds) HasValues() bool {
	return t != nil && (t.RiskWarn != 0 || t.RiskCrit != 0 || t.ThermalWarn != 0 || t.ThermalCrit != 0 || t.BufferWarn != 0 || t.BufferCrit != 0)
}

// Validate validates the thresholds.

func (t *Thresholds) Validate() error {
	
	// Risk scores should be 0-100
	if t.RiskWarn < 0 || t.RiskWarn > 100 {
		return ErrInvalidThreshold
	}
	if t.RiskCrit < 0 || t.RiskCrit > 100 {
		return ErrInvalidThreshold
	}
	// Thermal thresholds should be reasonable (0-200 degrees Celsius)
	if t.ThermalWarn < 0 || t.ThermalWarn > 200 {
		return ErrInvalidThreshold
	}
	if t.ThermalCrit < 0 || t.ThermalCrit > 200 {
		return ErrInvalidThreshold
	}
	// Buffer levels should be 0-100 (percentage)
	if t.BufferWarn < 0 || t.BufferWarn > 100 {
		return ErrInvalidThreshold
	}
	if t.BufferCrit < 0 || t.BufferCrit > 100 {
		return ErrInvalidThreshold
	}

	// Relative validation: warning must be less than critical for risk/thermal,
	// but for buffer, critical (low) must be less than warning (low)
	if t.RiskWarn >= t.RiskCrit && t.RiskWarn != 0 && t.RiskCrit != 0 {
		return ErrInvalidThreshold
	}
	if t.ThermalWarn >= t.ThermalCrit && t.ThermalWarn != 0 && t.ThermalCrit != 0 {
		return ErrInvalidThreshold
	}
	if t.BufferCrit >= t.BufferWarn && t.BufferCrit != 0 && t.BufferWarn != 0 {
		return ErrInvalidThreshold
	}
	return nil
}

// DeviceSettings represents settings for a specific device.
// These settings control the Android app behavior and can override organization defaults.
type DeviceSettings struct {
	ID         string     `json:"id"`
	DeviceIMEI string     `json:"deviceImei"`
	CustomName string     `json:"customName,omitempty"`
	Location   string     `json:"location,omitempty"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	Thresholds *Thresholds `json:"thresholds,omitempty"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  time.Time  `json:"updatedAt"`
}

// NewDeviceSettings creates a new DeviceSettings.
func NewDeviceSettings(deviceIMEI string) *DeviceSettings {
	now := time.Now()
	return &DeviceSettings{
		ID:         generateDeviceID("dev_set_"),
		DeviceIMEI: deviceIMEI,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}

// IsValid validates the device settings.
func (s *DeviceSettings) IsValid() bool {
	return s.DeviceIMEI != ""
}

// HasThresholds returns true if device-specific thresholds are set.
func (s *DeviceSettings) HasThresholds() bool {
	return s.Thresholds != nil && s.Thresholds.HasValues()
}

// UpdateThresholds updates the device-specific thresholds.
func (s *DeviceSettings) UpdateThresholds(t *Thresholds) error {
	if err := t.Validate(); err != nil {
		return err
	}
	s.Thresholds = t
	s.UpdatedAt = time.Now()
	return nil
}

// UpdateDeviceSettingsRequest represents a request to update device settings.
type UpdateDeviceSettingsRequest struct {
	CustomName *string            `json:"customName,omitempty"`
	Location   *string            `json:"location,omitempty"`
	Metadata   map[string]string  `json:"metadata,omitempty"`
	Thresholds *Thresholds        `json:"thresholds,omitempty"`
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

// ResolveThresholds resolves the effective thresholds using the hierarchy:
// device settings → organization settings → default thresholds
func ResolveThresholds(deviceSettings *DeviceSettings, orgThresholds *Thresholds) *Thresholds {
	// Start with organization thresholds or defaults
	result := &Thresholds{
		RiskWarn:    70,
		RiskCrit:    90,
		ThermalWarn: 75,
		ThermalCrit: 85,
		BufferWarn:  30,
		BufferCrit:  10,
	}

	// Apply organization thresholds if provided
	if orgThresholds != nil {
		if orgThresholds.RiskWarn != 0 {
			result.RiskWarn = orgThresholds.RiskWarn
		}
		if orgThresholds.RiskCrit != 0 {
			result.RiskCrit = orgThresholds.RiskCrit
		}
		if orgThresholds.ThermalWarn != 0 {
			result.ThermalWarn = orgThresholds.ThermalWarn
		}
		if orgThresholds.ThermalCrit != 0 {
			result.ThermalCrit = orgThresholds.ThermalCrit
		}
		if orgThresholds.BufferWarn != 0 {
			result.BufferWarn = orgThresholds.BufferWarn
		}
		if orgThresholds.BufferCrit != 0 {
			result.BufferCrit = orgThresholds.BufferCrit
		}
	}

	// Apply device-specific thresholds if set (they override org defaults)
	if deviceSettings != nil && deviceSettings.HasThresholds() {
		if deviceSettings.Thresholds.RiskWarn != 0 {
			result.RiskWarn = deviceSettings.Thresholds.RiskWarn
		}
		if deviceSettings.Thresholds.RiskCrit != 0 {
			result.RiskCrit = deviceSettings.Thresholds.RiskCrit
		}
		if deviceSettings.Thresholds.ThermalWarn != 0 {
			result.ThermalWarn = deviceSettings.Thresholds.ThermalWarn
		}
		if deviceSettings.Thresholds.ThermalCrit != 0 {
			result.ThermalCrit = deviceSettings.Thresholds.ThermalCrit
		}
		if deviceSettings.Thresholds.BufferWarn != 0 {
			result.BufferWarn = deviceSettings.Thresholds.BufferWarn
		}
		if deviceSettings.Thresholds.BufferCrit != 0 {
			result.BufferCrit = deviceSettings.Thresholds.BufferCrit
		}
	}

	return result
}

// FromOrgThresholds converts organization thresholds to device thresholds.
func FromOrgThresholds(org *orgdomain.Thresholds) *Thresholds {
	if org == nil {
		return nil
	}
	return &Thresholds{
		RiskWarn:    org.RiskWarn,
		RiskCrit:    org.RiskCrit,
		ThermalWarn: org.ThermalWarn,
		ThermalCrit: org.ThermalCrit,
		BufferWarn:  org.BufferWarn,
		BufferCrit:  org.BufferCrit,
	}
}
