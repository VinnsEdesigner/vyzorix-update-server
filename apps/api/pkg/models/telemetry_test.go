package models

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestTelemetryFrame_JSON(t *testing.T) {
	frame := TelemetryFrame{
		Timestamp:    1700000000000,
		Type:         "telemetry",
		DeviceID:     "device_123",
		ActiveDevice: "speaker",
		Raw:          json.RawMessage(`{}`),
		Uptime:       3600,
		RiskScore:    25,
		AudioMode:    1,
		BufferLevel:  75,
		ThermalTemp:  45.5,
		SpeakerOn:    true,
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled TelemetryFrame
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Type != frame.Type {
		t.Errorf("Type = %q, want %q", unmarshaled.Type, frame.Type)
	}
	if unmarshaled.DeviceID != frame.DeviceID {
		t.Errorf("DeviceID = %q, want %q", unmarshaled.DeviceID, frame.DeviceID)
	}
	if unmarshaled.RiskScore != frame.RiskScore {
		t.Errorf("RiskScore = %d, want %d", unmarshaled.RiskScore, frame.RiskScore)
	}
	if unmarshaled.SpeakerOn != frame.SpeakerOn {
		t.Errorf("SpeakerOn = %v, want %v", unmarshaled.SpeakerOn, frame.SpeakerOn)
	}
}

func TestTelemetryFrame_JSONTags(t *testing.T) {
	data := []byte(`{
		"timestamp": 1700000000000,
		"type": "telemetry",
		"deviceId": "dev_abc",
		"activeDevice": "bluetooth",
		"uptime": 7200,
		"riskScore": 10,
		"audioMode": 2,
		"bufferLevel": 50,
		"thermalTemp": 38.0,
		"speakerOn": false
	}`)

	var frame TelemetryFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if frame.DeviceID != "dev_abc" {
		t.Errorf("DeviceID = %q, want \"dev_abc\"", frame.DeviceID)
	}
	if frame.RiskScore != 10 {
		t.Errorf("RiskScore = %d, want 10", frame.RiskScore)
	}
	if frame.SpeakerOn != false {
		t.Error("SpeakerOn should be false")
	}
}

func TestTelemetryFrame_TimestampTypes(t *testing.T) {
	tests := []struct {
		name      string
		timestamp any
		wantValid bool
	}{
		{"integer timestamp", int64(1700000000000), true},
		{"float timestamp", float64(1700000000000), true},
		{"string timestamp", "2024-01-01T00:00:00Z", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			frame := TelemetryFrame{
				Timestamp: tt.timestamp,
				Type:      "telemetry",
			}

			data, err := json.Marshal(frame)
			if err != nil {
				t.Fatalf("json.Marshal() error = %v", err)
			}

			var unmarshaled TelemetryFrame
			if err := json.Unmarshal(data, &unmarshaled); err != nil {
				t.Fatalf("json.Unmarshal() error = %v", err)
			}

			if unmarshaled.Timestamp == nil && tt.wantValid {
				t.Error("Timestamp should not be nil")
			}
		})
	}
}

func TestTelemetryFrame_OptionalFields(t *testing.T) {
	frame := TelemetryFrame{
		Type: "telemetry",
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled TelemetryFrame
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Type != "telemetry" {
		t.Errorf("Type = %q, want \"telemetry\"", unmarshaled.Type)
	}
	if unmarshaled.DeviceID != "" {
		t.Errorf("DeviceID = %q, want \"\"", unmarshaled.DeviceID)
	}
}

func TestTelemetryFrame_RawNotSerialized(t *testing.T) {
	frame := TelemetryFrame{
		Type: "telemetry",
		Raw:  json.RawMessage(`{"extra":"data"}`),
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// Raw field should not appear in JSON output (has json:"-" tag)
	if strings.Contains(string(data), "extra") {
		t.Error("JSON should not contain Raw field")
	}
}

func TestTelemetryFrame_MaxBoundaryValues(t *testing.T) {
	frame := TelemetryFrame{
		Type:        "telemetry",
		RiskScore:   100,
		ThermalTemp: 120.0,
		BufferLevel: 100,
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled TelemetryFrame
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.RiskScore != 100 {
		t.Errorf("RiskScore = %d, want 100", unmarshaled.RiskScore)
	}
	if unmarshaled.BufferLevel != 100 {
		t.Errorf("BufferLevel = %d, want 100", unmarshaled.BufferLevel)
	}
}

func TestTelemetryFrame_MinBoundaryValues(t *testing.T) {
	frame := TelemetryFrame{
		Type:        "telemetry",
		RiskScore:   0,
		ThermalTemp: -40.0,
		BufferLevel: 0,
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled TelemetryFrame
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.RiskScore != 0 {
		t.Errorf("RiskScore = %d, want 0", unmarshaled.RiskScore)
	}
	if unmarshaled.BufferLevel != 0 {
		t.Errorf("BufferLevel = %d, want 0", unmarshaled.BufferLevel)
	}
}
