package models

import "time"

// Device represents a registered device in the system.
type Device struct {
	RegisteredAt      time.Time `json:"registeredAt" db:"registered_at"`
	LastSeen          time.Time `json:"lastSeen" db:"last_seen"`
	ID                string    `json:"deviceId" db:"id"`
	FirebaseInstallID string    `json:"firebaseInstallId" db:"firebase_install_id"`
	FCMToken          string    `json:"fcmToken" db:"fcm_token"`
	AppVersion        string    `json:"appVersion" db:"app_version"`
	DeviceClass       string    `json:"deviceClass" db:"device_class"`
	CommandSecret     string    `json:"commandSecret,omitempty" db:"command_secret"`
	Online            bool      `json:"online" db:"online"`
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

// DeviceStatus represents the current status of a device.
type DeviceStatus struct {
	DeviceID          string `json:"deviceId"`
	AppVersion        string `json:"appVersion"`
	DeviceClass       string `json:"deviceClass"`
	FirebaseInstallID string `json:"firebaseInstallId,omitempty"`
	FCMToken          string `json:"fcmToken,omitempty"`
	LastSeen          int64  `json:"lastSeen"`
	Online            bool   `json:"online"`
}

// UpdateDeviceRequest is the payload for updating device fields.
type UpdateDeviceRequest struct {
	FCMToken   *string `json:"fcmToken,omitempty"`
	AppVersion *string `json:"appVersion,omitempty"`
}
