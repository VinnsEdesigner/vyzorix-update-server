package models

import (
	"encoding/json"
	"testing"
	"time"
)

func TestOperatorRole_constants(t *testing.T) {
	if RoleSuperAdmin != "super_admin" {
		t.Errorf("RoleSuperAdmin = %s, want super_admin", RoleSuperAdmin)
	}
	if RoleOperator != "operator" {
		t.Errorf("RoleOperator = %s, want operator", RoleOperator)
	}
	if RoleViewer != "viewer" {
		t.Errorf("RoleViewer = %s, want viewer", RoleViewer)
	}
}

func TestOperator_JSONMarshal(t *testing.T) {
	op := Operator{
		ID:           "op-001",
		Email:        "test@example.com",
		Name:         "Test User",
		PasswordHash: "secret-hash",
		Role:         RoleOperator,
		GoogleID:     "google-123",
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	data, err := json.Marshal(op)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	// PasswordHash should be omitted (json:"-")
	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if _, ok := result["passwordHash"]; ok {
		t.Error("passwordHash should not be in JSON output")
	}
	if result["id"] != "op-001" {
		t.Errorf("id = %v, want op-001", result["id"])
	}
	if result["email"] != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", result["email"])
	}
	if result["role"] != "operator" {
		t.Errorf("role = %v, want operator", result["role"])
	}
}

func TestOperator_ToResponse(t *testing.T) {
	now := time.Now()
	op := Operator{
		ID:        "op-001",
		Email:     "test@example.com",
		Name:      "Test User",
		Role:      RoleOperator,
		CreatedAt: now,
		UpdatedAt: now,
	}

	resp := op.ToResponse()

	if resp.ID != "op-001" {
		t.Errorf("ID = %s, want op-001", resp.ID)
	}
	if resp.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", resp.Email)
	}
	if resp.Role != RoleOperator {
		t.Errorf("Role = %s, want operator", resp.Role)
	}
	if resp.CreatedAt != now.UnixMilli() {
		t.Errorf("CreatedAt = %d, want %d", resp.CreatedAt, now.UnixMilli())
	}
}

func TestSession_Fields(t *testing.T) {
	now := time.Now()
	sess := Session{
		ID:         "sess-001",
		OperatorID: "op-001",
		TokenHash:  "hash123",
		ExpiresAt:  now.Add(24 * time.Hour),
		CreatedAt:  now,
	}

	if sess.ID != "sess-001" {
		t.Errorf("ID = %s, want sess-001", sess.ID)
	}
	if sess.OperatorID != "op-001" {
		t.Errorf("OperatorID = %s, want op-001", sess.OperatorID)
	}
	if sess.TokenHash != "hash123" {
		t.Errorf("TokenHash = %s, want hash123", sess.TokenHash)
	}
	if sess.ExpiresAt.Before(now) {
		t.Error("ExpiresAt should be in the future")
	}
}

func TestSession_TokenHashOmitempty(t *testing.T) {
	sess := Session{
		ID:         "sess-001",
		OperatorID: "op-001",
		ExpiresAt:  time.Now().Add(24 * time.Hour),
		CreatedAt:  time.Now(),
	}

	data, err := json.Marshal(sess)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if _, ok := result["tokenHash"]; ok {
		t.Error("tokenHash should not be in JSON output")
	}
}

func TestDevice_Fields(t *testing.T) {
	now := time.Now()
	d := Device{
		ID:                "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "fcm-token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
		CommandSecret:     "secret123",
		Online:            true,
		RegisteredAt:      now,
		LastSeen:          now,
	}

	if d.ID != "device-001" {
		t.Errorf("ID = %s, want device-001", d.ID)
	}
	if d.CommandSecret != "secret123" {
		t.Errorf("CommandSecret = %s, want secret123", d.CommandSecret)
	}
	if !d.Online {
		t.Error("Online should be true")
	}
}

func TestRegisterRequest_Fields(t *testing.T) {
	req := RegisterRequest{
		DeviceID:          "device-001",
		FirebaseInstallID: "firebase-abc",
		FCMToken:          "fcm-token",
		AppVersion:        "1.0.0",
		DeviceClass:       "phone",
	}

	if req.DeviceID != "device-001" {
		t.Errorf("DeviceID = %s, want device-001", req.DeviceID)
	}
	if req.FirebaseInstallID != "firebase-abc" {
		t.Errorf("FirebaseInstallID = %s, want firebase-abc", req.FirebaseInstallID)
	}
	if req.FCMToken != "fcm-token" {
		t.Errorf("FCMToken = %s, want fcm-token", req.FCMToken)
	}
	if req.AppVersion != "1.0.0" {
		t.Errorf("AppVersion = %s, want 1.0.0", req.AppVersion)
	}
	if req.DeviceClass != "phone" {
		t.Errorf("DeviceClass = %s, want phone", req.DeviceClass)
	}
}

func TestRegisterRequest_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
		"deviceId": "device-002",
		"firebaseInstallId": "firebase-xyz",
		"fcmToken": "new-token",
		"appVersion": "2.0.0",
		"deviceClass": "tablet"
	}`)

	var req RegisterRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if req.DeviceID != "device-002" {
		t.Errorf("DeviceID = %s, want device-002", req.DeviceID)
	}
	if req.FirebaseInstallID != "firebase-xyz" {
		t.Errorf("FirebaseInstallID = %s, want firebase-xyz", req.FirebaseInstallID)
	}
	if req.FCMToken != "new-token" {
		t.Errorf("FCMToken = %s, want new-token", req.FCMToken)
	}
	if req.AppVersion != "2.0.0" {
		t.Errorf("AppVersion = %s, want 2.0.0", req.AppVersion)
	}
	if req.DeviceClass != "tablet" {
		t.Errorf("DeviceClass = %s, want tablet", req.DeviceClass)
	}
}

func TestCommandFrame_Fields(t *testing.T) {
	frame := CommandFrame{
		Type:       "update",
		DispatchID: "dispatch-001",
		Args:       []byte(`{"version":"2.0.0"}`),
	}

	if frame.Type != "update" {
		t.Errorf("Type = %s, want update", frame.Type)
	}
	if frame.DispatchID != "dispatch-001" {
		t.Errorf("DispatchID = %s, want dispatch-001", frame.DispatchID)
	}
	if string(frame.Args) != `{"version":"2.0.0"}` {
		t.Errorf("Args = %s, want {\"version\":\"2.0.0\"}", string(frame.Args))
	}
}

func TestCommandFrame_JSONMarshal(t *testing.T) {
	frame := CommandFrame{
		Type:       "update",
		DispatchID: "dispatch-001",
		Args:       []byte(`{"version":"2.0.0"}`),
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
}

func TestTelemetryFrame_Fields(t *testing.T) {
	frame := TelemetryFrame{
		DeviceID:    "device-001",
		RiskScore:   25,
		BufferLevel: 60,
		ThermalTemp: 35.5,
		Raw:         []byte(`{"test":true}`),
	}

	if frame.DeviceID != "device-001" {
		t.Errorf("DeviceID = %s, want device-001", frame.DeviceID)
	}
	if frame.RiskScore != 25 {
		t.Errorf("RiskScore = %d, want 25", frame.RiskScore)
	}
	if frame.BufferLevel != 60 {
		t.Errorf("BufferLevel = %d, want 60", frame.BufferLevel)
	}
	if frame.ThermalTemp != 35.5 {
		t.Errorf("ThermalTemp = %f, want 35.5", frame.ThermalTemp)
	}
}

func TestTelemetryFrame_JSONUnmarshal(t *testing.T) {
	data := []byte(`{
		"deviceId": "device-001",
		"riskScore": 30,
		"bufferLevel": 75,
		"thermalTemp": 40.2
	}`)

	var frame TelemetryFrame
	if err := json.Unmarshal(data, &frame); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if frame.DeviceID != "device-001" {
		t.Errorf("DeviceID = %s, want device-001", frame.DeviceID)
	}
	if frame.RiskScore != 30 {
		t.Errorf("RiskScore = %d, want 30", frame.RiskScore)
	}
	if frame.BufferLevel != 75 {
		t.Errorf("BufferLevel = %d, want 75", frame.BufferLevel)
	}
	if frame.ThermalTemp != 40.2 {
		t.Errorf("ThermalTemp = %f, want 40.2", frame.ThermalTemp)
	}
}

func TestErrorResponse_Fields(t *testing.T) {
	resp := ErrorResponse{
		Error:   "unauthorized",
		Message: "Access denied",
	}

	if resp.Error != "unauthorized" {
		t.Errorf("Error = %s, want unauthorized", resp.Error)
	}
	if resp.Message != "Access denied" {
		t.Errorf("Message = %s, want Access denied", resp.Message)
	}
}

func TestErrorResponse_JSONMarshal(t *testing.T) {
	resp := ErrorResponse{
		Error:   "not_found",
		Message: "Resource not found",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["error"] != "not_found" {
		t.Errorf("error = %v, want not_found", result["error"])
	}
	if result["message"] != "Resource not found" {
		t.Errorf("message = %v, want Resource not found", result["message"])
	}
}

func TestOKResponse_Fields(t *testing.T) {
	resp := OKResponse{
		OK:         true,
		Database:   "sqlite",
		ServerTime: time.Now().UnixMilli(),
	}

	if !resp.OK {
		t.Error("OK should be true")
	}
	if resp.Database != "sqlite" {
		t.Errorf("Database = %s, want sqlite", resp.Database)
	}
}

func TestLoginRequest_Fields(t *testing.T) {
	req := LoginRequest{
		Email:    "test@example.com",
		Password: "secret123",
	}

	if req.Email != "test@example.com" {
		t.Errorf("Email = %s, want test@example.com", req.Email)
	}
	if req.Password != "secret123" {
		t.Errorf("Password = %s, want secret123", req.Password)
	}
}

func TestOperatorRegisterRequest_Fields(t *testing.T) {
	req := OperatorRegisterRequest{
		Email:    "new@example.com",
		Password: "newpassword",
		Name:     "New User",
	}

	if req.Email != "new@example.com" {
		t.Errorf("Email = %s, want new@example.com", req.Email)
	}
	if req.Password != "newpassword" {
		t.Errorf("Password = %s, want newpassword", req.Password)
	}
	if req.Name != "New User" {
		t.Errorf("Name = %s, want New User", req.Name)
	}
}

func TestAuthResponse_Fields(t *testing.T) {
	resp := AuthResponse{
		Token: "jwt-token-here",
		Operator: OperatorResponse{
			ID:    "op-001",
			Email: "test@example.com",
			Name:  "Test User",
			Role:  RoleOperator,
		},
	}

	if resp.Token != "jwt-token-here" {
		t.Errorf("Token = %s, want jwt-token-here", resp.Token)
	}
	if resp.Operator.ID != "op-001" {
		t.Errorf("Operator.ID = %s, want op-001", resp.Operator.ID)
	}
}

func TestAuthResponse_Omitempty(t *testing.T) {
	// Operator can be empty for token-only responses
	resp := AuthResponse{
		Token: "token-only",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["token"] != "token-only" {
		t.Errorf("token = %v, want token-only", result["token"])
	}
}

func TestCommandRequest_Fields(t *testing.T) {
	req := CommandRequest{
		Command:   "update",
		Args:      []byte(`{"version":"2.0.0"}`),
		Nonce:     "nonce123",
		Timestamp: 1234567890,
	}

	if req.Command != "update" {
		t.Errorf("Command = %s, want update", req.Command)
	}
	if string(req.Args) != `{"version":"2.0.0"}` {
		t.Errorf("Args = %s, want {\"version\":\"2.0.0\"}", string(req.Args))
	}
	if req.Nonce != "nonce123" {
		t.Errorf("Nonce = %s, want nonce123", req.Nonce)
	}
	if req.Timestamp != 1234567890 {
		t.Errorf("Timestamp = %d, want 1234567890", req.Timestamp)
	}
}

func TestCommandResponse_Fields(t *testing.T) {
	resp := CommandResponse{
		DispatchID: "dispatch-001",
		Delivery:   "queued",
		ServerTime: 1234567890,
	}

	if resp.DispatchID != "dispatch-001" {
		t.Errorf("DispatchID = %s, want dispatch-001", resp.DispatchID)
	}
	if resp.Delivery != "queued" {
		t.Errorf("Delivery = %s, want queued", resp.Delivery)
	}
	if resp.ServerTime != 1234567890 {
		t.Errorf("ServerTime = %d, want 1234567890", resp.ServerTime)
	}
}

func TestVersionManifest_Fields(t *testing.T) {
	vm := VersionManifest{
		Version:      "2.0.0",
		VersionCode:  200,
		APKFilename:  "vyzorix-2.0.0.apk",
		APKSHA256:    "abc123def456",
		APKSizeBytes: 10485760,
		ReleaseNotes: "Bug fixes and improvements",
	}

	if vm.Version != "2.0.0" {
		t.Errorf("Version = %s, want 2.0.0", vm.Version)
	}
	if vm.VersionCode != 200 {
		t.Errorf("VersionCode = %d, want 200", vm.VersionCode)
	}
	if vm.APKFilename != "vyzorix-2.0.0.apk" {
		t.Errorf("APKFilename = %s, want vyzorix-2.0.0.apk", vm.APKFilename)
	}
	if vm.APKSHA256 != "abc123def456" {
		t.Errorf("APKSHA256 = %s, want abc123def456", vm.APKSHA256)
	}
	if vm.APKSizeBytes != 10485760 {
		t.Errorf("APKSizeBytes = %d, want 10485760", vm.APKSizeBytes)
	}
}
