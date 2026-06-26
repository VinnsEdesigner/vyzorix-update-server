package logs

// ListLogsRequest represents the request for GET /v1/device/:imei/logs.
type ListLogsRequest struct {
	DeviceID  string `param:"imei" validate:"required"`
	EventType string `query:"type"`
	Cursor    string `query:"cursor"`
	StartTime int64  `query:"startTime"`
	EndTime   int64  `query:"endTime"`
	Limit     int    `query:"limit"`
}

// ListLogsResponse represents the response for GET /v1/device/:imei/logs.
type ListLogsResponse struct {
	Events     []LogEvent       `json:"events"`
	Pagination CursorPagination `json:"pagination"`
}

// LogEvent represents a single log event.
type LogEvent struct {
	Data      map[string]interface{} `json:"data,omitempty"`
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Timestamp int64                  `json:"timestamp"`
}

// CursorPagination represents cursor-based pagination.
type CursorPagination struct {
	NextCursor string `json:"nextCursor,omitempty"`
	Limit      int    `json:"limit"`
	HasMore    bool   `json:"hasMore"`
}
