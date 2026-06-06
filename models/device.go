package models

import "time"

type Device struct {
	ID                string    `json:"deviceId" db:"id"`
	FirebaseInstallID string    `json:"firebaseInstallId" db:"firebase_install_id"`
	FCMToken          string    `json:"fcmToken" db:"fcm_token"`
	AppVersion        string    `json:"appVersion" db:"app_version"`
	DeviceClass       string    `json:"deviceClass" db:"device_class"`
	CommandSecret     string    `json:"commandSecret,omitempty" db:"command_secret"`
	Online            bool      `json:"online" db:"online"`
	RegisteredAt      time.Time `json:"registeredAt" db:"registered_at"`
	LastSeen          time.Time `json:"lastSeen" db:"last_seen"`
}

type RegisterRequest struct {
	DeviceID          string `json:"deviceId"`
	FirebaseInstallID string `json:"firebaseInstallId"`
	FCMToken          string `json:"fcmToken"`
	AppVersion        string `json:"appVersion"`
	DeviceClass       string `json:"deviceClass"`
}

type RegisterResponse struct {
	DeviceID      string `json:"deviceId"`
	CommandSecret string `json:"commandSecret"`
	RegisteredAt  int64  `json:"registeredAt"`
	ServerTime    int64  `json:"serverTime"`
}

type DeviceStatus struct {
	DeviceID          string `json:"deviceId"`
	Online            bool   `json:"online"`
	LastSeen          int64  `json:"lastSeen"`
	AppVersion        string `json:"appVersion"`
	DeviceClass       string `json:"deviceClass"`
	FirebaseInstallID string `json:"firebaseInstallId,omitempty"`
	FCMToken          string `json:"fcmToken,omitempty"`
}

// UpdateDeviceRequest is the payload for updating device fields.
type UpdateDeviceRequest struct {
	FCMToken   *string `json:"fcmToken,omitempty"`
	AppVersion *string `json:"appVersion,omitempty"`
}
