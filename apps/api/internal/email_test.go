package services

import (
	"context"
	"os"
	"testing"
)

func TestEmailService_IsConfigured(t *testing.T) {
	// Save original values
	origAPIKey := os.Getenv("RESEND_API_KEY")
	origEmail := os.Getenv("EMAIL_FROM")

	defer func() {
		os.Setenv("RESEND_API_KEY", origAPIKey)
		os.Setenv("EMAIL_FROM", origEmail)
	}()

	tests := []struct {
		name      string
		apiKey    string
		fromEmail string
		expected  bool
	}{
		{"configured", "test-key", "test@example.com", true},
		{"missing api key", "", "test@example.com", false},
		{"missing email", "test-key", "", false},
		{"both missing", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Setenv("RESEND_API_KEY", tt.apiKey)
			os.Setenv("EMAIL_FROM", tt.fromEmail)

			svc := NewEmailService()
			if svc.IsConfigured() != tt.expected {
				t.Errorf("IsConfigured() = %v, want %v", svc.IsConfigured(), tt.expected)
			}
		})
	}
}

func TestEmailService_SendVerificationEmail_NotConfigured(t *testing.T) {
	// Clear environment
	os.Unsetenv("RESEND_API_KEY")
	os.Unsetenv("EMAIL_FROM")

	svc := NewEmailService()
	err := svc.SendVerificationEmail(context.Background(), "test@example.com", "Test User", "token123")
	if err == nil {
		t.Error("expected error when not configured")
	}
}

func TestEmailService_SendPasswordResetEmail_NotConfigured(t *testing.T) {
	os.Unsetenv("RESEND_API_KEY")
	os.Unsetenv("EMAIL_FROM")

	svc := NewEmailService()
	err := svc.SendPasswordResetEmail(context.Background(), "test@example.com", "Test User", "token123")
	if err == nil {
		t.Error("expected error when not configured")
	}
}

func TestEmailService_SendPasswordChangedEmail_NotConfigured(t *testing.T) {
	os.Unsetenv("RESEND_API_KEY")
	os.Unsetenv("EMAIL_FROM")

	svc := NewEmailService()
	err := svc.SendPasswordChangedEmail(context.Background(), "test@example.com", "Test User")
	if err == nil {
		t.Error("expected error when not configured")
	}
}

func TestEmailService_ParseTemplate(t *testing.T) {
	svc := &EmailService{}

	data := EmailData{
		Name:        "John Doe",
		VerifyURL:   "https://example.com/verify?token=abc123",
		ExpiryHours: 24,
	}

	html, err := svc.parseTemplate(verificationEmailTemplate, data)
	if err != nil {
		t.Fatalf("parseTemplate failed: %v", err)
	}

	// Check that template was rendered with data
	if !contains(html, "John Doe") {
		t.Error("output should contain name")
	}
	if !contains(html, "https://example.com/verify?token=abc123") {
		t.Error("output should contain verify URL")
	}
	if !contains(html, "24 hours") {
		t.Error("output should contain expiry hours")
	}
}

func TestEmailService_ParseTemplate_PasswordReset(t *testing.T) {
	svc := &EmailService{}

	data := EmailData{
		Name:       "Jane Smith",
		ResetURL:   "https://example.com/reset?token=xyz789",
		ExpiryMins: 60,
	}

	html, err := svc.parseTemplate(passwordResetEmailTemplate, data)
	if err != nil {
		t.Fatalf("parseTemplate failed: %v", err)
	}

	if !contains(html, "Jane Smith") {
		t.Error("output should contain name")
	}
	if !contains(html, "https://example.com/reset?token=xyz789") {
		t.Error("output should contain reset URL")
	}
	if !contains(html, "60 minutes") {
		t.Error("output should contain expiry minutes")
	}
}

func TestEmailService_ParseTemplate_PasswordChanged(t *testing.T) {
	svc := &EmailService{}

	data := EmailData{
		Name: "Bob Wilson",
	}

	html, err := svc.parseTemplate(passwordChangedEmailTemplate, data)
	if err != nil {
		t.Fatalf("parseTemplate failed: %v", err)
	}

	if !contains(html, "Bob Wilson") {
		t.Error("output should contain name")
	}
	if !contains(html, "successfully") {
		t.Error("output should mention success")
	}
}

func TestEmailService_ParseTemplate_Invalid(t *testing.T) {
	svc := &EmailService{}

	_, err := svc.parseTemplate("{{.InvalidField}}", EmailData{Name: "Test"})
	if err == nil {
		t.Error("expected error for invalid template field")
	}
}

func TestEmailService_NewEmailService_Defaults(t *testing.T) {
	// Clear environment
	os.Unsetenv("RESEND_API_KEY")
	os.Unsetenv("EMAIL_FROM")
	os.Unsetenv("EMAIL_FROM_NAME")
	os.Unsetenv("BASE_URL")

	svc := NewEmailService()

	if svc.apiKey != "" {
		t.Errorf("apiKey should be empty, got %q", svc.apiKey)
	}
	if svc.fromEmail != "" {
		t.Errorf("fromEmail should be empty, got %q", svc.fromEmail)
	}
	if svc.fromName != "" {
		t.Errorf("fromName should be empty, got %q", svc.fromName)
	}
}

func TestEmailService_EmailData(t *testing.T) {
	data := EmailData{
		Name:        "Test User",
		VerifyURL:   "https://example.com/verify",
		ResetURL:    "https://example.com/reset",
		ExpiryHours: 24,
		ExpiryMins:  60,
	}

	if data.Name != "Test User" {
		t.Errorf("Name = %q, want %q", data.Name, "Test User")
	}
	if data.VerifyURL != "https://example.com/verify" {
		t.Errorf("VerifyURL = %q, want %q", data.VerifyURL, "https://example.com/verify")
	}
	if data.ExpiryHours != 24 {
		t.Errorf("ExpiryHours = %d, want %d", data.ExpiryHours, 24)
	}
}

// contains is a helper to check if a string contains a substring.
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
