package dto

import "time"

// RegisterDeviceRequest represents a device registration request.
type RegisterDeviceRequest struct {
	DeviceID          string `json:"device_id"`
	FirebaseInstallID string `json:"firebase_install_id"`
	FCMToken         string `json:"fcm_token"`
	AppVersion        string `json:"app_version"`
	DeviceClass       string `json:"device_class"`
}

// RegisterDeviceResponse represents a device registration response.
type RegisterDeviceResponse struct {
	DeviceID      string `json:"device_id"`
	CommandSecret string `json:"command_secret"`
	RegisteredAt  int64  `json:"registered_at"`
}

// DeviceStatusResponse represents device status.
type DeviceStatusResponse struct {
	DeviceID    string `json:"device_id"`
	Online      bool   `json:"online"`
	LastSeen    int64  `json:"last_seen"`
	AppVersion  string `json:"app_version"`
	DeviceClass string `json:"device_class"`
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
	ID         string                 `json:"id"`
	DeviceID   string                 `json:"device_id"`
	Command    string                 `json:"command"`
	Args       []byte                 `json:"args,omitempty"`
	DispatchID string                `json:"dispatch_id"`
	Status     string                 `json:"status"`
	CreatedAt  time.Time              `json:"created_at"`
	UpdatedAt  time.Time              `json:"updated_at"`
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
