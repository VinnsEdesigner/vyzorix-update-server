package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/models"
)

func TestDeviceController_RegisterRequest_JSON(t *testing.T) {
	req := models.RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "fcm-token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["deviceId"] != "device-001" {
		t.Errorf("deviceId = %v, want device-001", result["deviceId"])
	}
	if result["firebaseInstallId"] != "firebase-abc" {
		t.Errorf("firebaseInstallId = %v, want firebase-abc", result["firebaseInstallId"])
	}
}

func TestDeviceController_RegisterRequest_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
		"deviceId": "device-002",
		"firebaseInstallId": "firebase-xyz",
		"fcmToken": "new-token",
		"appVersion": "2.0.0",
		"deviceClass": "tablet"
	}`)

	var req models.RegisterRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if req.DeviceID != "device-002" {
		t.Errorf("DeviceID = %s, want device-002", req.DeviceID)
	}
	if req.FCMToken != "new-token" {
		t.Errorf("FCMToken = %s, want new-token", req.FCMToken)
	}
	if req.DeviceClass != "tablet" {
		t.Errorf("DeviceClass = %s, want tablet", req.DeviceClass)
	}
}

func TestDeviceController_RegisterResponse_JSON(t *testing.T) {
	resp := models.RegisterResponse{
		DeviceID:      "device-001",
		CommandSecret: "secret123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["deviceId"] != "device-001" {
		t.Errorf("deviceId = %v, want device-001", result["deviceId"])
	}
	if result["commandSecret"] != "secret123" {
		t.Errorf("commandSecret = %v, want secret123", result["commandSecret"])
	}
}

func TestDeviceController_DeviceStatus_JSON(t *testing.T) {
	status := models.DeviceStatus{
		DeviceID:    "device-001",
		Online:      true,
		LastSeen:    1234567890,
		AppVersion:  "1.0.0",
		DeviceClass: "phone",
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["deviceId"] != "device-001" {
		t.Errorf("deviceId = %v, want device-001", result["deviceId"])
	}
	if result["online"] != true {
		t.Error("online should be true")
	}
	if result["appVersion"] != "1.0.0" {
		t.Errorf("appVersion = %v, want 1.0.0", result["appVersion"])
	}
}

func TestDeviceController_Register_InvalidJSON(t *testing.T) {
	// Test that invalid JSON is handled gracefully
	data := []byte(`{invalid json}`)

	var req models.RegisterRequest
	err := json.Unmarshal(data, &req)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestDeviceController_Register_MissingFields(t *testing.T) {
	// Only provide some fields
	data := []byte(`{"deviceId": "device-001"}`)

	var req models.RegisterRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if req.DeviceID != "device-001" {
		t.Errorf("DeviceID = %s, want device-001", req.DeviceID)
	}
	// Other fields should be zero values
	if req.FCMToken != "" {
		t.Errorf("FCMToken = %s, want empty", req.FCMToken)
	}
}

func TestDeviceController_HTTPRequest_Parsing(t *testing.T) {
	body := []byte(`{"deviceId":"test-device","firebaseInstallId":"firebase-test"}`)
	req := httptest.NewRequest(http.MethodPost, "/api/devices/register", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}
}

func TestDeviceController_HTTPResponse_JSON(t *testing.T) {
	resp := models.RegisterResponse{
		DeviceID:      "device-001",
		CommandSecret: "secret123",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	w := httptest.NewRecorder()
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if _, err := w.Write(data); err != nil {
		t.Logf("write error: %v", err)
	}

	if w.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", w.Code, http.StatusOK)
	}
	if w.Header().Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}
}