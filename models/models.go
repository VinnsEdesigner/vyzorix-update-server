package models

import (
	"encoding/json"
	"time"
)

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
	DeviceID    string `json:"deviceId"`
	Online      bool   `json:"online"`
	LastSeen    int64  `json:"lastSeen"`
	AppVersion  string `json:"appVersion"`
	DeviceClass string `json:"deviceClass"`
}

type TelemetryFrame struct {
	Type         string          `json:"type"`
	DeviceID     string          `json:"deviceId,omitempty"`
	Uptime       int64           `json:"uptime,omitempty"`
	RiskScore    int             `json:"riskScore,omitempty"`
	AudioMode    int             `json:"audioMode,omitempty"`
	SpeakerOn    bool            `json:"speakerOn,omitempty"`
	ActiveDevice string          `json:"activeDevice,omitempty"`
	BufferLevel  int             `json:"bufferLevel,omitempty"`
	ThermalTemp  float64         `json:"thermalTemp,omitempty"`
	Timestamp    any             `json:"timestamp,omitempty"`
	Raw          json.RawMessage `json:"-"`
}

type CommandRequest struct {
	Command   string          `json:"command"`
	Args      json.RawMessage `json:"args,omitempty"`
	Nonce     string          `json:"nonce"`
	Timestamp int64           `json:"timestamp"`
	Signature string          `json:"signature,omitempty"`
}
type CommandFrame struct {
	Type       string          `json:"type"`
	DispatchID string          `json:"dispatchId"`
	Command    string          `json:"command"`
	Args       json.RawMessage `json:"args,omitempty"`
	Nonce      string          `json:"nonce"`
	Timestamp  int64           `json:"timestamp"`
	Signature  string          `json:"signature,omitempty"`
}
type CommandResponse struct {
	DispatchID string `json:"dispatchId"`
	Delivery   string `json:"delivery"`
	ServerTime int64  `json:"serverTime"`
}
