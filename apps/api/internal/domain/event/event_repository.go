// Package event provides the repository interface for events.
package event

import (
	"context"
	"time"
)

// Repository defines the interface for event storage operations.
type Repository interface {
	// Store saves an event to the repository.
	Store(ctx context.Context, evt *Event) error

	// StoreBatch saves multiple events in a single transaction.
	StoreBatch(ctx context.Context, events []*Event) error

	// GetByID retrieves an event by its ID.
	GetByID(ctx context.Context, id string) (*Event, error)

	// GetByDevice retrieves events for a specific device.
	GetByDevice(ctx context.Context, deviceID string, filter *EventFilter) (*EventResult, error)

	// GetByDevices retrieves events for multiple devices.
	GetByDevices(ctx context.Context, deviceIDs []string, filter *EventFilter) (*EventResult, error)

	// GetByType retrieves events of a specific type.
	GetByType(ctx context.Context, eventType EventType, filter *EventFilter) (*EventResult, error)

	// GetByOperator retrieves events for devices owned by an operator.
	GetByOperator(ctx context.Context, operatorID string, filter *EventFilter) (*EventResult, error)

	// GetRecent retrieves the most recent events.
	GetRecent(ctx context.Context, limit int) ([]Event, error)

	// GetRecentByDevice retrieves recent events for a specific device.
	GetRecentByDevice(ctx context.Context, deviceID string, limit int) ([]Event, error)

	// CountByType counts events by type within a time range.
	CountByType(ctx context.Context, eventType EventType, startTime, endTime time.Time) (int, error)

	// DeleteOld removes events older than the specified time.
	DeleteOld(ctx context.Context, olderThan time.Time) (int64, error)
}
