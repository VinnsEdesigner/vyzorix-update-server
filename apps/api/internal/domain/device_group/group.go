// Package device_group contains the device-group (team) entity and repository.
// Devices are partitioned into groups within an org; a member of a group can
// access the group's devices (Issue 5: teams / device groups).
package device_group

import (
	"context"
	"errors"
	"time"
)

// ErrNotFound is returned when a device group is not found.
var ErrNotFound = errors.New("device group not found")

// Group is a named partition of devices within an organization.
type Group struct {
	CreatedAt time.Time
	ID        string
	OrgID     string
	Name      string
}

// Repository persists device groups, their members, and the devices in them.
type Repository interface {
	// Save upserts a group.
	Save(ctx context.Context, g *Group) error
	// GetByID returns a group or ErrNotFound.
	GetByID(ctx context.Context, id string) (*Group, error)
	// AddMember adds an operator to a group (idempotent).
	AddMember(ctx context.Context, groupID, operatorID string) error
	// RemoveMember removes an operator from a group, returning whether it existed.
	RemoveMember(ctx context.Context, groupID, operatorID string) (bool, error)
	// IsMember reports whether the operator belongs to the group.
	IsMember(ctx context.Context, groupID, operatorID string) (bool, error)
	// AddDevice assigns a device to a group (idempotent).
	AddDevice(ctx context.Context, groupID, deviceID string) error
	// RemoveDevice unassigns a device from a group, returning whether it existed.
	RemoveDevice(ctx context.Context, groupID, deviceID string) (bool, error)
	// GroupIDsForDevice returns the group IDs a device belongs to.
	GroupIDsForDevice(ctx context.Context, deviceID string) ([]string, error)
}
