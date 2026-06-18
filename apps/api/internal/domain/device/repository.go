package device

import (
	"context"
)

// Repository defines the interface for device data access.
type Repository interface {
	// FindByID retrieves a device by ID.
	FindByID(ctx context.Context, id string) (*Device, error)
	
	// FindByFirebaseInstallID retrieves a device by Firebase install ID.
	FindByFirebaseInstallID(ctx context.Context, fid string) (*Device, error)
	
	// Create creates a new device.
	Create(ctx context.Context, d *Device) error
	
	// Update updates an existing device.
	Update(ctx context.Context, d *Device) error
	
	// Delete deletes a device.
	Delete(ctx context.Context, id string) error
	
	// UpdateFCMToken updates the FCM token for a device.
	UpdateFCMToken(ctx context.Context, id, fcmToken string) error
	
	// SetOnline sets the online status of a device.
	SetOnline(ctx context.Context, id string, online bool) error
	
	// UpdateLastSeen updates the last seen timestamp.
	UpdateLastSeen(ctx context.Context, id string) error
	
	// List returns a paginated list of devices.
	List(ctx context.Context, limit, offset int) ([]*Device, int, error)
	
	// ListByOperator returns all devices for an operator.
	ListByOperator(ctx context.Context, operatorID string) ([]*Device, error)
	
	// Count returns the total number of devices.
	Count(ctx context.Context) (int, error)
	
	// CountByOperator returns the number of devices for an operator.
	CountByOperator(ctx context.Context, operatorID string) (int, error)
}
