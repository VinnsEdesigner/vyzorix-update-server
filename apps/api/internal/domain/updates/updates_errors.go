package updates

import "errors"

var (
	// ErrVersionNotFound is returned when an update version is not found.
	ErrVersionNotFound = errors.New("update version not found")

	// ErrPushNotFound is returned when an update push is not found.
	ErrPushNotFound = errors.New("update push not found")

	// ErrPushNotCancellable is returned when trying to cancel a non-cancellable push.
	ErrPushNotCancellable = errors.New("push cannot be cancelled")

	// ErrSyncInProgress is returned when a sync is already in progress.
	ErrSyncInProgress = errors.New("sync already in progress")

	// ErrInvalidVersion is returned when the version format is invalid.
	ErrInvalidVersion = errors.New("invalid version format")

	// ErrNoDevicesSpecified is returned when no devices are specified for push.
	ErrNoDevicesSpecified = errors.New("no devices specified")

	// ErrDeviceNotFound is returned when a device is not found.
	ErrDeviceNotFound = errors.New("device not found")
)
