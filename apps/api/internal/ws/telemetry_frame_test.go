package hub

import (
	"encoding/json"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/telemetry"
)

// TestTelemetryFrameClassification guards the single-decode path in ReadPump: the
// frame type must be read from the decoded struct, so a frame whose nested object
// also contains a "type" key is still classified by its top-level type.
func TestTelemetryFrameClassification(t *testing.T) {
	cases := []struct {
		name string
		raw  string
		want bool
	}{
		{"plain telemetry", `{"type":"telemetry","riskScore":42}`, true},
		{"telemetry with nested type", `{"type":"telemetry","nested":{"type":"other"},"riskScore":1}`, true},
		{"nested type before top-level", `{"nested":{"type":"telemetry"},"type":"command"}`, false},
		{"command frame", `{"type":"command","dispatchId":"d1"}`, false},
		{"empty type", `{"riskScore":1}`, false},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var frame telemetry.TelemetryFrame
			if err := json.Unmarshal([]byte(tc.raw), &frame); err != nil {
				t.Fatalf("unmarshal: %v", err)
			}
			if got := frame.Type == "telemetry"; got != tc.want {
				t.Fatalf("classification = %v, want %v (type=%q)", got, tc.want, frame.Type)
			}
		})
	}
}
