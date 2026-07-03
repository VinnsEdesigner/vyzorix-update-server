package dto

import "time"

// DeviceRegisterRequest is an alias for RegisterDeviceRequest.
type DeviceRegisterRequest = RegisterDeviceRequest
type DeviceRegisterResponse = RegisterDeviceResponse
type DeviceStatus = DeviceStatusResponse

// RegisterDeviceRequest represents a device registration request.
type RegisterDeviceRequest struct {
	DeviceID          string `json:"deviceId"`
	FirebaseInstallID string `json:"firebaseInstallId"`
	FCMToken          string `json:"fcmToken"`
	AppVersion        string `json:"appVersion"`
	DeviceClass       string `json:"deviceClass"`
}

// RegisterDeviceResponse represents a device registration response.
type RegisterDeviceResponse struct {
	DeviceID      string `json:"deviceId"`
	CommandSecret string `json:"commandSecret"`
	RegisteredAt  int64  `json:"registeredAt"`
}

// DeviceStatusResponse represents device status.
type DeviceStatusResponse struct {
	DeviceID    string `json:"deviceId"`
	AppVersion  string `json:"appVersion"`
	DeviceClass string `json:"deviceClass"`
	LastSeen    int64  `json:"lastSeen"`
	Online      bool   `json:"online"`
}

// UpdateFCMTokenRequest represents FCM token update request.
type UpdateFCMTokenRequest struct {
	FCMToken string `json:"fcm_token"`
}

// DeviceResponse represents a device in responses.
type DeviceResponse struct {
	ID                string `json:"id"`
	FirebaseInstallID string `json:"firebase_install_id"`
	AppVersion        string `json:"app_version"`
	DeviceClass       string `json:"device_class"`
	Online            bool   `json:"online"`
	RegisteredAt      int64  `json:"registered_at"`
	LastSeen          int64  `json:"last_seen"`
}

// DeviceListResponse represents a list of devices.
type DeviceListResponse struct {
	Devices    []DeviceResponse `json:"devices"`
	Total      int              `json:"total"`
	Limit      int              `json:"limit"`
	Offset     int              `json:"offset"`
	Page       int              `json:"page"`
	TotalPages int              `json:"totalPages"`
}

// DeviceDetailResponse represents detailed device info per spec.
type DeviceDetailResponse struct {
	ID                string `json:"id"`
	IMEI              string `json:"imei"`
	DeviceName        string `json:"deviceName"`
	Model             string `json:"model"`
	Manufacturer      string `json:"manufacturer"`
	OSVersion         string `json:"osVersion"`
	AppVersion        string `json:"appVersion"`
	SecurityPatch     string `json:"securityPatch"`
	Status            string `json:"status"`
	RegisteredAt      int64  `json:"registeredAt"`
	LastSeen          int64  `json:"lastSeen"`
	FCMTokenValid     bool   `json:"fcmTokenValid"`
	CommandSecretSet  bool   `json:"commandSecretSet"`
	Connection        *ConnectionInfo `json:"connection,omitempty"`
}

// ConnectionInfo represents WebSocket connection details.
type ConnectionInfo struct {
	WebSocketStatus string `json:"webSocketStatus"`
	ConnectedAt     int64  `json:"connectedAt"`
	Protocol        string `json:"protocol"`
	ClientIP        string `json:"clientIp"`
}

// SendCommandRequest represents a command request.
type SendCommandRequest struct {
	DeviceID   string      `json:"device_id"`
	Command    string      `json:"command"`
	Args       interface{} `json:"args,omitempty"`
	DispatchID string      `json:"dispatch_id,omitempty"`
}

// SendCommandResponse represents a command response.
type SendCommandResponse struct {
	CreatedAt  time.Time `json:"created_at"`
	CommandID  string    `json:"command_id"`
	DeviceID   string    `json:"device_id"`
	DispatchID string    `json:"dispatch_id"`
	Status     string    `json:"status"`
}

// CommandResponse represents a command in responses.
type CommandResponse struct {
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
	ID         string    `json:"id,omitempty"`
	DeviceID   string    `json:"deviceId,omitempty"`
	Command    string    `json:"command,omitempty"`
	DispatchID string    `json:"dispatchId,omitempty"`
	Status     string    `json:"status,omitempty"`
	Delivery   string    `json:"delivery,omitempty"`
	Args       []byte    `json:"args,omitempty"`
	ServerTime int64     `json:"serverTime,omitempty"`
}

// CommandRequest represents an incoming command request.
type CommandRequest struct {
	Args      interface{} `json:"args,omitempty"`
	Command   string      `json:"command"`
	Nonce     string      `json:"nonce"`
	Signature string      `json:"signature"`
	Timestamp int64       `json:"timestamp"`
}

// CommandFrame represents a command frame for device communication.
type CommandFrame struct {
	Type       string      `json:"type"`
	DispatchID string      `json:"dispatchId,omitempty"`
	Command    string      `json:"command,omitempty"`
	Args       interface{} `json:"args,omitempty"`
	Nonce      string      `json:"nonce,omitempty"`
	Timestamp  int64       `json:"timestamp,omitempty"`
}

// CommandStatusResponse represents command status.
type CommandStatusResponse struct {
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CommandID   string     `json:"command_id"`
	DispatchID  string     `json:"dispatch_id,omitempty"`
	DeviceID    string     `json:"device_id"`
	Command     string     `json:"command"`
	Status      string     `json:"status"`
}

// ClientResponse represents an API client in responses.
type ClientResponse struct {
	LastRequestAt  *int64   `json:"last_request_at,omitempty"`
	ID             string   `json:"id"`
	OperatorID     string   `json:"operator_id"`
	Name           string   `json:"name"`
	Platform       string   `json:"platform"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedPaths   []string `json:"allowed_paths"`
	RateLimit      int      `json:"rate_limit"`
	RequestCount   int64    `json:"request_count"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
	IsActive       bool     `json:"is_active"`
}

// UpdateClientRequest represents a request to update a client.
type UpdateClientRequest struct {
	Name           *string  `json:"name"`
	RateLimit      *int     `json:"rate_limit"`
	IsActive       *bool    `json:"is_active"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedPaths   []string `json:"allowed_paths"`
}
