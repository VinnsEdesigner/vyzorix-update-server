package device

import "context"

// DeviceSettingsRepository defines the interface for device settings data access.
type DeviceSettingsRepository interface {
	// Create creates new device settings.
	Create(ctx context.Context, settings *DeviceSettings) error

	// FindByID retrieves settings by ID.
	FindByID(ctx context.Context, id string) (*DeviceSettings, error)

	// FindByDeviceIMEI retrieves settings by device IMEI.
	FindByDeviceIMEI(ctx context.Context, imei string) (*DeviceSettings, error)

	// Update updates device settings.
	Update(ctx context.Context, settings *DeviceSettings) error

	// Delete deletes device settings.
	Delete(ctx context.Context, id string) error

	// DeleteByDeviceIMEI deletes settings by device IMEI.
	DeleteByDeviceIMEI(ctx context.Context, imei string) error
}
