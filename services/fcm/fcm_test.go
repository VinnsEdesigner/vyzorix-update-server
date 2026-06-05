package fcm

import (
	"context"
	"log/slog"
	"testing"
)

func TestInit_Disabled(t *testing.T) {
	logger := slog.Default()
	
	// Empty credentials should return disabled client
	c, err := Init(logger, "")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	if c == nil {
		t.Fatal("Init() returned nil client")
	}
	if c.Enabled() {
		t.Error("client should be disabled with empty credentials")
	}
}

func TestInit_Disabled_EmptyCredentials(t *testing.T) {
	logger := slog.Default()
	
	// Explicitly empty string
	c, err := Init(logger, "")
	if err != nil {
		t.Fatalf("Init() returned error: %v", err)
	}
	
	if c.enabled {
		t.Error("client.enabled should be false")
	}
	if c.projectID != "" {
		t.Errorf("projectID should be empty, got %s", c.projectID)
	}
}

func TestClient_Enabled_NilClient(t *testing.T) {
	var c *Client
	
	if c.Enabled() {
		t.Error("nil client should not be enabled")
	}
}

func TestClient_Enabled_DisabledClient(t *testing.T) {
	c := &Client{
		log:     slog.Default(),
		enabled: false,
	}
	
	if c.Enabled() {
		t.Error("disabled client should not be enabled")
	}
}

func TestClient_Enabled_EnabledClient(t *testing.T) {
	c := &Client{
		log:     slog.Default(),
		enabled: true,
	}
	
	if !c.Enabled() {
		t.Error("enabled client should be enabled")
	}
}

func TestMessaging_NilApp(t *testing.T) {
	c := &Client{
		log:     slog.Default(),
		enabled: true,
		app:     nil,
	}
	
	client := c.Messaging()
	if client != nil {
		t.Error("Messaging() should return nil when app is nil")
	}
}

func TestMessaging_NilClient(t *testing.T) {
	var c *Client
	
	client := c.Messaging()
	if client != nil {
		t.Error("Messaging() should return nil for nil client")
	}
}

func TestGetProjectID_ValidCredentials(t *testing.T) {
	cred := `{
		"project_id": "test-project-123",
		"client_email": "test@example.com"
	}`
	
	projectID := getProjectID(cred)
	if projectID != "test-project-123" {
		t.Errorf("projectID = %s, want test-project-123", projectID)
	}
}

func TestGetProjectID_InvalidJSON(t *testing.T) {
	cred := `{invalid json}`
	
	projectID := getProjectID(cred)
	if projectID != "" {
		t.Errorf("projectID = %s, want empty", projectID)
	}
}

func TestGetProjectID_EmptyProjectID(t *testing.T) {
	cred := `{"project_id": ""}`
	
	projectID := getProjectID(cred)
	if projectID != "" {
		t.Errorf("projectID = %s, want empty", projectID)
	}
}

func TestGetProjectID_MissingProjectID(t *testing.T) {
	cred := `{"client_email": "test@example.com"}`
	
	projectID := getProjectID(cred)
	if projectID != "" {
		t.Errorf("projectID = %s, want empty", projectID)
	}
}

func TestSendSilentWake_Disabled(t *testing.T) {
	c := &Client{
		log:     slog.Default(),
		enabled: false,
	}
	
	wake := SilentWake{
		Token:      "token-123",
		Command:    "update",
		DispatchID: "dispatch-001",
		DeviceID:   "device-001",
	}
	
	err := c.SendSilentWake(context.Background(), wake)
	if err != ErrDisabled {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
}

func TestSendSilentWake_NilClient(t *testing.T) {
	var c *Client
	
	wake := SilentWake{
		Token:      "token-123",
		Command:    "update",
		DispatchID: "dispatch-001",
		DeviceID:   "device-001",
	}
	
	err := c.SendSilentWake(context.Background(), wake)
	if err != ErrDisabled {
		t.Errorf("err = %v, want ErrDisabled", err)
	}
}

func TestSendSilentWake_MissingToken(t *testing.T) {
	c := &Client{
		log:     slog.Default(),
		enabled: true,
		app:     nil, // app is nil but client is "enabled" - token check should fail first
	}
	
	wake := SilentWake{
		Token:      "", // empty token
		Command:    "update",
		DispatchID: "dispatch-001",
		DeviceID:   "device-001",
	}
	
	err := c.SendSilentWake(context.Background(), wake)
	if err == nil {
		t.Error("expected error for missing token")
	}
}

func TestSilentWake_Fields(t *testing.T) {
	wake := SilentWake{
		Token:      "token-abc",
		Command:    "reboot",
		DispatchID: "dispatch-xyz",
		DeviceID:   "device-123",
	}
	
	if wake.Token != "token-abc" {
		t.Errorf("Token = %s, want token-abc", wake.Token)
	}
	if wake.Command != "reboot" {
		t.Errorf("Command = %s, want reboot", wake.Command)
	}
	if wake.DispatchID != "dispatch-xyz" {
		t.Errorf("DispatchID = %s, want dispatch-xyz", wake.DispatchID)
	}
	if wake.DeviceID != "device-123" {
		t.Errorf("DeviceID = %s, want device-123", wake.DeviceID)
	}
}

func TestPtr24Hours(t *testing.T) {
	d := ptr24Hours()
	if d == nil {
		t.Fatal("ptr24Hours() returned nil")
	}
	if d.Hours() != 24 {
		t.Errorf("duration = %v hours, want 24 hours", d.Hours())
	}
}

func TestNotifierInterface(t *testing.T) {
	// Verify SilentWake implements Notifier interface
	var _ Notifier = (*Client)(nil)
}