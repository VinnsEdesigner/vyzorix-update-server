package logs

import (
	"encoding/json"
	"time"
)

// DeviceLog represents a device event log entry.
type DeviceLog struct {
	ID        string          `json:"id"`
	DeviceID  string          `json:"deviceId"`
	EventType string          `json:"eventType"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data,omitempty"`
}

// EventType constants for device logs.
const (
	EventTypeConnection     = "connection"
	EventTypeCommand        = "command"
	EventTypeTelemetry      = "telemetry"
	EventTypeError          = "error"
	EventTypeWarning        = "warning"
	EventTypeInfo           = "info"
	EventTypeUpdate         = "update"
	EventTypeRegistration   = "registration"
	EventTypeDeregistration = "deregistration"
)

// IsValidEventType checks if the event type is valid.
func IsValidEventType(eventType string) bool {
	switch eventType {
	case EventTypeConnection,
		EventTypeCommand,
		EventTypeTelemetry,
		EventTypeError,
		EventTypeWarning,
		EventTypeInfo,
		EventTypeUpdate,
		EventTypeRegistration,
		EventTypeDeregistration:
		return true
	default:
		return false
	}
}
