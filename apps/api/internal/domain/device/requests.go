package device

import "encoding/json"

// UpdateDeviceRequest is the payload for updating device settings.
type UpdateDeviceRequest struct {
	FirmwareVersion *string         `json:"firmwareVersion,omitempty"`
	Config          json.RawMessage `json:"config,omitempty"`
	Tags            []string        `json:"tags,omitempty"`
	Name            *string         `json:"name,omitempty"`
}

// RegisterRequest is the payload for device registration.
type RegisterRequest struct {
    DeviceID          string `json:"deviceId"`
    FirebaseInstallID string `json:"firebaseInstallId"`
    FCMToken          string `json:"fcmToken"`
    AppVersion        string `json:"appVersion"`
    DeviceClass       string `json:"deviceClass"`
}

// RegisterResponse is returned after successful device registration.
type RegisterResponse struct {
    DeviceID      string `json:"deviceId"`
    CommandSecret string `json:"commandSecret"`
    RegisteredAt  int64  `json:"registeredAt"`
    ServerTime    int64  `json:"serverTime"`
}

