package telemetry

import (
	"context"
)

// Repository defines the interface for telemetry data persistence.
type Repository interface {
	// Save saves a telemetry frame for a device.
	// Implements auto-pruning to keep only the latest 5000 entries.
	Save(ctx context.Context, deviceID string, raw []byte, frame TelemetryFrame) error

	// List retrieves telemetry entries for a device with pagination.
	List(ctx context.Context, deviceID string, limit int) ([]TelemetryEntry, error)

	// ListSince retrieves telemetry entries since a given timestamp.
	ListSince(ctx context.Context, deviceID string, sinceTimestamp int64, limit int) ([]TelemetryEntry, error)

	// Count returns the number of telemetry entries for a device.
	Count(ctx context.Context, deviceID string) (int, error)

	// DeleteOlderThan removes telemetry entries older than the given timestamp.
	DeleteOlderThan(ctx context.Context, olderThanTimestamp int64) (int64, error)

	// DeleteByDeviceIDs deletes all telemetry entries for the given device IDs.
	// This is used during organization deletion.
	DeleteByDeviceIDs(ctx context.Context, deviceIDs []string) (int64, error)
}
