package device

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a device is not found.
var ErrNotFound = errors.New("device not found")

// Device represents a registered device.
type Device struct {
	ID                string
	FirebaseInstallID string
	FCMToken         string
	AppVersion       string
	DeviceClass      string

	// Command signing.
	CommandSecretHash string

	// Status.
	Online       bool
	RegisteredAt int64 // Unix milliseconds
	LastSeen     int64 // Unix milliseconds

	// Ownership.
	OperatorID string

	// Metadata.
	Metadata map[string]string

	// Timestamps.
	CreatedAt time.Time
	UpdatedAt time.Time
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

// DeviceClassPhone returns true if the device is a phone.
func (d *Device) IsPhone() bool {
	return d.DeviceClass == "phone"
}

// DeviceClassTablet returns true if the device is a tablet.
func (d *Device) IsTablet() bool {
	return d.DeviceClass == "tablet"
}
