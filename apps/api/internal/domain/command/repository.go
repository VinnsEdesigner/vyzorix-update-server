package command

import (
	"context"
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
}
