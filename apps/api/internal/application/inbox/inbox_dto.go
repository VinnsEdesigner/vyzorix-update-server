package inbox

import "time"

// InboxRequest represents an incoming device registration request.
type InboxRequest struct {
	IMEI              string `json:"imei"`
	DeviceName        string `json:"deviceName"`
	DeviceClass       string `json:"deviceClass"`
	Model             string `json:"model"`
	Manufacturer      string `json:"manufacturer"`
	OSVersion         string `json:"osVersion"`
	AppVersion        string `json:"appVersion"`
	FCMToken          string `json:"fcmToken"`
	FirebaseInstallID string `json:"firebaseInstallId"`
	// IdempotencyKey is an optional client-provided key to prevent duplicate registrations.
	// If provided, duplicate requests with the same key will return the existing entry.
	IdempotencyKey string `json:"idempotencyKey,omitempty"`
}

// InboxListResponse represents the response for GET /v1/device/inbox.
type InboxListResponse struct {
	Requests   []InboxEntryResponse `json:"requests"`
	Pagination PaginationResponse   `json:"pagination"`
}

// InboxEntryResponse represents an inbox entry in API responses.
type InboxEntryResponse struct {
	AcknowledgedAt    *int64 `json:"acknowledgedAt,omitempty"`
	ApprovingAt       *int64 `json:"approvingAt,omitempty"`
	ApprovedAt        *int64 `json:"approvedAt,omitempty"`
	RejectedAt        *int64 `json:"rejectedAt,omitempty"`
	Model             string `json:"model"`
	FirebaseInstallID string `json:"firebaseInstallId"`
	ID                string `json:"id"`
	Manufacturer      string `json:"manufacturer"`
	OSVersion         string `json:"osVersion"`
	AppVersion        string `json:"appVersion"`
	FCMToken          string `json:"fcmToken"`
	DeviceClass       string `json:"deviceClass,omitempty"`
	Status            string `json:"status"`
	OperatorID        string `json:"operatorId,omitempty"`
	DeviceName        string `json:"deviceName,omitempty"`
	IMEI              string `json:"imei"`
	Notes             string `json:"notes,omitempty"`
	CreatedAt         int64  `json:"createdAt"`
}

// AckRequest represents the request for POST /v1/device/inbox/:imei/ack.
type AckRequest struct {
	Action string `json:"action"`
	Notes  string `json:"notes,omitempty"`
}

// AckResponse represents the response for POST /v1/device/inbox/:imei/ack.
type AckResponse struct {
	AcknowledgedAt *int64 `json:"acknowledgedAt,omitempty"`
	ApprovedAt     *int64 `json:"approvedAt,omitempty"`
	RejectedAt     *int64 `json:"rejectedAt,omitempty"`
	ID             string `json:"id"`
	IMEI           string `json:"imei"`
	Status         string `json:"status"`
	CommandSecret  string `json:"commandSecret,omitempty"`
	Notes          string `json:"notes,omitempty"`
	FCMPushSent    bool   `json:"fcmPushSent"`
}

// PaginationResponse represents pagination info in responses.
type PaginationResponse struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"totalPages"`
}

// DeviceListResponse represents the response for GET /v1/devices.
type DeviceListResponse struct {
	Devices    []DeviceResponse   `json:"devices"`
	Pagination PaginationResponse `json:"pagination"`
}

// DeviceResponse represents a device in API responses.
type DeviceResponse struct {
	RegisteredAt *int64 `json:"registeredAt,omitempty"`
	ID           string `json:"id"`
	IMEI         string `json:"imei"`
	DeviceName   string `json:"deviceName,omitempty"`
	Model        string `json:"model,omitempty"`
	Manufacturer string `json:"manufacturer,omitempty"`
	OSVersion    string `json:"osVersion,omitempty"`
	AppVersion   string `json:"appVersion,omitempty"`
	Status       string `json:"status"`
	LastSeen     int64  `json:"lastSeen,omitempty"`
	Online       bool   `json:"online"`
}

// DeregisterResponse represents the response for DELETE /v1/device/:imei.
type DeregisterResponse struct {
	IMEI           string `json:"imei"`
	Status         string `json:"status"`
	DeregisteredAt int64  `json:"deregisteredAt"`
	RetentionUntil int64  `json:"retentionUntil"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Details interface{} `json:"details,omitempty"`
	Code    string      `json:"error"`
	Message string      `json:"message"`
}

// NewTimestamp creates a new timestamp in milliseconds.
func NewTimestamp(t time.Time) int64 {
	return t.UnixMilli()
}

// PtrToInt64 returns a pointer to an int64.
func PtrToInt64(v int64) *int64 {
	return &v
}
