package device

import (
	"errors"
	"time"
)

// ErrNotFound is returned when a device is not found.
var ErrNotFound = errors.New("device not found")

// Device represents a registered device with explicit lifecycle management.
type Device struct {
	// Lifecycle tracks the registration lifecycle state (pending → registered → deregistered)
	Lifecycle Lifecycle

	// OrganizationID is the organization this device belongs to (for multi-tenant).
	OrganizationID string

	// Infrastructure fields (kept as-is for backward compatibility)
	UpdatedAt            time.Time
	CreatedAt            time.Time
	Metadata             map[string]string
	FCMTokenRefreshedAt  *int64
	DeletionScheduledAt  *int64
	DeregisteredAt       *int64
	Model                string
	OSVersion            string
	FCMToken             string
	FirebaseInstallID    string
	OperatorID           string
	DeviceClass          string
	CommandSecretHash    string
	DeviceName           string
	Manufacturer         string
	ID                   string
	AppVersion           string
	SecurityPatch        string
	LastSeen             int64
	RegisteredAt         int64
	Online               bool
}

// NewDevice creates a new Device with pending lifecycle.
func NewDevice(id string, firebaseInstallID string) *Device {
	return &Device{
		ID:                id,
		FirebaseInstallID: firebaseInstallID,
		Lifecycle:         LifecyclePending,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}
}

// Approve transitions the device from pending to registered.
// Returns ErrInvalidTransition if the device is not in pending state.
func (d *Device) Approve() error {
	return d.Lifecycle.TransitionTo(LifecycleRegistered)
}

// Deregister transitions the device to deregistered state.
// Returns ErrInvalidTransition if the transition is not allowed.
func (d *Device) Deregister() error {
	if err := d.Lifecycle.TransitionTo(LifecycleDeregistered); err != nil {
		return err
	}
	now := time.Now().UnixMilli()
	d.DeregisteredAt = &now
	deletionTime := now + (30 * 24 * int64(time.Hour/time.Millisecond))
	d.DeletionScheduledAt = &deletionTime
	return nil
}

// IsPending returns true if the device is waiting for approval.
func (d *Device) IsPending() bool {
	return d.Lifecycle.IsPending()
}

// IsRegistered returns true if the device is approved and active.
func (d *Device) IsRegistered() bool {
	return d.Lifecycle.IsRegistered()
}

// IsDeregistered returns true if the device has been deregistered.
func (d *Device) IsDeregistered() bool {
	return d.Lifecycle.IsDeregistered()
}

// IsActive returns true if the device can accept commands.
func (d *Device) IsActive() bool {
	return d.Lifecycle.IsActive() && d.Online
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

// GetStatus returns the device status string based on lifecycle and connection state.
func (d *Device) GetStatus() string {
	if d.Lifecycle.IsDeregistered() {
		return "deregistered"
	}
	if d.Lifecycle.IsPending() {
		return "pending"
	}
	if d.Online {
		return "online"
	}
	return "offline"
}

// TransferToOrganization transfers the device to a new organization.
// The device must be offline (not online) to be transferred.
func (d *Device) TransferToOrganization(orgID string) error {
	if d.Online {
		return errors.New("device must be offline to transfer")
	}
	d.OrganizationID = orgID
	return nil
}
