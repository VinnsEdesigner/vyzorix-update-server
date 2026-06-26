package logs

import (
	"context"
	"time"
)

// Repository defines the interface for device logs data access.
type Repository interface {
	// CreateLog creates a new device log entry.
	CreateLog(ctx context.Context, log *DeviceLog) error

	// GetLogByID retrieves a log entry by ID.
	GetLogByID(ctx context.Context, id string) (*DeviceLog, error)

	// ListLogs retrieves paginated device logs with cursor-based pagination.
	// Returns logs ordered by timestamp DESC, id DESC.
	ListLogs(ctx context.Context, deviceID string, eventType string, startTime, endTime time.Time, limit int, cursor string) ([]*DeviceLog, string, error)

	// CountLogs counts logs matching the criteria.
	CountLogs(ctx context.Context, deviceID string, eventType string, startTime, endTime time.Time) (int, error)
}
