package redaction

import "testing"

func TestRedactBasic(t *testing.T) {
	r := DefaultRedactor
	input := "api_key=sk_test1234567890"
	result := r.Redact(input)
	if result == input {
		t.Error("api_key should be redacted")
	}
}

func TestRedactEmpty(t *testing.T) {
	r := NewRedactor(DefaultConfig())
	if got := r.Redact(""); got != "" {
		t.Errorf("Redact(\"\") = %q, want \"\"", got)
	}
}

func TestRedactMapKey(t *testing.T) {
	r := DefaultRedactor
	input := map[string]string{
		"email":    "user@example.com",
		"password": "secret123",
	}
	result := r.RedactMapKey(input)
	if result["email"] == "[REDACTED]" {
		t.Error("email should not be redacted")
	}
	if result["password"] != "[REDACTED]" {
		t.Errorf("password = %q, want [REDACTED]", result["password"])
	}
}
