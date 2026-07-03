// Package diagnostics provides repository interface for device diagnostics.
package diagnostics

import "context"

// Repository defines the interface for diagnostics data access.
type Repository interface {
	// GetTimelineEvents retrieves paginated timeline events for a device.
	GetTimelineEvents(ctx context.Context, deviceID string, filter *TimelineFilter) (*TimelineResult, error)
	
	// RecordEvent records a new device event.
	RecordEvent(ctx context.Context, event *TimelineEvent) error
	
	// GetTelemetryStats retrieves telemetry statistics for a device.
	GetTelemetryStats(ctx context.Context, deviceID string) (*TelemetryInfo, error)
	
	// GetLastTelemetry retrieves the most recent telemetry data for a device.
	// Returns ErrNoTelemetryData if no telemetry exists for the device.
	GetLastTelemetry(ctx context.Context, deviceID string) (*TimelineEvent, error)
}
