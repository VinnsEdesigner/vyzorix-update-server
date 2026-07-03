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
