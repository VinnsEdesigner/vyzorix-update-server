// Package diagnostics provides DTOs for the diagnostics API.
package diagnostics

// InspectRequest represents the request for device inspection.
type InspectRequest struct {
	IMEI string `json:"imei"`
}

// TimelineRequest represents the request for device timeline.
type TimelineRequest struct {
	IMEI      string `json:"imei" form:"imei"`
	EventType string `json:"eventType" form:"eventType"`
	Cursor    string `json:"cursor" form:"cursor"`
	StartTime int64  `json:"startTime" form:"startTime"`
	EndTime   int64  `json:"endTime" form:"endTime"`
	Limit     int    `json:"limit" form:"limit"`
}

// TimelineResponse represents the response for device timeline.
type TimelineResponse struct {
	Events     []EventDTO    `json:"events"`
	Pagination PaginationDTO `json:"pagination"`
}

// EventDTO represents a timeline event in the response.
type EventDTO struct {
	Data      map[string]any `json:"data,omitempty"`
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Timestamp int64          `json:"timestamp"`
}

// PaginationDTO represents pagination info in the response.
type PaginationDTO struct {
	NextCursor string `json:"nextCursor,omitempty"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// HTTP-specific response types with int64 timestamps per spec.

// HTTPInspectionResponse is the HTTP API response for device inspection.
type HTTPInspectionResponse struct {
	Connection   HTTPConnectionInfo   `json:"connection"`
	Identity     HTTPIdentityInfo     `json:"identity"`
	Software     HTTPSoftwareInfo     `json:"software"`
	Registration HTTPRegistrationInfo `json:"registration"`
	Telemetry    HTTPTelemetryInfo    `json:"telemetry"`
}

// HTTPIdentityInfo contains device identity for HTTP API.
type HTTPIdentityInfo struct {
	IMEI         string `json:"imei"`
	DeviceName   string `json:"deviceName,omitempty"`
	Model        string `json:"model,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
}

// HTTPSoftwareInfo contains device software info for HTTP API.
type HTTPSoftwareInfo struct {
	OSVersion     string `json:"osVersion"`
	AppVersion    string `json:"appVersion"`
	SecurityPatch string `json:"securityPatch,omitempty"`
	BuildID       string `json:"buildId,omitempty"`
}

// HTTPRegistrationInfo contains registration info for HTTP API with int64 timestamps.
type HTTPRegistrationInfo struct {
	Status              string `json:"status"`
	RegisteredAt        int64  `json:"registeredAt,omitempty"`
	FCMTokenRefreshedAt int64  `json:"fcmTokenRefreshedAt,omitempty"`
	FCMTokenValid       bool   `json:"fcmTokenValid"`
	CommandSecretSet    bool   `json:"commandSecretSet"`
}

// HTTPConnectionInfo contains connection info for HTTP API with int64 timestamps.
type HTTPConnectionInfo struct {
	WebSocketStatus string `json:"webSocketStatus"`
	FCMStatus       string `json:"fcmStatus"`
	ClientIP        string `json:"clientIp,omitempty"`
	Protocol        string `json:"protocol"`
	ConnectedAt     int64  `json:"connectedAt,omitempty"`
	LastSeen        int64  `json:"lastSeen,omitempty"`
}

// HTTPTelemetryInfo contains telemetry stats for HTTP API with int64 timestamps.
type HTTPTelemetryInfo struct {
	LastTimestamp   int64 `json:"lastTimestamp"`
	FramesToday     int   `json:"framesToday"`
	AvgLatencyMs    int   `json:"avgLatencyMs,omitempty"`
	TotalBytesToday int64 `json:"totalBytesToday"`
	SessionsToday   int   `json:"sessionsToday"`
}
