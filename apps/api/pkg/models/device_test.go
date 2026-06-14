package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestDevice_JSON(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Second)
	device := Device{
		RegisteredAt:      now,
		LastSeen:          now,
		ID:                "device_abc123",
		FirebaseInstallID: "firebase-id-xyz",
		FCMToken:          "fcm-token-123",
		AppVersion:        "1.2.3",
		DeviceClass:       "phone",
		CommandSecret:     "cmd-secret-456",
		Online:            true,
	}

	data, err := json.Marshal(device)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Device
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ID != device.ID {
		t.Errorf("ID = %q, want %q", unmarshaled.ID, device.ID)
	}
	if unmarshaled.AppVersion != device.AppVersion {
		t.Errorf("AppVersion = %q, want %q", unmarshaled.AppVersion, device.AppVersion)
	}
	if unmarshaled.Online != device.Online {
		t.Errorf("Online = %v, want %v", unmarshaled.Online, device.Online)
	}
}

func TestDevice_JSONTags(t *testing.T) {
	data := []byte(`{
		"registeredAt": "2024-01-01T00:00:00Z",
		"lastSeen": "2024-01-15T12:30:00Z",
		"deviceId": "dev_test",
		"firebaseInstallId": "fi_test",
		"fcmToken": "fcm_test",
		"appVersion": "2.0.0",
		"deviceClass": "tablet",
		"commandSecret": "cmd_test",
		"online": true
	}`)

	var device Device
	if err := json.Unmarshal(data, &device); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if device.ID != "dev_test" {
		t.Errorf("ID = %q, want \"dev_test\"", device.ID)
	}
	if device.DeviceClass != "tablet" {
		t.Errorf("DeviceClass = %q, want \"tablet\"", device.DeviceClass)
	}
}

func TestRegisterRequest_JSON(t *testing.T) {
	req := RegisterRequest{
		DeviceID:          "device_new",
		FirebaseInstallID: "firebase_new",
		FCMToken:          "fcm_new",
		AppVersion:        "1.0.0",
		DeviceClass:       "auto",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled RegisterRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.DeviceID != req.DeviceID {
		t.Errorf("DeviceID = %q, want %q", unmarshaled.DeviceID, req.DeviceID)
	}
	if unmarshaled.AppVersion != req.AppVersion {
		t.Errorf("AppVersion = %q, want %q", unmarshaled.AppVersion, req.AppVersion)
	}
}

func TestRegisterResponse_JSON(t *testing.T) {
	resp := RegisterResponse{
		DeviceID:      "device_reg",
		CommandSecret: "secret_xyz",
		RegisteredAt:  1700000000000,
		ServerTime:    1700000060000,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled RegisterResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.DeviceID != resp.DeviceID {
		t.Errorf("DeviceID = %q, want %q", unmarshaled.DeviceID, resp.DeviceID)
	}
	if unmarshaled.CommandSecret != resp.CommandSecret {
		t.Errorf("CommandSecret = %q, want %q", unmarshaled.CommandSecret, resp.CommandSecret)
	}
}

func TestDeviceStatus_JSON(t *testing.T) {
	status := DeviceStatus{
		DeviceID:          "device_status",
		AppVersion:        "3.0.0",
		DeviceClass:       "watch",
		FirebaseInstallID: "fi_status",
		FCMToken:          "fcm_status",
		LastSeen:          1700000000000,
		Online:            false,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled DeviceStatus
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.DeviceID != status.DeviceID {
		t.Errorf("DeviceID = %q, want %q", unmarshaled.DeviceID, status.DeviceID)
	}
	if unmarshaled.Online != status.Online {
		t.Errorf("Online = %v, want %v", unmarshaled.Online, status.Online)
	}
}

func TestDeviceStatus_OptionalFields(t *testing.T) {
	// Only required fields
	status := DeviceStatus{
		DeviceID: "device_min",
		LastSeen: 1700000000000,
		Online:   true,
	}

	data, err := json.Marshal(status)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled DeviceStatus
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.FirebaseInstallID != "" {
		t.Errorf("FirebaseInstallID = %q, want \"\"", unmarshaled.FirebaseInstallID)
	}
	if unmarshaled.FCMToken != "" {
		t.Errorf("FCMToken = %q, want \"\"", unmarshaled.FCMToken)
	}
}

func TestUpdateDeviceRequest_JSON(t *testing.T) {
	newToken := "new-fcm-token"
	newVersion := "4.0.0"

	req := UpdateDeviceRequest{
		FCMToken:   &newToken,
		AppVersion: &newVersion,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled UpdateDeviceRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.FCMToken == nil || *unmarshaled.FCMToken != newToken {
		t.Errorf("FCMToken = %v, want %q", unmarshaled.FCMToken, newToken)
	}
	if unmarshaled.AppVersion == nil || *unmarshaled.AppVersion != newVersion {
		t.Errorf("AppVersion = %v, want %q", unmarshaled.AppVersion, newVersion)
	}
}

func TestUpdateDeviceRequest_PartialUpdate(t *testing.T) {
	// Only update FCM token
	newToken := "token-only"
	req := UpdateDeviceRequest{
		FCMToken: &newToken,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled UpdateDeviceRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.FCMToken == nil {
		t.Error("FCMToken should not be nil")
	}
	if unmarshaled.AppVersion != nil {
		t.Error("AppVersion should be nil")
	}
}

func TestUpdateDeviceRequest_EmptyPointers(t *testing.T) {
	req := UpdateDeviceRequest{}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled UpdateDeviceRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.FCMToken != nil {
		t.Error("FCMToken should be nil")
	}
	if unmarshaled.AppVersion != nil {
		t.Error("AppVersion should be nil")
	}
}
