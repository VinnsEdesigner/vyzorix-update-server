// Package devicestore defines the persistence port for known-device tracking used
// by new-device login notifications.
//
// The port lives in domain (not application or infrastructure) so both the
// in-memory implementation in application/auth and the SQL implementation in
// infrastructure/storage can satisfy it without an import cycle: application/auth
// is imported by infrastructure/storage, so a port defined in application/auth
// could not be referenced from storage.
package devicestore

import "context"

// Signal kinds recorded per device.
const (
	KindIP        = "ip"
	KindUserAgent = "ua"
)

// Store tracks which devices an operator has logged in from, so a login from an
// unrecognised device or IP can be flagged.
//
// All methods must be safe for concurrent use. Implementations are responsible for
// keeping their own storage bounded: the fingerprint is derived from request
// signals, so it is client-influenced.
type Store interface {
	// IsKnownDevice reports whether the operator has a device with this fingerprint.
	IsKnownDevice(ctx context.Context, operatorID, fingerprint string) (bool, error)

	// HasSignal reports whether the device has been seen from this value of kind
	// (KindIP or KindUserAgent).
	HasSignal(ctx context.Context, operatorID, fingerprint, kind, value string) (bool, error)

	// RegisterDevice records the device and the observed IP / User-Agent.
	RegisterDevice(ctx context.Context, operatorID, fingerprint, ipAddress, userAgent string) error

	// TouchDevice updates the device's last-seen timestamp.
	TouchDevice(ctx context.Context, operatorID, fingerprint string) error
}
