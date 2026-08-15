// Package event provides domain types for real-time events.
package event

import "time"

// EventType represents the type of real-time event.
type EventType string

const (
	// Device connection events.
	EventTypeDeviceConnected    EventType = "DEVICE_CONNECTED"
	EventTypeDeviceDisconnected EventType = "DEVICE_DISCONNECTED"
	EventTypeDeviceReconnected  EventType = "DEVICE_RECONNECTED"

	// Telemetry events.
	EventTypeTelemetryReceived EventType = "TELEMETRY_RECEIVED"
	EventTypeThresholdBreach   EventType = "THRESHOLD_BREACH"
	EventTypeRiskScoreAlert    EventType = "RISK_SCORE_ALERT"
	EventTypeThermalAlert      EventType = "THERMAL_ALERT"
	EventTypeBufferLevelAlert  EventType = "BUFFER_LEVEL_ALERT"

	EventTypeResolved EventType = "THRESHOLD_RESOLVED"

	// Command events.
	EventTypeCommandSent         EventType = "COMMAND_SENT"
	EventTypeCommandDelivered    EventType = "COMMAND_DELIVERED"
	EventTypeCommandFailed       EventType = "COMMAND_FAILED"
	EventTypeCommandAcknowledged EventType = "COMMAND_ACKNOWLEDGED"

	// System events.
	EventTypeFCMFallback    EventType = "FCM_FALLBACK"
	EventTypeDeviceOffline  EventType = "DEVICE_OFFLINE"
	EventTypeDeviceOnline   EventType = "DEVICE_ONLINE"
	EventTypeError          EventType = "ERROR"
	EventTypeRegistration   EventType = "REGISTRATION"
	EventTypeDeregistration EventType = "DEREGISTRATION"
)

// Severity represents the severity level of an event.
type Severity string

const (
	SeverityInfo     Severity = "info"
	SeverityWarning  Severity = "warning"
	SeverityCritical Severity = "critical"
)

// Event represents a real-time event in the system.
type Event struct {
	ID         string         `json:"id"`
	DeviceID   string         `json:"deviceId"`
	OperatorID string         `json:"operatorId,omitempty"`
	Type       EventType      `json:"type"`
	Severity   Severity       `json:"severity"`
	Timestamp  time.Time      `json:"timestamp"`
	Data       map[string]any `json:"data,omitempty"`
	Source     string         `json:"source"` // "device", "server", "dashboard".
}

// EventFilter contains filter parameters for querying events.
type EventFilter struct {
	StartTime  time.Time
	EndTime    time.Time
	DeviceIDs  []string
	EventTypes []EventType
	Severities []Severity
	Limit      int
	Offset     int
}

// EventResult contains the result of an event query.
type EventResult struct {
	Events     []Event `json:"events"`
	TotalCount int     `json:"totalCount"`
	HasMore    bool    `json:"hasMore"`
}

// GetSeverity returns the appropriate severity for an event type.
func GetSeverity(eventType EventType) Severity {
	switch eventType {
	case EventTypeThresholdBreach, EventTypeRiskScoreAlert, EventTypeThermalAlert,
		EventTypeCommandFailed, EventTypeFCMFallback, EventTypeDeviceOffline,
		EventTypeError, EventTypeDeregistration:
		return SeverityCritical
	case EventTypeBufferLevelAlert, EventTypeDeviceDisconnected:
		return SeverityWarning
	default:
		return SeverityInfo
	}
}

// IsConnectivityEvent returns true if the event is related to device connectivity.
func IsConnectivityEvent(eventType EventType) bool {
	switch eventType {
	case EventTypeDeviceConnected, EventTypeDeviceDisconnected, EventTypeDeviceReconnected,
		EventTypeDeviceOnline, EventTypeDeviceOffline:
		return true
	default:
		return false
	}
}

// IsTelemetryEvent returns true if the event is related to telemetry.
func IsTelemetryEvent(eventType EventType) bool {
	switch eventType {
	case EventTypeTelemetryReceived, EventTypeThresholdBreach, EventTypeRiskScoreAlert,
		EventTypeThermalAlert, EventTypeBufferLevelAlert, EventTypeResolved:
		return true
	default:
		return false
	}
}

// IsCommandEvent returns true if the event is related to commands.
func IsCommandEvent(eventType EventType) bool {
	switch eventType {
	case EventTypeCommandSent, EventTypeCommandDelivered, EventTypeCommandFailed,
		EventTypeCommandAcknowledged:
		return true
	default:
		return false
	}
}
