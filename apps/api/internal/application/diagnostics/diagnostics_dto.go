// Package diagnostics provides DTOs for the diagnostics API.
package diagnostics

// InspectRequest represents the request for device inspection.
type InspectRequest struct {
	IMEI string `json:"imei"`
}

// TimelineRequest represents the request for device timeline.
type TimelineRequest struct {
	IMEI     string `json:"imei" form:"imei"`
	EventType string `json:"eventType" form:"eventType"`
	StartTime int64  `json:"startTime" form:"startTime"`
	EndTime   int64  `json:"endTime" form:"endTime"`
	Limit     int    `json:"limit" form:"limit"`
	Cursor    string `json:"cursor" form:"cursor"`
}

// TimelineResponse represents the response for device timeline.
type TimelineResponse struct {
	Events     []EventDTO      `json:"events"`
	Pagination PaginationDTO   `json:"pagination"`
}

// EventDTO represents a timeline event in the response.
type EventDTO struct {
	ID        string         `json:"id"`
	Type      string         `json:"type"`
	Timestamp int64          `json:"timestamp"`
	Data      map[string]any `json:"data,omitempty"`
}

// PaginationDTO represents pagination info in the response.
type PaginationDTO struct {
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
	NextCursor string `json:"nextCursor,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
	Code    string `json:"code,omitempty"`
}

// HTTP-specific response types with int64 timestamps per spec

// HTTPInspectionResponse is the HTTP API response for device inspection.
type HTTPInspectionResponse struct {
	Identity     HTTPIdentityInfo     `json:"identity"`
	Software     HTTPSoftwareInfo     `json:"software"`
	Registration HTTPRegistrationInfo `json:"registration"`
	Connection   HTTPConnectionInfo   `json:"connection"`
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
	FCMTokenValid       bool   `json:"fcmTokenValid"`
	FCMTokenRefreshedAt int64  `json:"fcmTokenRefreshedAt,omitempty"`
	CommandSecretSet    bool   `json:"commandSecretSet"`
}

// HTTPConnectionInfo contains connection info for HTTP API with int64 timestamps.
type HTTPConnectionInfo struct {
	WebSocketStatus string `json:"webSocketStatus"`
	ConnectedAt     int64  `json:"connectedAt,omitempty"`
	FCMStatus       string `json:"fcmStatus"`
	LastSeen        int64  `json:"lastSeen,omitempty"`
	ClientIP        string `json:"clientIp,omitempty"`
	Protocol        string `json:"protocol"`
}

// HTTPTelemetryInfo contains telemetry stats for HTTP API with int64 timestamps.
type HTTPTelemetryInfo struct {
	LastTimestamp   int64 `json:"lastTimestamp"`
	FramesToday     int   `json:"framesToday"`
	AvgLatencyMs    int   `json:"avgLatencyMs,omitempty"`
	TotalBytesToday int64 `json:"totalBytesToday"`
	SessionsToday   int   `json:"sessionsToday"`
}
