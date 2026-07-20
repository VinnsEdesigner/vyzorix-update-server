package device

import (
	"context"
)

// Repository defines the interface for device data access.
type Repository interface {
	// FindByID retrieves a device by ID.
	FindByID(ctx context.Context, id string) (*Device, error)

	// FindByDeviceID retrieves a device by DeviceID value object.
	FindByDeviceID(ctx context.Context, id DeviceID) (*Device, error)

	// FindByIMEI retrieves a device by IMEI (device ID).
	FindByIMEI(ctx context.Context, imei string) (*Device, error)

	// FindByFirebaseInstallID retrieves a device by Firebase install ID.
	FindByFirebaseInstallID(ctx context.Context, fid string) (*Device, error)

	// FindByIDAndOperator retrieves a device by ID and verifies ownership.
	// Returns ErrNotFound if device doesn't exist or doesn't belong to operator.
	// This is used for DOA (Data Ownership Attribution) checks.
	FindByIDAndOperator(ctx context.Context, id, operatorID string) (*Device, error)

	// FindByIDAndOperatorID retrieves a device by DeviceID and OperatorID value objects.
	FindByIDAndOperatorID(ctx context.Context, id DeviceID, operatorID OperatorID) (*Device, error)

	// FindByIMEIAndOperator retrieves a device by IMEI and verifies ownership.
	// Returns ErrNotFound if device doesn't exist or doesn't belong to operator.
	// This is used for DOA (Data Ownership Attribution) checks on deregistration.
	FindByIMEIAndOperator(ctx context.Context, imei, operatorID string) (*Device, error)

	// FindByIMEIAndOrganization retrieves a device by IMEI within an organization.
	// Returns ErrNotFound if device doesn't exist or doesn't belong to the organization.
	FindByIMEIAndOrganization(ctx context.Context, imei, orgID string) (*Device, error)

	// FindByIDAndOrganization retrieves a device by ID within an organization.
	// Returns ErrNotFound if device doesn't exist or doesn't belong to the organization.
	FindByIDAndOrganization(ctx context.Context, id, orgID string) (*Device, error)

	// Create creates a new device.
	Create(ctx context.Context, d *Device) error

	// Update updates an existing device.
	Update(ctx context.Context, d *Device) error

	// Delete deletes a device.
	Delete(ctx context.Context, id string) error

	// DeleteByDeviceID deletes a device by DeviceID.
	DeleteByDeviceID(ctx context.Context, id DeviceID) error

	// UpdateFCMToken updates the FCM token for a device.
	UpdateFCMToken(ctx context.Context, id, fcmToken string) error

	// SetOnline sets the online status of a device.
	SetOnline(ctx context.Context, id string, online bool) error

	// SetOnlineByDeviceID sets the online status by DeviceID.
	SetOnlineByDeviceID(ctx context.Context, id DeviceID, online bool) error

	// UpdateLastSeen updates the last seen timestamp.
	UpdateLastSeen(ctx context.Context, id string) error

	// Touch updates the last seen timestamp (alias for UpdateLastSeen).
	Touch(ctx context.Context, deviceID string) error

	// SetSecretHash sets the command secret hash for a device.
	SetSecretHash(ctx context.Context, deviceID, hash string) error

	// GetSecretHash retrieves the command secret hash for a device.
	GetSecretHash(ctx context.Context, deviceID string) (string, error)

	// HashAllSecrets hashes all existing command secrets that don't have a hash.
	// This is a migration helper for existing databases.
	HashAllSecrets(ctx context.Context) (int, error)

	// List returns a paginated list of devices.
	List(ctx context.Context, limit, offset int) ([]*Device, int, error)

	// ListByOperator returns all devices for an operator.
	ListByOperator(ctx context.Context, operatorID string) ([]*Device, error)

	// ListByOperatorID returns all devices for an OperatorID.
	ListByOperatorID(ctx context.Context, operatorID OperatorID) ([]*Device, error)

	// ListByOrganization returns all devices for an organization.
	ListByOrganization(ctx context.Context, orgID string) ([]*Device, error)

	// Count returns the total number of devices.
	Count(ctx context.Context) (int, error)

	// CountByOperator returns the number of devices for an operator.
	CountByOperator(ctx context.Context, operatorID string) (int, error)

	// CountByOrganization returns the number of devices for an organization.
	CountByOrganization(ctx context.Context, orgID string) (int, error)

	// SoftDelete marks a device as deregistered (soft delete).
	// Sets deregistered_at and deletion_scheduled_at for 30-day retention.
	SoftDelete(ctx context.Context, id string, deregisteredAt, deletionScheduledAt int64) error

	// SoftDeleteByIMEI marks a device as deregistered by IMEI.
	SoftDeleteByIMEI(ctx context.Context, imei string, deregisteredAt, deletionScheduledAt int64) error

	// ListActive returns all non-deregistered devices.
	ListActive(ctx context.Context, limit, offset int) ([]*Device, int, error)

	// ListActiveByOperator returns all non-deregistered devices for an operator.
	ListActiveByOperator(ctx context.Context, operatorID string) ([]*Device, error)

	// ListPending returns all devices in pending lifecycle state.
	ListPending(ctx context.Context) ([]*Device, error)

	// ListPendingByOperator returns all pending devices for an operator.
	ListPendingByOperator(ctx context.Context, operatorID OperatorID) ([]*Device, error)
}
