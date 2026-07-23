package command

import (
	"context"
	"time"
)

// Repository defines the interface for command data access.
type Repository interface {
	// FindByID retrieves a command by ID.
	FindByID(ctx context.Context, id string) (*Command, error)

	// FindByDispatchID retrieves a command by dispatch ID (for idempotency).
	FindByDispatchID(ctx context.Context, deviceID, dispatchID string) (*Command, error)

	// FindByDispatchIDOnly retrieves a command by dispatch ID only (dispatch ID should be globally unique).
	FindByDispatchIDOnly(ctx context.Context, dispatchID string) (*Command, error)

	// FindByDeviceID retrieves commands for a device.
	FindByDeviceID(ctx context.Context, deviceID string, limit int) ([]*Command, error)

	// FindPendingByDeviceID retrieves pending commands for a device.
	FindPendingByDeviceID(ctx context.Context, deviceID string) ([]*Command, error)

	// FindHistoryByDeviceID retrieves paginated command history for a device with time range filtering.
	FindHistoryByDeviceID(ctx context.Context, deviceID string, status string, startTime, endTime time.Time, limit, offset int) ([]*Command, int, error)

	// Create creates a new command.
	Create(ctx context.Context, cmd *Command) error

	// Update updates a command.
	Update(ctx context.Context, cmd *Command) error

	// UpdateStatus updates the status of a command.
	UpdateStatus(ctx context.Context, id string, status Status) error

	// Delete deletes a command.
	Delete(ctx context.Context, id string) error

	// DeleteByDeviceID deletes all commands for a device.
	DeleteByDeviceID(ctx context.Context, deviceID string) error

	// Count returns the total number of commands.
	Count(ctx context.Context) (int, error)

	// CountPending returns the number of pending commands.
	CountPending(ctx context.Context) (int, error)
	CountPendingByDevice(ctx context.Context, deviceID string) (int, error)

	// MarkWake marks whether a wake command was sent successfully for a command dispatch.
	MarkWake(ctx context.Context, dispatchID string, errText string) error

	// MarkDelivered marks a command as delivered by dispatch ID.
	MarkDelivered(ctx context.Context, dispatchID string) error

	// MarkCompleted marks a command as completed by dispatch ID with result.
	MarkCompleted(ctx context.Context, dispatchID, result string) error

	// MarkFailed marks a command as failed by dispatch ID with error message.
	MarkFailed(ctx context.Context, dispatchID, errMsg string) error

	// DeleteOldCommands removes commands older than the given timestamp.
	DeleteOldCommands(ctx context.Context, olderThan int64) (int64, error)

	// FindByDispatchPrefix retrieves all commands whose dispatch_id starts with the given prefix.
	// Used by CancelPush to cancel all commands for a specific push ID.
	FindByDispatchPrefix(ctx context.Context, prefix string) ([]*Command, error)

	// CancelByDispatchPrefix marks all pending commands whose dispatch_id starts with the given prefix as cancelled.
	CancelByDispatchPrefix(ctx context.Context, prefix string) (int64, error)

	// FindPending retrieves pending commands for the outbox worker.
	// Returns commands with status=pending where next_retry_at is null or in the past,.
	// ordered by creation time (oldest first).
	FindPending(ctx context.Context, limit int) ([]*Command, error)

	// UpdateRetryInfo updates the retry tracking fields for a command.
	// Used by the outbox worker to track delivery attempts with exponential backoff.
	UpdateRetryInfo(ctx context.Context, id string, retryCount int, maxRetries int, nextRetryAt *time.Time) error

	// DeleteByDeviceIDs deletes all commands for the given device IDs.
	// This is used during organization deletion.
	DeleteByDeviceIDs(ctx context.Context, deviceIDs []string) (int64, error)
}
