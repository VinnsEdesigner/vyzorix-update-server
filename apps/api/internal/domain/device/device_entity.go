package device

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a device is not found.
var ErrNotFound = errors.New("device not found")

// Device represents a registered device.
type Device struct {
	UpdatedAt             time.Time
	CreatedAt             time.Time
	Metadata              map[string]string
	OperatorID            string
	DeviceClass           string
	CommandSecretHash     string
	ID                    string
	AppVersion            string
	FCMToken              string
	FirebaseInstallID     string
	RegisteredAt          int64
	LastSeen              int64
	Online                bool
	// Additional fields per spec
	DeviceName            string // device_name column
	Manufacturer          string // manufacturer column (stored via model)
	Model                string // model column
	OSVersion            string // os_version column
	SecurityPatch        string // security_patch column
	DeregisteredAt        *int64 // deregistered_at column (UnixMilli)
	DeletionScheduledAt   *int64 // deletion_scheduled_at column (UnixMilli)
	FCMTokenRefreshedAt   *int64 // fcm_token_refreshed_at column (UnixMilli)
}

// IsOnline returns true if the device is currently online.
func (d *Device) IsOnline() bool {
	return d.Online
}

// IsStale returns true if the device hasn't been seen recently.
func (d *Device) IsStale(threshold time.Duration) bool {
	return time.Since(time.UnixMilli(d.LastSeen)) > threshold
}

// LastSeenTime returns the LastSeen as a time.Time.
func (d *Device) LastSeenTime() time.Time {
	return time.UnixMilli(d.LastSeen)
}

// RegisteredAtTime returns the RegisteredAt as a time.Time.
func (d *Device) RegisteredAtTime() time.Time {
	return time.UnixMilli(d.RegisteredAt)
}

// IsValid returns true if the device has all required fields.
func (d *Device) IsValid() bool {
	return d.ID != "" && d.FirebaseInstallID != ""
}

// IsPhone returns true if the device is a phone.
func (d *Device) IsPhone() bool {
	return d.DeviceClass == "phone"
}

// IsTablet returns true if the device is a tablet.
func (d *Device) IsTablet() bool {
	return d.DeviceClass == "tablet"
}

// IsDeregistered returns true if the device has been deregistered.
func (d *Device) IsDeregistered() bool {
	return d.DeregisteredAt != nil && *d.DeregisteredAt > 0
}

// DeregisteredAtTime returns the DeregisteredAt as a time.Time if set.
func (d *Device) DeregisteredAtTime() *time.Time {
	if d.DeregisteredAt == nil {
		return nil
	}
	t := time.UnixMilli(*d.DeregisteredAt)
	return &t
}

// DeletionScheduledAtTime returns the DeletionScheduledAt as a time.Time if set.
func (d *Device) DeletionScheduledAtTime() *time.Time {
	if d.DeletionScheduledAt == nil {
		return nil
	}
	t := time.UnixMilli(*d.DeletionScheduledAt)
	return &t
}

// FCMTokenRefreshedAtTime returns the FCMTokenRefreshedAt as a time.Time if set.
func (d *Device) FCMTokenRefreshedAtTime() *time.Time {
	if d.FCMTokenRefreshedAt == nil {
		return nil
	}
	t := time.UnixMilli(*d.FCMTokenRefreshedAt)
	return &t
}

// IsFCMTokenValid returns true if the FCM token is valid.
// Token is valid if it's non-empty and was refreshed within 30 days.
func (d *Device) IsFCMTokenValid() bool {
	if d.FCMToken == "" {
		return false
	}
	if d.FCMTokenRefreshedAt == nil || *d.FCMTokenRefreshedAt == 0 {
		// If never refreshed, check if token was set recently (within 30 days)
		// For now, consider empty refresh time as valid
		return true
	}
	refreshTime := time.UnixMilli(*d.FCMTokenRefreshedAt)
	return time.Since(refreshTime) < 30*24*time.Hour
}

// IsCommandSecretSet returns true if command secret hash is set.
func (d *Device) IsCommandSecretSet() bool {
	return d.CommandSecretHash != ""
}

// GetStatus returns the device status string.
func (d *Device) GetStatus() string {
	if d.IsDeregistered() {
		return "deregistered"
	}
	if d.Online {
		return "online"
	}
	return "offline"
}
