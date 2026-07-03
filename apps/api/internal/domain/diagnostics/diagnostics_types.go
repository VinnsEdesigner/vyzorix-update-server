// Package diagnostics provides domain types for device diagnostics.
package diagnostics

import "time"

// TimelineEventType represents the type of device event.
type TimelineEventType string

const (
	EventTypeTelemetry        TimelineEventType = "TELEMETRY"
	EventTypeCommandSent       TimelineEventType = "COMMAND_SENT"
	EventTypeCommandAck        TimelineEventType = "COMMAND_ACK"
	EventTypeCommandFailed     TimelineEventType = "COMMAND_FAILED"
	EventTypeConnectionOpen    TimelineEventType = "CONNECTION_OPEN"
	EventTypeConnectionLost    TimelineEventType = "CONNECTION_LOST"
	EventTypeFCMFallback      TimelineEventType = "FCM_FALLBACK"
	EventTypeReconnected      TimelineEventType = "RECONNECTED"
	EventTypeThresholdBreach   TimelineEventType = "THRESHOLD_BREACH"
	EventTypeRegistered        TimelineEventType = "REGISTERED"
	EventTypeDeregistered     TimelineEventType = "DEREGISTERED"
	EventTypeError             TimelineEventType = "ERROR"
)

// EventCategory maps event types to frontend categories.
var EventCategory = map[TimelineEventType]string{
	EventTypeTelemetry:       "telemetry",
	EventTypeCommandSent:     "command",
	EventTypeCommandAck:      "command",
	EventTypeCommandFailed:   "command",
	EventTypeConnectionOpen:  "connection",
	EventTypeConnectionLost:  "connection",
	EventTypeFCMFallback:    "connection",
	EventTypeReconnected:    "connection",
	EventTypeThresholdBreach: "error",
	EventTypeRegistered:     "connection",
	EventTypeDeregistered:   "connection",
	EventTypeError:           "error",
}

// TimelineEvent represents a single event in the device timeline.
type TimelineEvent struct {
	ID        string            `json:"id"`
	Type      TimelineEventType `json:"type"`
	Timestamp time.Time         `json:"timestamp"`
	Data      map[string]any    `json:"data,omitempty"`
}

// DeviceInspection represents the full device inspection data.
type DeviceInspection struct {
	Identity     IdentityInfo     `json:"identity"`
	Software     SoftwareInfo     `json:"software"`
	Registration RegistrationInfo `json:"registration"`
	Connection   ConnectionInfo   `json:"connection"`
	Telemetry    TelemetryInfo    `json:"telemetry"`
}

// IdentityInfo contains device identity information.
type IdentityInfo struct {
	IMEI         string `json:"imei"`
	DeviceName   string `json:"deviceName,omitempty"`
	Model        string `json:"model,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// SoftwareInfo contains device software information.
type SoftwareInfo struct {
	OSVersion     string `json:"osVersion"`
	AppVersion    string `json:"appVersion"`
	SecurityPatch string `json:"securityPatch,omitempty"`
	BuildID       string `json:"buildId,omitempty"`
}

// RegistrationInfo contains device registration information.
type RegistrationInfo struct {
	Status              string     `json:"status"`
	RegisteredAt        *time.Time `json:"registeredAt,omitempty"`
	FCMTokenValid       bool      `json:"fcmTokenValid"`
	FCMTokenRefreshedAt *time.Time `json:"fcmTokenRefreshedAt,omitempty"`
	CommandSecretSet    bool       `json:"commandSecretSet"`
}

// ConnectionInfo contains device connection information.
type ConnectionInfo struct {
	WebSocketStatus string     `json:"webSocketStatus"`
	ConnectedAt     *time.Time `json:"connectedAt,omitempty"`
	FCMStatus       string     `json:"fcmStatus"`
	LastSeen        *time.Time `json:"lastSeen,omitempty"`
	ClientIP        string     `json:"clientIp,omitempty"`
	Protocol        string     `json:"protocol"`
}

// TelemetryInfo contains device telemetry statistics.
type TelemetryInfo struct {
	LastTimestamp time.Time `json:"lastTimestamp"`
	FramesToday  int       `json:"framesToday"`
	AvgLatencyMs int       `json:"avgLatencyMs,omitempty"`
	TotalBytesToday int64  `json:"totalBytesToday"`
	SessionsToday int      `json:"sessionsToday"`
}

// TimelineResult contains the paginated timeline result.
type TimelineResult struct {
	Events     []TimelineEvent `json:"events"`
	HasMore    bool            `json:"hasMore"`
	NextCursor string          `json:"nextCursor,omitempty"`
}

// TimelineFilter contains filter parameters for timeline queries.
type TimelineFilter struct {
	EventType string
	StartTime time.Time
	EndTime   time.Time
	Limit     int
	Cursor    string
}

// WebSocketConnectionInfo holds WebSocket connection state.
type WebSocketConnectionInfo struct {
	Connected   bool
	ConnectedAt time.Time
	ClientIP    string
}
