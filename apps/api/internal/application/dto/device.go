package dto

import "time"

// Type aliases for backward compatibility with tests.
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
	Online      bool   `json:"online"`
	LastSeen    int64  `json:"lastSeen"`
	AppVersion  string `json:"appVersion"`
	DeviceClass string `json:"deviceClass"`
}

// UpdateFCMTokenRequest represents FCM token update request.
type UpdateFCMTokenRequest struct {
	FCMToken string `json:"fcm_token"`
}

// DeviceResponse represents a device in responses.
type DeviceResponse struct {
	ID                 string `json:"id"`
	FirebaseInstallID  string `json:"firebase_install_id"`
	AppVersion         string `json:"app_version"`
	DeviceClass        string `json:"device_class"`
	Online             bool   `json:"online"`
	RegisteredAt        int64  `json:"registered_at"`
	LastSeen           int64  `json:"last_seen"`
}

// DeviceListResponse represents a list of devices.
type DeviceListResponse struct {
	Devices   []DeviceResponse `json:"devices"`
	Total     int              `json:"total"`
	Limit     int              `json:"limit"`
	Offset    int              `json:"offset"`
}

// SendCommandRequest represents a command request.
type SendCommandRequest struct {
	DeviceID   string                 `json:"device_id"`
	Command    string                 `json:"command"`
	Args       interface{}            `json:"args,omitempty"`
	DispatchID string                `json:"dispatch_id,omitempty"`
}

// SendCommandResponse represents a command response.
type SendCommandResponse struct {
	CommandID  string    `json:"command_id"`
	DeviceID   string    `json:"device_id"`
	DispatchID string    `json:"dispatch_id"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
}

// CommandResponse represents a command in responses.
type CommandResponse struct {
	ID         string    `json:"id,omitempty"`
	DeviceID   string    `json:"deviceId,omitempty"`
	Command    string    `json:"command,omitempty"`
	Args       []byte    `json:"args,omitempty"`
	DispatchID string    `json:"dispatchId,omitempty"`
	Status     string    `json:"status,omitempty"`
	Delivery   string    `json:"delivery,omitempty"`
	ServerTime int64     `json:"serverTime,omitempty"`
	CreatedAt  time.Time `json:"createdAt,omitempty"`
	UpdatedAt  time.Time `json:"updatedAt,omitempty"`
}

// CommandRequest represents an incoming command request.
type CommandRequest struct {
	Command   string                 `json:"command"`
	Args      interface{}            `json:"args,omitempty"`
	Nonce     string                 `json:"nonce"`
	Timestamp int64                  `json:"timestamp"`
	Signature string                 `json:"signature"`
}

// CommandFrame represents a command frame for device communication.
type CommandFrame struct {
	Type       string                 `json:"type"`
	DispatchID string                 `json:"dispatchId,omitempty"`
	Command    string                 `json:"command,omitempty"`
	Args       interface{}            `json:"args,omitempty"`
	Nonce      string                 `json:"nonce,omitempty"`
	Timestamp  int64                  `json:"timestamp,omitempty"`
}

// CommandStatusResponse represents command status.
type CommandStatusResponse struct {
	CommandID    string     `json:"command_id"`
	DeviceID    string     `json:"device_id"`
	Command     string     `json:"command"`
	Status      string     `json:"status"`
	DeliveredAt *time.Time `json:"delivered_at,omitempty"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

// ClientResponse represents an API client in responses.
type ClientResponse struct {
	ID             string   `json:"id"`
	OperatorID     string   `json:"operator_id"`
	Name           string   `json:"name"`
	Platform       string   `json:"platform"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedPaths   []string `json:"allowed_paths"`
	RateLimit      int      `json:"rate_limit"`
	IsActive       bool     `json:"is_active"`
	RequestCount   int64    `json:"request_count"`
	LastRequestAt  *int64   `json:"last_request_at,omitempty"`
	CreatedAt      int64    `json:"created_at"`
	UpdatedAt      int64    `json:"updated_at"`
}

// UpdateClientRequest represents a request to update a client.
type UpdateClientRequest struct {
	Name           *string  `json:"name"`
	AllowedOrigins []string `json:"allowed_origins"`
	AllowedPaths   []string `json:"allowed_paths"`
	RateLimit     *int     `json:"rate_limit"`
	IsActive      *bool    `json:"is_active"`
}
