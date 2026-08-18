
package errors

import (
	"testing"
)

func TestErrorCodeCategory(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected string
	}{
		{CodeAuthInvalidCredentials, "Authentication"},
		{CodeAuthSessionExpired, "Authentication"},
		{CodeAuthzInsufficientPermissions, "Authorization"},
		{CodeAuthzOrgMembershipRequired, "Authorization"},
		{CodeResourceNotFound, "Resource"},
		{CodeResourceAlreadyExists, "Resource"},
		{CodeValidationFailed, "Validation"},
		{CodeValidationRequiredField, "Validation"},
		{CodeRateLimitExceeded, "Rate Limiting"},
		{CodeSecurityThreatDetected, "Security"},
		{CodeDeviceNotOnline, "Device"},
		{CodeOrgNotFound, "Organization"},
		{CodeUpdateNotFound, "Update"},
		{CodeInternalServerError, "Internal"},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := tt.code.Category(); got != tt.expected {
				t.Errorf("Category() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorCodeHTTPStatusCode(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected int
	}{
		{CodeAuthInvalidCredentials, 401},
		{CodeAuthSessionExpired, 401},
		{CodeAuthCSRFInvalid, 401},
		{CodeAuthzInsufficientPermissions, 403},
		{CodeAuthzScopeForbidden, 403},
		{CodeResourceNotFound, 404},
		{CodeResourceDeleted, 404},
		{CodeResourceAlreadyExists, 409},
		{CodeResourceConflict, 409},
		{CodeResourceLimitExceeded, 429},
		{CodeValidationFailed, 400},
		{CodeValidationRequiredField, 400},
		{CodeRateLimitExceeded, 429},
		{CodeSecurityThreatDetected, 403},
		{CodeSecurityIPBlocked, 403},
		{CodeSecurityRiskUnconfirmed, 449},
		{CodeDeviceNotOnline, 400},
		{CodeDeviceCommandTimeout, 504},
		{CodeOrgNotFound, 404},
		{CodeOrgMemberLimit, 400},
		{CodeUpdateNotFound, 404},
		{CodeUpdateInProgress, 409},
		{CodeUpdateDeviceIncompatible, 409},
		{CodeInternalServerError, 500},
		{CodeInternalDatabaseError, 500},
		{CodeInternalTimeout, 504},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := tt.code.HTTPStatusCode(); got != tt.expected {
				t.Errorf("HTTPStatusCode() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestErrorCodeIsRetryable(t *testing.T) {
	tests := []struct {
		code     ErrorCode
		expected bool
	}{
		{CodeRateLimitExceeded, true},
		{CodeInternalTimeout, true},
		{CodeInternalServerError, true},
		{CodeInternalExternalServiceError, true},
		{CodeDeviceCommandTimeout, true},
		{CodeAuthInvalidCredentials, false},
		{CodeResourceNotFound, false},
		{CodeValidationFailed, false},
	}

	for _, tt := range tests {
		t.Run(string(tt.code), func(t *testing.T) {
			if got := tt.code.IsRetryable(); got != tt.expected {
				t.Errorf("IsRetryable() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestNewValidationDetail(t *testing.T) {
	detail := NewValidationDetail("email", "invalid format")

	if detail.Field != "email" {
		t.Errorf("Field = %v, want email", detail.Field)
	}
	if detail.Message != "invalid format" {
		t.Errorf("Message = %v, want 'invalid format'", detail.Message)
	}
}

func TestNewValidationDetailWithCode(t *testing.T) {
	detail := NewValidationDetailWithCode("password", "too weak", "VALIDATION_INVALID_PASSWORD")

	if detail.Field != "password" {
		t.Errorf("Field = %v, want password", detail.Field)
	}
	if detail.Message != "too weak" {
		t.Errorf("Message = %v, want 'too weak'", detail.Message)
	}
	if detail.Code != "VALIDATION_INVALID_PASSWORD" {
		t.Errorf("Code = %v, want VALIDATION_INVALID_PASSWORD", detail.Code)
	}
}

func TestWithTimestamp(t *testing.T) {
	ts := WithTimestamp()
	if ts.IsZero() {
		t.Error("WithTimestamp() returned zero time")
	}
}