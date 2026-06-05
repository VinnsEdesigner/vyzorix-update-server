package controllers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/VinnsEdesigner/vyzorix-update-server/models"
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

func TestAuthController_UpdateNameRequest_JSON(t *testing.T) {
	req := models.UpdateNameRequest{
		Name: "Updated Name",
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