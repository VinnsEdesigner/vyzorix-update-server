package models

import (
	"encoding/json"
	"strings"
	"testing"
	"time"
)

func TestOperator_ToResponse_ThresholdsAndClient(t *testing.T) {
	now := time.Now().UTC().Truncate(time.Millisecond)
	op := &Operator{
		ID:            "op_123",
		Email:         "test@example.com",
		Name:          "Test User",
		Role:          RoleOperator,
		PasswordHash:  "secret-hash-should-not-appear",
		EmailVerified: true,
		CreatedAt:     now,
		UpdatedAt:     now,
		Thresholds: Thresholds{
			RiskWarn:    50,
			RiskCrit:    80,
			ThermalWarn: 70,
			ThermalCrit: 90,
			BufferWarn:  30,
			BufferCrit:  10,
		},
		Client: ClientSettings{
			StrictHmac:           true,
			AutoReconnect:        false,
			NotificationsEnabled: true,
		},
	}

	resp := op.ToResponse()

	if resp.ID != op.ID {
		t.Errorf("ID = %q, want %q", resp.ID, op.ID)
	}
	if resp.Email != op.Email {
		t.Errorf("Email = %q, want %q", resp.Email, op.Email)
	}
	if resp.Name != op.Name {
		t.Errorf("Name = %q, want %q", resp.Name, op.Name)
	}
	if resp.Role != op.Role {
		t.Errorf("Role = %v, want %v", resp.Role, op.Role)
	}
	if resp.EmailVerified != op.EmailVerified {
		t.Errorf("EmailVerified = %v, want %v", resp.EmailVerified, op.EmailVerified)
	}
	if resp.CreatedAt != op.CreatedAt.UnixMilli() {
		t.Errorf("CreatedAt = %d, want %d", resp.CreatedAt, op.CreatedAt.UnixMilli())
	}
	if resp.Thresholds == nil {
		t.Error("Thresholds should not be nil")
	} else {
		if resp.Thresholds.RiskWarn != op.Thresholds.RiskWarn {
			t.Errorf("Thresholds.RiskWarn = %d, want %d", resp.Thresholds.RiskWarn, op.Thresholds.RiskWarn)
		}
	}
	if resp.Client == nil {
		t.Error("Client should not be nil")
	} else {
		if resp.Client.StrictHmac != op.Client.StrictHmac {
			t.Errorf("Client.StrictHmac = %v, want %v", resp.Client.StrictHmac, op.Client.StrictHmac)
		}
	}
}

func TestOperator_ToResponse_JSONSerialization(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC)
	op := &Operator{
		ID:            "op_abc",
		Email:         "user@test.com",
		Name:          "JSON Test",
		Role:          RoleSuperAdmin,
		EmailVerified: true,
		CreatedAt:     now,
		Thresholds: Thresholds{
			RiskWarn: 40,
		},
	}

	resp := op.ToResponse()

	// Marshal to JSON and back
	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled OperatorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ID != resp.ID {
		t.Errorf("unmarshaled.ID = %q, want %q", unmarshaled.ID, resp.ID)
	}
	if unmarshaled.Email != resp.Email {
		t.Errorf("unmarshaled.Email = %q, want %q", unmarshaled.Email, resp.Email)
	}

	// Verify JSON doesn't contain password hash
	dataStr := string(data)
	if strings.Contains(dataStr, "secret-hash") || strings.Contains(dataStr, "PasswordHash") {
		t.Error("JSON should not contain PasswordHash field")
	}
}

func TestOperatorRole_Constants(t *testing.T) {
	if RoleViewer != "viewer" {
		t.Errorf("RoleViewer = %q, want \"viewer\"", RoleViewer)
	}
	if RoleOperator != "operator" {
		t.Errorf("RoleOperator = %q, want \"operator\"", RoleOperator)
	}
	if RoleSuperAdmin != "super_admin" {
		t.Errorf("RoleSuperAdmin = %q, want \"super_admin\"", RoleSuperAdmin)
	}
}

func TestOperatorRole_IsValid(t *testing.T) {
	validRoles := []OperatorRole{RoleViewer, RoleOperator, RoleSuperAdmin}
	for _, role := range validRoles {
		if role == "" {
			t.Errorf("role %q should not be empty", role)
		}
	}

	// Test JSON marshaling
	for _, role := range validRoles {
		data, err := json.Marshal(role)
		if err != nil {
			t.Errorf("json.Marshal(%q) error = %v", role, err)
		}
		var unmarshaled OperatorRole
		if err := json.Unmarshal(data, &unmarshaled); err != nil {
			t.Errorf("json.Unmarshal() error = %v", err)
		}
		if unmarshaled != role {
			t.Errorf("round-trip failed for %q", role)
		}
	}
}

func TestThresholds_Defaults(t *testing.T) {
	thresholds := Thresholds{}

	// All fields should be zero by default
	if thresholds.RiskWarn != 0 {
		t.Errorf("RiskWarn = %d, want 0", thresholds.RiskWarn)
	}
	if thresholds.RiskCrit != 0 {
		t.Errorf("RiskCrit = %d, want 0", thresholds.RiskCrit)
	}
}

func TestThresholds_JSON(t *testing.T) {
	thresholds := Thresholds{
		RiskWarn:    50,
		RiskCrit:    80,
		ThermalWarn: 70,
		ThermalCrit: 90,
		BufferWarn:  30,
		BufferCrit:  10,
	}

	data, err := json.Marshal(thresholds)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled Thresholds
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.RiskWarn != thresholds.RiskWarn {
		t.Errorf("RiskWarn = %d, want %d", unmarshaled.RiskWarn, thresholds.RiskWarn)
	}
}

func TestClientSettings_Defaults(t *testing.T) {
	settings := ClientSettings{}

	if settings.StrictHmac != false {
		t.Error("StrictHmac should default to false")
	}
	if settings.AutoReconnect != false {
		t.Error("AutoReconnect should default to false")
	}
	if settings.NotificationsEnabled != false {
		t.Error("NotificationsEnabled should default to false")
	}
}

func TestClientSettings_JSON(t *testing.T) {
	settings := ClientSettings{
		StrictHmac:           true,
		AutoReconnect:        true,
		NotificationsEnabled: false,
	}

	data, err := json.Marshal(settings)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ClientSettings
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.StrictHmac != settings.StrictHmac {
		t.Errorf("StrictHmac = %v, want %v", unmarshaled.StrictHmac, settings.StrictHmac)
	}
}

func TestLoginRequest_JSON(t *testing.T) {
	req := LoginRequest{
		Email:    "test@example.com",
		Password: "secretpassword",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled LoginRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Email != req.Email {
		t.Errorf("Email = %q, want %q", unmarshaled.Email, req.Email)
	}
	if unmarshaled.Password != req.Password {
		t.Errorf("Password = %q, want %q", unmarshaled.Password, req.Password)
	}
}

func TestOperatorRegisterRequest_JSON(t *testing.T) {
	req := OperatorRegisterRequest{
		Email:    "new@example.com",
		Password: "newpassword123",
		Name:     "New User",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled OperatorRegisterRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Email != req.Email {
		t.Errorf("Email = %q, want %q", unmarshaled.Email, req.Email)
	}
	if unmarshaled.Name != req.Name {
		t.Errorf("Name = %q, want %q", unmarshaled.Name, req.Name)
	}
}

func TestAuthResponse_JSON(t *testing.T) {
	resp := AuthResponse{
		Token:     "jwt-token-here",
		ExpiresAt: 1700000000000,
		Operator: OperatorResponse{
			ID:        "op_123",
			Email:     "test@example.com",
			Name:      "Test",
			Role:      RoleOperator,
			CreatedAt: 1699990000000,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled AuthResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Token != resp.Token {
		t.Errorf("Token = %q, want %q", unmarshaled.Token, resp.Token)
	}
	if unmarshaled.ExpiresAt != resp.ExpiresAt {
		t.Errorf("ExpiresAt = %d, want %d", unmarshaled.ExpiresAt, resp.ExpiresAt)
	}
}

func TestSession_JSON(t *testing.T) {
	now := time.Now().UTC()
	session := Session{
		ID:         "sess_123",
		OperatorID: "op_456",
		TokenHash:  "should-not-appear",
		ExpiresAt:  now.Add(24 * time.Hour),
		CreatedAt:  now,
		UserAgent:  "TestAgent/1.0",
		IPAddress:  "192.168.1.1",
	}

	data, err := json.Marshal(session)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	// TokenHash should not appear in JSON (has json:"-" tag)
	if strings.Contains(string(data), "should-not-appear") {
		t.Error("Session JSON should not contain TokenHash")
	}

	var unmarshaled Session
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.ID != session.ID {
		t.Errorf("ID = %q, want %q", unmarshaled.ID, session.ID)
	}
	if unmarshaled.OperatorID != session.OperatorID {
		t.Errorf("OperatorID = %q, want %q", unmarshaled.OperatorID, session.OperatorID)
	}
}

func TestUpdateNameRequest_JSON(t *testing.T) {
	name := "New Name"
	req := UpdateNameRequest{
		Name: &name,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled UpdateNameRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Name == nil {
		t.Fatal("Name should not be nil")
	}
	if *unmarshaled.Name != name {
		t.Errorf("Name = %q, want %q", *unmarshaled.Name, name)
	}
}

func TestUpdateNameRequest_NilName(t *testing.T) {
	req := UpdateNameRequest{
		Name: nil,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled UpdateNameRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Name != nil {
		t.Error("Name should be nil")
	}
}

func TestUpdateSettingsRequest_JSON(t *testing.T) {
	thresholds := &Thresholds{RiskWarn: 60}
	client := &ClientSettings{StrictHmac: true}
	name := "Updated Name"

	req := UpdateSettingsRequest{
		Name:       &name,
		Thresholds: thresholds,
		Client:     client,
		Reset:      false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled UpdateSettingsRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Name == nil || *unmarshaled.Name != name {
		t.Errorf("Name = %v, want %q", unmarshaled.Name, name)
	}
	if unmarshaled.Thresholds == nil {
		t.Error("Thresholds should not be nil")
	}
	if unmarshaled.Client == nil {
		t.Error("Client should not be nil")
	}
}

func TestUpdateSettingsRequest_Reset(t *testing.T) {
	req := UpdateSettingsRequest{
		Reset: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled UpdateSettingsRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if !unmarshaled.Reset {
		t.Error("Reset should be true")
	}
}

func TestVerifyEmailRequest_JSON(t *testing.T) {
	req := VerifyEmailRequest{
		Token: "verify-token-abc123",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled VerifyEmailRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Token != req.Token {
		t.Errorf("Token = %q, want %q", unmarshaled.Token, req.Token)
	}
}

func TestForgotPasswordRequest_JSON(t *testing.T) {
	req := ForgotPasswordRequest{
		Email: "forgot@example.com",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ForgotPasswordRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Email != req.Email {
		t.Errorf("Email = %q, want %q", unmarshaled.Email, req.Email)
	}
}

func TestResetPasswordRequest_JSON(t *testing.T) {
	req := ResetPasswordRequest{
		Token:       "reset-token-xyz",
		NewPassword: "NewSecurePass123!",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled ResetPasswordRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Token != req.Token {
		t.Errorf("Token = %q, want %q", unmarshaled.Token, req.Token)
	}
	if unmarshaled.NewPassword != req.NewPassword {
		t.Errorf("NewPassword = %q, want %q", unmarshaled.NewPassword, req.NewPassword)
	}
}

func TestMessageResponse_JSON(t *testing.T) {
	resp := MessageResponse{
		Message: "Operation completed successfully",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled MessageResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Message != resp.Message {
		t.Errorf("Message = %q, want %q", unmarshaled.Message, resp.Message)
	}
}

func TestEmailVerifiedResponse_JSON(t *testing.T) {
	resp := EmailVerifiedResponse{
		Email:    "verified@example.com",
		Verified: true,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled EmailVerifiedResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Email != resp.Email {
		t.Errorf("Email = %q, want %q", unmarshaled.Email, resp.Email)
	}
	if unmarshaled.Verified != resp.Verified {
		t.Errorf("Verified = %v, want %v", unmarshaled.Verified, resp.Verified)
	}
}

func TestGoogleOAuthURLRequest_JSON(t *testing.T) {
	req := GoogleOAuthURLRequest{}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled GoogleOAuthURLRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}
}

func TestGoogleOAuthURLResponse_JSON(t *testing.T) {
	resp := GoogleOAuthURLResponse{
		URL: "https://accounts.google.com/o/oauth2/auth?...",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled GoogleOAuthURLResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.URL != resp.URL {
		t.Errorf("URL = %q, want %q", unmarshaled.URL, resp.URL)
	}
}

func TestGoogleOAuthCallbackRequest_JSON(t *testing.T) {
	req := GoogleOAuthCallbackRequest{
		Code:  "authorization-code-from-google",
		State: "https://app.example.com/dashboard",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled GoogleOAuthCallbackRequest
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Code != req.Code {
		t.Errorf("Code = %q, want %q", unmarshaled.Code, req.Code)
	}
	if unmarshaled.State != req.State {
		t.Errorf("State = %q, want %q", unmarshaled.State, req.State)
	}
}

func TestAuthErrorResponse_JSON(t *testing.T) {
	resp := AuthErrorResponse{
		Error:   "invalid_credentials",
		Message: "The email or password you entered is incorrect.",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() error = %v", err)
	}

	var unmarshaled AuthErrorResponse
	if err := json.Unmarshal(data, &unmarshaled); err != nil {
		t.Fatalf("json.Unmarshal() error = %v", err)
	}

	if unmarshaled.Error != resp.Error {
		t.Errorf("Error = %q, want %q", unmarshaled.Error, resp.Error)
	}
	if unmarshaled.Message != resp.Message {
		t.Errorf("Message = %q, want %q", unmarshaled.Message, resp.Message)
	}
}
