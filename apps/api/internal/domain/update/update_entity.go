// Package update provides domain models for the update system.
// This is a compatibility package that re-exports from the updates package.
package update

import (
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// UpdateVersion represents a version of the application that can be pushed to devices.
type UpdateVersion = updates.UpdateVersion

// UpdatePush represents an update push operation to one or more devices.
type UpdatePush = updates.UpdatePush

// UpdatePushDevice represents a device targeted by an update push.
type UpdatePushDevice = updates.UpdatePushDevice

// SyncState represents the state of the GitHub synchronization process.
type SyncState = updates.SyncState

// DeviceUpdateStatus represents the status of a device update.
type DeviceUpdateStatus = updates.DeviceUpdateStatus

const (
	// DeviceUpdateStatusPending indicates the update is pending.
	DeviceUpdateStatusPending = updates.DeviceUpdateStatusPending
	// DeviceUpdateStatusSent indicates the update was sent.
	DeviceUpdateStatusSent = updates.DeviceUpdateStatusSent
	// DeviceUpdateStatusAcknowledged indicates the device acknowledged the update.
	DeviceUpdateStatusAcknowledged = updates.DeviceUpdateStatusAcknowledged
	// DeviceUpdateStatusFailed indicates the update failed.
	DeviceUpdateStatusFailed = updates.DeviceUpdateStatusFailed
)

const (
	// InstallTypeImmediate indicates immediate installation.
	InstallTypeImmediate = updates.InstallTypeImmediate
	// InstallTypeBackground indicates background installation.
	InstallTypeBackground = updates.InstallTypeBackground
	// InstallTypeForced indicates forced installation.
	InstallTypeForced = updates.InstallTypeForced
)
