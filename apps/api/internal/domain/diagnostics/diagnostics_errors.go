// Package diagnostics provides domain errors for device diagnostics.
package diagnostics

import "errors"

// Domain errors for diagnostics operations.
var (
	// ErrDeviceNotFound indicates the device does not exist.
	ErrDeviceNotFound = errors.New("device not found")
	
	// ErrInvalidIMEI indicates the IMEI format is invalid.
	ErrInvalidIMEI = errors.New("invalid IMEI format")
	
	// ErrInvalidCursor indicates the pagination cursor is invalid.
	ErrInvalidCursor = errors.New("invalid pagination cursor")
	
	// ErrInvalidTimeRange indicates the time range is invalid.
	ErrInvalidTimeRange = errors.New("invalid time range")
	
	// ErrEventNotFound indicates the event does not exist.
	ErrEventNotFound = errors.New("event not found")
	
	// ErrNoTelemetryData indicates no telemetry data exists for the device.
	ErrNoTelemetryData = errors.New("no telemetry data found")
	
	// ErrUnauthorized indicates the operator is not authorized.
	ErrUnauthorized = errors.New("unauthorized access")
)
