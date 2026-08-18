
package errors

import (
	"errors"
	"testing"
)

func TestNewServerError(t *testing.T) {
	err := NewServerError(CodeAuthInvalidCredentials, "Invalid credentials")

	if err.Code != CodeAuthInvalidCredentials {
		t.Errorf("Code = %v, want AUTH_INVALID_CREDENTIALS", err.Code)
	}
	if err.Message != "Invalid credentials" {
		t.Errorf("Message = %v, want 'Invalid credentials'", err.Message)
	}
	if err.TraceID == "" {
		t.Error("TraceID should be generated")
	}
	if err.Timestamp.IsZero() {
		t.Error("Timestamp should be set")
	}
}

func TestNewServerErrorWithTrace(t *testing.T) {
	traceID := "test-trace-id-123"
	err := NewServerErrorWithTrace(CodeResourceNotFound, "Not found", traceID)

	if err.TraceID != traceID {
		t.Errorf("TraceID = %v, want %v", err.TraceID, traceID)
	}
}

func TestWrapError(t *testing.T) {
	originalErr := errors.New("database connection failed")
	err := WrapError(CodeInternalDatabaseError, "Database operation failed", originalErr)

	if err.Code != CodeInternalDatabaseError {
		t.Errorf("Code = %v, want INTERNAL_DATABASE_ERROR", err.Code)
	}
	if err.Internal == nil || !errors.Is(err, originalErr) {
		t.Error("Internal error should wrap original")
	}
}

func TestServerErrorWithDetails(t *testing.T) {
	err := NewServerError(CodeValidationFailed, "Validation failed")
	details := []ValidationDetail{
		{Field: "email", Message: "invalid format"},
	}

	result := err.WithDetails(details)

	if result != err {
		t.Error("WithDetails should return same error")
	}
	if len(err.Details.([]ValidationDetail)) != 1 {
		t.Error("Details should contain one item")
	}
}

func TestServerErrorWithInternal(t *testing.T) {
	originalErr := errors.New("token expired")
	err := NewServerError(CodeAuthTokenExpired, "Token expired")

	result := err.WithInternal(originalErr, "/api/v1/auth", "POST", "user-123")

	if result != err {
		t.Error("WithInternal should return same error")
	}
	if err.Internal == nil {
		t.Fatal("Internal should be set")
	}
	if !errors.Is(err, originalErr) {
		t.Error("Internal.Err should wrap original error")
	}
	if err.Internal.Path != "/api/v1/auth" {
		t.Errorf("Internal.Path = %v, want /api/v1/auth", err.Internal.Path)
	}
}

func TestServerErrorWithInternalContext(t *testing.T) {
	err := NewServerError(CodeValidationFailed, "Validation failed")

	_ = err.WithInternalContext("device_id", "device-abc")
	_ = err.WithInternalContext("field", "password")

	if err.Internal == nil || err.Internal.Context == nil {
		t.Fatal("Internal context should be set")
	}
	if err.Internal.Context["device_id"] != "device-abc" {
		t.Errorf("device_id = %v, want device-abc", err.Internal.Context["device_id"])
	}
}

func TestErrInvalidCredentials(t *testing.T) {
	err := ErrInvalidCredentials()

	if err.Code != CodeAuthInvalidCredentials {
		t.Errorf("Code = %v, want AUTH_INVALID_CREDENTIALS", err.Code)
	}
}

func TestErrAccountLocked(t *testing.T) {
	err := ErrAccountLocked(30)

	if err.Code != CodeAuthAccountLocked {
		t.Errorf("Code = %v, want AUTH_ACCOUNT_LOCKED", err.Code)
	}

	details := err.Details.(map[string]int)
	if details["retry_after_minutes"] != 30 {
		t.Errorf("retry_after_minutes = %v, want 30", details["retry_after_minutes"])
	}
}

func TestErrNotFound(t *testing.T) {
	err := ErrNotFound("Device", "device-123")

	if err.Code != CodeResourceNotFound {
		t.Errorf("Code = %v, want RESOURCE_NOT_FOUND", err.Code)
	}

	details := err.Details.(map[string]string)
	if details["resource_type"] != "Device" {
		t.Errorf("resource_type = %v, want Device", details["resource_type"])
	}
	if details["resource_id"] != "device-123" {
		t.Errorf("resource_id = %v, want device-123", details["resource_id"])
	}
}

func TestErrRateLimitExceeded(t *testing.T) {
	err := ErrRateLimitExceeded(60)

	if err.Code != CodeRateLimitExceeded {
		t.Errorf("Code = %v, want RATE_LIMIT_EXCEEDED", err.Code)
	}

	details := err.Details.(map[string]int)
	if details["retry_after_seconds"] != 60 {
		t.Errorf("retry_after_seconds = %v, want 60", details["retry_after_seconds"])
	}
}

func TestErrDeviceNotOnline(t *testing.T) {
	err := ErrDeviceNotOnline("device-xyz")

	if err.Code != CodeDeviceNotOnline {
		t.Errorf("Code = %v, want DEVICE_NOT_ONLINE", err.Code)
	}

	details := err.Details.(map[string]string)
	if details["device_id"] != "device-xyz" {
		t.Errorf("device_id = %v, want device-xyz", details["device_id"])
	}
}

func TestServerErrorErrorMethod(t *testing.T) {
	err := &ServerError{
		Code:    CodeResourceNotFound,
		Message: "Device not found",
		TraceID: "trace-123",
	}

	expected := "[trace-123] RESOURCE_NOT_FOUND: Device not found"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %v, want %v", got, expected)
	}
}

func TestServerErrorWithoutTraceID(t *testing.T) {
	err := &ServerError{
		Code:    CodeValidationFailed,
		Message: "Invalid input",
	}

	expected := "VALIDATION_FAILED: Invalid input"
	if got := err.Error(); got != expected {
		t.Errorf("Error() = %v, want %v", got, expected)
	}
}