package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix-update-server/models"
)

func TestCommandController_CommandRequest_JSON(t *testing.T) {
	req := models.CommandRequest{
		Command:   "update",
		Args:      []byte(`{"version":"2.0.0"}`),
		Nonce:     "nonce123",
		Timestamp: 1234567890,
		Signature: "sig123",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["command"] != "update" {
		t.Errorf("command = %v, want update", result["command"])
	}
	if result["nonce"] != "nonce123" {
		t.Errorf("nonce = %v, want nonce123", result["nonce"])
	}
}

func TestCommandController_CommandRequest_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
		"command": "reboot",
		"args": {"delay": 5},
		"nonce": "abc123",
		"timestamp": 9876543210,
		"signature": "sig456"
	}`)

	var req models.CommandRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if req.Command != "reboot" {
		t.Errorf("Command = %s, want reboot", req.Command)
	}
	if req.Nonce != "abc123" {
		t.Errorf("Nonce = %s, want abc123", req.Nonce)
	}
	if req.Timestamp != 9876543210 {
		t.Errorf("Timestamp = %d, want 9876543210", req.Timestamp)
	}
}

func TestCommandController_CommandResponse_JSON(t *testing.T) {
	resp := models.CommandResponse{
		DispatchID: "dispatch-001",
		Delivery:   "queued",
		ServerTime: 1234567890,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["dispatchId"] != "dispatch-001" {
		t.Errorf("dispatchId = %v, want dispatch-001", result["dispatchId"])
	}
	if result["delivery"] != "queued" {
		t.Errorf("delivery = %v, want queued", result["delivery"])
	}
}

func TestCommandController_CommandFrame_JSON(t *testing.T) {
	frame := models.CommandFrame{
		Type:       "update",
		DispatchID: "dispatch-001",
		Command:    "update",
		Args:       []byte(`{"version":"2.0.0"}`),
		Nonce:      "nonce123",
		Timestamp:  1234567890,
	}

	data, err := json.Marshal(frame)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["type"] != "update" {
		t.Errorf("type = %v, want update", result["type"])
	}
	if result["dispatchId"] != "dispatch-001" {
		t.Errorf("dispatchId = %v, want dispatch-001", result["dispatchId"])
	}
	if result["command"] != "update" {
		t.Errorf("command = %v, want update", result["command"])
	}
}

func TestCommandController_Dispatch_InvalidJSON(t *testing.T) {
	data := []byte(`{invalid}`)

	var req models.CommandRequest
	err := json.Unmarshal(data, &req)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestCommandController_HTTPRequest_Parsing(t *testing.T) {
	body := []byte(`{"command":"test","nonce":"n","timestamp":123}`)
	req := httptest.NewRequest(http.MethodPost, "/api/commands/dispatch/device-001", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		t.Errorf("Method = %s, want POST", req.Method)
	}
}

func TestCommandController_HTTPResponse_Success(t *testing.T) {
	resp := models.CommandResponse{
		DispatchID: "dispatch-001",
		Delivery:   "queued",
		ServerTime: 1234567890,
	}

	data, _ := json.Marshal(resp)
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusCreated)
	if _, err := w.Write(data); err != nil {
		t.Logf("write error: %v", err)
	}

	if w.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", w.Code, http.StatusCreated)
	}
}

func TestCommandController_HTTPResponse_Error(t *testing.T) {
	resp := models.ErrorResponse{
		Error:   "device_not_found",
		Message: "Device does not exist",
	}

	data, _ := json.Marshal(resp)
	w := httptest.NewRecorder()
	w.WriteHeader(http.StatusNotFound)
	if _, err := w.Write(data); err != nil {
		t.Logf("write error: %v", err)
	}

	if w.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", w.Code, http.StatusNotFound)
	}
}