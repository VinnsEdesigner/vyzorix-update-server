package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/pkg/models"
)

func TestAuthController_LoginRequest_JSON(t *testing.T) {
	req := models.LoginRequest{
		Email:    "test@example.com",
		Password: "secret123",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["email"] != "test@example.com" {
		t.Errorf("email = %v, want test@example.com", result["email"])
	}
	if result["password"] != "secret123" {
		t.Errorf("password = %v, want secret123", result["password"])
	}
}

func TestAuthController_LoginRequest_JSONUnmarshal(t *testing.T) {
	data := []byte(`{"email":"user@example.com","password":"pass123"}`)

	var req models.LoginRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if req.Email != "user@example.com" {
		t.Errorf("Email = %s, want user@example.com", req.Email)
	}
	if req.Password != "pass123" {
		t.Errorf("Password = %s, want pass123", req.Password)
	}
}

func TestAuthController_RegisterRequest_JSON(t *testing.T) {
	req := models.OperatorRegisterRequest{
		Email:    "new@example.com",
		Password: "newpassword",
		Name:     "New User",
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["email"] != "new@example.com" {
		t.Errorf("email = %v, want new@example.com", result["email"])
	}
	if result["name"] != "New User" {
		t.Errorf("name = %v, want New User", result["name"])
	}
}

func TestAuthController_AuthResponse_JSON(t *testing.T) {
	resp := models.AuthResponse{
		Token: "jwt-token-abc123",
		Operator: models.OperatorResponse{
			ID:    "op-001",
			Email: "test@example.com",
			Name:  "Test User",
			Role:  models.RoleOperator,
		},
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["token"] != "jwt-token-abc123" {
		t.Errorf("token = %v, want jwt-token-abc123", result["token"])
	}
	if result["operator"] == nil {
		t.Error("operator should not be nil")
	}
}

func TestAuthController_ErrorResponse_JSON(t *testing.T) {
	resp := models.ErrorResponse{
		Error:   "unauthorized",
		Message: "Invalid credentials",
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["error"] != "unauthorized" {
		t.Errorf("error = %v, want unauthorized", result["error"])
	}
	if result["message"] != "Invalid credentials" {
		t.Errorf("message = %v, want Invalid credentials", result["message"])
	}
}

func strPtr(s string) *string { return &s }

func TestAuthController_UpdateNameRequest_JSON(t *testing.T) {
	req := models.UpdateNameRequest{
		Name: strPtr("Updated Name"),
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["name"] != "Updated Name" {
		t.Errorf("name = %v, want Updated Name", result["name"])
	}
}

func TestAuthController_UpdateSettingsRequest_JSON(t *testing.T) {
	thresholds := models.Thresholds{
		RiskWarn:    40,
		RiskCrit:    70,
		ThermalWarn: 42,
		ThermalCrit: 52,
		BufferWarn:  45,
		BufferCrit:  75,
	}
	client := models.ClientSettings{
		StrictHmac:           true,
		AutoReconnect:        false,
		NotificationsEnabled: true,
	}
	name := "Updated Name"
	req := models.UpdateSettingsRequest{
		Name:       &name,
		Thresholds: &thresholds,
		Client:     &client,
		Reset:      false,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["name"] != "Updated Name" {
		t.Errorf("name = %v, want Updated Name", result["name"])
	}
	// reset=false is omitted due to omitempty, so the key should not exist
	if _, ok := result["reset"]; ok {
		t.Error("reset=false should be omitted from JSON (omitempty)")
	}

	// Verify thresholds
	th, ok := result["thresholds"].(map[string]interface{})
	if !ok {
		t.Fatal("thresholds is not a map")
	}
	if th["riskWarn"].(float64) != 40 {
		t.Errorf("thresholds.riskWarn = %v, want 40", th["riskWarn"])
	}

	// Verify client settings
	cl, ok := result["client"].(map[string]interface{})
	if !ok {
		t.Fatal("client is not a map")
	}
	if cl["strictHmac"] != true {
		t.Errorf("client.strictHmac = %v, want true", cl["strictHmac"])
	}
	if cl["autoReconnect"] != false {
		t.Errorf("client.autoReconnect = %v, want false", cl["autoReconnect"])
	}
}

func TestAuthController_UpdateSettingsRequest_PartialUpdate(t *testing.T) {
	// Only update thresholds, no name or client
	thresholds := models.Thresholds{
		RiskWarn:   35,
		RiskCrit:   65,
		BufferWarn: 55,
		BufferCrit: 80,
	}
	req := models.UpdateSettingsRequest{
		Thresholds: &thresholds,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Name should be omitted (nil pointer)
	if _, ok := result["name"]; ok {
		t.Error("name should be omitted for partial update")
	}
	// Reset should be omitted (zero value false)
	if _, ok := result["reset"]; ok {
		t.Error("reset should be omitted when false (omitempty)")
	}

	// Verify thresholds only
	th, ok := result["thresholds"].(map[string]interface{})
	if !ok {
		t.Fatal("thresholds is not a map")
	}
	if th["riskWarn"].(float64) != 35 {
		t.Errorf("thresholds.riskWarn = %v, want 35", th["riskWarn"])
	}
}

func TestAuthController_UpdateSettingsRequest_ResetFlag(t *testing.T) {
	req := models.UpdateSettingsRequest{
		Reset: true,
	}

	data, err := json.Marshal(req)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["reset"] != true {
		t.Errorf("reset = %v, want true", result["reset"])
	}
}

func TestAuthController_ClientSettings_JSON(t *testing.T) {
	client := models.ClientSettings{
		StrictHmac:           true,
		AutoReconnect:        false,
		NotificationsEnabled: true,
	}

	data, err := json.Marshal(client)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if result["strictHmac"] != true {
		t.Errorf("strictHmac = %v, want true", result["strictHmac"])
	}
	if result["autoReconnect"] != false {
		t.Errorf("autoReconnect = %v, want false", result["autoReconnect"])
	}
	if result["notificationsEnabled"] != true {
		t.Errorf("notificationsEnabled = %v, want true", result["notificationsEnabled"])
	}
}

func TestAuthController_ClientSettings_PartialJSON(t *testing.T) {
	data := []byte(`{"strictHmac": true}`)

	var client models.ClientSettings
	if err := json.Unmarshal(data, &client); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if !client.StrictHmac {
		t.Error("StrictHmac = false, want true")
	}
	// Other fields should be zero values when not provided
	if client.AutoReconnect {
		t.Error("AutoReconnect should be false when omitted")
	}
}

func TestAuthController_OperatorResponse_IncludesClientSettings(t *testing.T) {
	thresholds := models.Thresholds{
		RiskWarn:    50,
		RiskCrit:    75,
		ThermalWarn: 45,
		ThermalCrit: 55,
		BufferWarn:  50,
		BufferCrit:  80,
	}
	client := models.ClientSettings{
		StrictHmac:           false,
		AutoReconnect:        true,
		NotificationsEnabled: true,
	}
	resp := models.OperatorResponse{
		ID:         "op-001",
		Email:      "test@example.com",
		Name:       "Test User",
		Role:       models.RoleOperator,
		Thresholds: &thresholds,
		Client:     &client,
	}

	data, err := json.Marshal(resp)
	if err != nil {
		t.Fatalf("json.Marshal() failed: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	// Check thresholds present
	th, ok := result["thresholds"].(map[string]interface{})
	if !ok {
		t.Fatal("thresholds not in response")
	}
	if th["riskWarn"].(float64) != 50 {
		t.Errorf("thresholds.riskWarn = %v, want 50", th["riskWarn"])
	}

	// Check client settings present
	cl, ok := result["client"].(map[string]interface{})
	if !ok {
		t.Fatal("client not in response")
	}
	if cl["autoReconnect"] != true {
		t.Errorf("client.autoReconnect = %v, want true", cl["autoReconnect"])
	}
}

func TestAuthController_GoogleOAuthCallbackRequest_JSON(t *testing.T) {
	data := []byte(`{"code":"auth-code-123","state":"state-abc"}`)

	var req models.GoogleOAuthCallbackRequest
	if err := json.Unmarshal(data, &req); err != nil {
		t.Fatalf("json.Unmarshal() failed: %v", err)
	}

	if req.Code != "auth-code-123" {
		t.Errorf("Code = %s, want auth-code-123", req.Code)
	}
	if req.State != "state-abc" {
		t.Errorf("State = %s, want state-abc", req.State)
	}
}

func TestAuthController_HTTPRequest_Parsing(t *testing.T) {
	body := []byte(`{"email":"test@example.com","password":"secret"}`)
	req := httptest.NewRequest(http.MethodPost, "/auth/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	if req.Method != http.MethodPost {
		t.Errorf("Method = %s, want POST", req.Method)
	}
	if req.Header.Get("Content-Type") != "application/json" {
		t.Error("Content-Type should be application/json")
	}
}
