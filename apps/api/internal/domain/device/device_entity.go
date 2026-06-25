package device

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a device is not found.
var ErrNotFound = errors.New("device not found")

// Device represents a registered device.
type Device struct {
	UpdatedAt         time.Time
	CreatedAt         time.Time
	Metadata          map[string]string
	OperatorID        string
	DeviceClass       string
	CommandSecretHash string
	ID                string
	AppVersion        string
	FCMToken          string
	FirebaseInstallID string
	RegisteredAt      int64
	LastSeen          int64
	Online            bool
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
