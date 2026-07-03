// Package update provides domain models for the update system.
// This is a compatibility package that re-exports from the updates package.
package update

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/updates"
)

// Repository defines the interface for update data access.
type Repository = updates.Repository

// Ensure Repository is implemented
var _ Repository = (*updates.Repository)(nil)

// Errors provides update-specific errors.
var Errors = updates.Errors

// ErrVersionNotFound = updates.ErrVersionNotFound
var ErrVersionNotFound = updates.ErrVersionNotFound

// ErrPushNotFound = updates.ErrPushNotFound
var ErrPushNotFound = updates.ErrPushNotFound

// ErrPushNotCancellable = updates.ErrPushNotCancellable
var ErrPushNotCancellable = updates.ErrPushNotCancellable

// ErrSyncInProgress = updates.ErrSyncInProgress
var ErrSyncInProgress = updates.ErrSyncInProgress

// ErrInvalidVersion = updates.ErrInvalidVersion
var ErrInvalidVersion = updates.ErrInvalidVersion

// ErrNoDevicesSpecified = updates.ErrNoDevicesSpecified
var ErrNoDevicesSpecified = updates.ErrNoDevicesSpecified

// ErrDeviceNotFound = updates.ErrDeviceNotFound
var ErrDeviceNotFound = updates.ErrDeviceNotFound
