package models

import "encoding/json"

type TelemetryFrame struct {
	Timestamp    any             `json:"timestamp,omitempty"`
	Type         string          `json:"type"`
	DeviceID     string          `json:"deviceId,omitempty"`
	ActiveDevice string          `json:"activeDevice,omitempty"`
	Raw          json.RawMessage `json:"-"`
	Uptime       int64           `json:"uptime,omitempty"`
	RiskScore    int             `json:"riskScore,omitempty"`
	AudioMode    int             `json:"audioMode,omitempty"`
	BufferLevel  int             `json:"bufferLevel,omitempty"`
	ThermalTemp  float64         `json:"thermalTemp,omitempty"`
	SpeakerOn    bool            `json:"speakerOn,omitempty"`
}
