package models

import "encoding/json"

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
