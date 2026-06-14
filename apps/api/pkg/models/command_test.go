package models

import (
	"encoding/json"
	"testing"
)

func TestCommandRequest_JSON(t *testing.T) {
	req := CommandRequest{
		Command:   "FORCE_SPEAKER",
		Nonce:     "nonce123abc",
		Signature: "sig_abc123",
		Args:      json.RawMessage(`{"volume":50}`),
		Timestamp: 1700000000000,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled CommandRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Command != req.Command {
		t.Errorf("Command = %q, want %q", unmarshaled.Command, req.Command)
	}
	if unmarshaled.Nonce != req.Nonce {
		t.Errorf("Nonce = %q, want %q", unmarshaled.Nonce, req.Nonce)
	}
	if unmarshaled.Signature != req.Signature {
		t.Errorf("Signature = %q, want %q", unmarshaled.Signature, req.Signature)
	}
	if unmarshaled.Timestamp != req.Timestamp {
		t.Errorf("Timestamp = %d, want %d", unmarshaled.Timestamp, req.Timestamp)
	}
}

func TestCommandRequest_JSONTags(t *testing.T) {
	data := []byte(`{
		"command": "REBOOT",
		"nonce": "xyz789",
		"signature": "sig_xyz",
		"args": {"force": true},
		"timestamp": 1700000000001
	}`)

	var req CommandRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if req.Command != "REBOOT" {
		t.Errorf("Command = %q, want \"REBOOT\"", req.Command)
	}
}

func TestCommandRequest_OptionalSignature(t *testing.T) {
	// Signature is optional
	req := CommandRequest{
		Command:   "PING",
		Nonce:     "nonce456",
		Timestamp: 1700000000000,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled CommandRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Signature != "" {
		t.Errorf("Signature = %q, want \"\"", unmarshaled.Signature)
	}
}

func TestCommandRequest_EmptyArgs(t *testing.T) {
	req := CommandRequest{
		Command:   "STATUS",
		Nonce:     "nonce789",
		Timestamp: 1700000000000,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled CommandRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Args should be nil or empty
	if len(unmarshaled.Args) != 0 {
		t.Errorf("Args length = %d, want 0", len(unmarshaled.Args))
	}
}

func TestCommandFrame_JSON(t *testing.T) {
	frame := CommandFrame{
		Type:       "command",
		DispatchID: "dispatch_abc",
		Command:    "FORCE_SPEAKER",
		Nonce:      "frame_nonce",
		Signature:  "frame_sig",
		Args:       json.RawMessage(`{"level":75}`),
		Timestamp:  1700000000000,
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled CommandFrame
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Type != frame.Type {
		t.Errorf("Type = %q, want %q", unmarshaled.Type, frame.Type)
	}
	if unmarshaled.DispatchID != frame.DispatchID {
		t.Errorf("DispatchID = %q, want %q", unmarshaled.DispatchID, frame.DispatchID)
	}
	if unmarshaled.Command != frame.Command {
		t.Errorf("Command = %q, want %q", unmarshaled.Command, frame.Command)
	}
}

func TestCommandFrame_TypeCommand(t *testing.T) {
	frame := CommandFrame{
		Type:       "command",
		DispatchID: "dispatch_1",
		Command:    "REINIT_PROJECTION",
		Nonce:      "n1",
		Timestamp:  1700000000000,
	}

	if frame.Type != "command" {
		t.Errorf("Type = %q, want \"command\"", frame.Type)
	}
}

func TestCommandResponse_JSON(t *testing.T) {
	resp := CommandResponse{
		DispatchID: "dispatch_xyz",
		Delivery:   "sent",
		ServerTime: 1700000000000,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled CommandResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.DispatchID != resp.DispatchID {
		t.Errorf("DispatchID = %q, want %q", unmarshaled.DispatchID, resp.DispatchID)
	}
	if unmarshaled.Delivery != resp.Delivery {
		t.Errorf("Delivery = %q, want %q", unmarshaled.Delivery, resp.Delivery)
	}
}

func TestCommandResponse_DeliveryValues(t *testing.T) {
	testCases := []struct {
		delivery string
		valid    bool
	}{
		{"sent", true},
		{"queued", true},
		{"unknown", false},
		{"", false},
	}

	for _, tc := range testCases {
		resp := CommandResponse{
			DispatchID: "test",
			Delivery:   tc.delivery,
			ServerTime: 1700000000000,
		}

		data, err := json.Marshal(resp)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		var unmarshaled CommandResponse
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if tc.valid && unmarshaled.Delivery != tc.delivery {
			t.Errorf("Delivery = %q, want %q", unmarshaled.Delivery, tc.delivery)
		}
	}
}

func TestCommandFrame_ArgsContent(t *testing.T) {
	complexArgs := `{"volume":50,"mute":false,"mode":"normal","devices":["speaker","headphones"]}`
	frame := CommandFrame{
		Type:       "command",
		DispatchID: "dispatch_complex",
		Command:    "SET_AUDIO",
		Nonce:      "nonce_complex",
		Args:       json.RawMessage(complexArgs),
		Timestamp:  1700000000000,
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled CommandFrame
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	// Verify Args content is preserved
	var args map[string]any
	if err := json.Unmarshal(unmarshaled.Args, &args); err != nil {
		t.Fatalf("Failed to parse Args: %v", err)
	}

	if volume, ok := args["volume"].(float64); !ok || int(volume) != 50 {
		t.Errorf("Args.volume = %v, want 50", args["volume"])
	}
}

func TestCommandFrame_DispatchIDFormat(t *testing.T) {
	testIDs := []string{
		"dispatch_abc123",
		"tx-123",
		"unique-id-456",
		"ABCDEF123456",
	}

	for _, id := range testIDs {
		frame := CommandFrame{
			Type:       "command",
			DispatchID: id,
			Command:    "PING",
			Nonce:      "n",
			Timestamp:  1700000000000,
		}

		data, err := json.Marshal(frame)
		if err != nil {
			t.Fatalf("json.Marshal() error = %v", err)
		}

		var unmarshaled CommandFrame
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Fatalf("json.Unmarshal() error = %v", err)
		}

		if unmarshaled.DispatchID != id {
			t.Errorf("DispatchID round-trip failed: got %q, want %q", unmarshaled.DispatchID, id)
		}
	}
}
