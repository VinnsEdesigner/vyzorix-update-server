package logs

import "errors"

// Domain errors for logs.
var (
	// ErrLogNotFound is returned when a log entry is not found.
	ErrLogNotFound = errors.New("log entry not found")

	// ErrInvalidEventType is returned when an invalid event type is provided.
	ErrInvalidEventType = errors.New("invalid event type")

	// ErrInvalidTimeRange is returned when the time range is invalid.
	ErrInvalidTimeRange = errors.New("invalid time range")
)
