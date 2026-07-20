package inbox

import (
	"context"
)

// Repository defines the interface for inbox data access.
type Repository interface {
	// Create creates a new inbox entry.
	Create(ctx context.Context, e *InboxEntry) error

	// GetByID retrieves an inbox entry by ID.
	GetByID(ctx context.Context, id string) (*InboxEntry, error)

	// GetByIMEI retrieves an inbox entry by IMEI.
	GetByIMEI(ctx context.Context, imei string) (*InboxEntry, error)

	// ListByOperator retrieves paginated inbox entries for a specific operator within an organization with optional status filter.
	ListByOperator(ctx context.Context, operatorID, orgID, status string, limit, offset int) ([]*InboxEntry, int, error)

	// Update updates an existing inbox entry.
	Update(ctx context.Context, e *InboxEntry) error

	// Delete deletes an inbox entry by ID.
	Delete(ctx context.Context, id string) error

	// DeleteByIMEI deletes all inbox entries for a given IMEI.
	// Used when device re-registers to clean up stale entries.
	DeleteByIMEI(ctx context.Context, imei string) error

	// ExistsByIMEI checks if an inbox entry exists for the given IMEI.
	ExistsByIMEI(ctx context.Context, imei string) (bool, error)

	// ExistsByFirebaseInstallID checks if an inbox entry exists for the given Firebase install ID.
	ExistsByFirebaseInstallID(ctx context.Context, firebaseInstallID string) (bool, error)

	// Count returns the total number of inbox entries with optional status filter.
	Count(ctx context.Context, status string) (int, error)
}

// RegistrationLogRepository defines the interface for registration audit log data access.
type RegistrationLogRepository interface {
	// Create creates a new registration log entry.
	Create(ctx context.Context, log *RegistrationLog) error

	// ListByDeviceID retrieves all registration logs for a device.
	ListByDeviceID(ctx context.Context, deviceID string, limit, offset int) ([]*RegistrationLog, int, error)

	// ListByIMEI retrieves all registration logs for an IMEI.
	ListByIMEI(ctx context.Context, imei string, limit, offset int) ([]*RegistrationLog, int, error)

	// ListByOperator retrieves all registration logs for an operator.
	ListByOperator(ctx context.Context, operatorID string, limit, offset int) ([]*RegistrationLog, int, error)

	// CountByOperator returns the number of registration logs for an operator.
	CountByOperator(ctx context.Context, operatorID string) (int, error)
}
