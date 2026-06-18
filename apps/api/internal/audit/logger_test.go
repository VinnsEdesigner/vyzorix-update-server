// Package audit provides tests for audit logging functionality.
package audit

import (
	"testing"
)

func TestLoggerConfig_Default(t *testing.T) {
	cfg := DefaultLoggerConfig()
	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}
	if cfg.RetentionDays != 90 {
		t.Errorf("Expected RetentionDays to be 90, got %d", cfg.RetentionDays)
	}
}

func TestEntry_Fields(t *testing.T) {
	entry := &Entry{
		ID:           "entry-123",
		OperatorID:   "op-456",
		Action:       ActionLoginSuccess,
		ResourceType: "session",
		ResourceID:   "session-789",
		IPAddress:    "192.168.1.1",
		UserAgent:    "TestAgent/1.0",
		Result:       ResultSuccess,
		Metadata:     map[string]string{"key": "value"},
	}

	if entry.ID != "entry-123" {
		t.Errorf("Expected ID to be entry-123, got %s", entry.ID)
	}
	if entry.Action != ActionLoginSuccess {
		t.Errorf("Expected Action to be ActionLoginSuccess")
	}
	if entry.Result != ResultSuccess {
		t.Errorf("Expected Result to be ResultSuccess")
	}
}

func TestAction_Constants(t *testing.T) {
	// Verify action constants are defined.
	actions := []Action{
		ActionLoginSuccess,
		ActionLoginFailed,
		ActionLogout,
		ActionRegister,
		ActionPasswordChange,
		ActionPasswordReset,
		ActionEmailVerify,
		ActionMFAEnabled,
		ActionMFADisabled,
		ActionMFALogin,
		ActionSessionRevoked,
		ActionAccountLocked,
		ActionAccountUnlocked,
		ActionCSRFFailure,
		ActionSigningFailure,
		ActionRateLimitExceeded,
		ActionAPIClientCreated,
		ActionAPIClientRevoked,
		ActionSigningKeyRotated,
		ActionAdminAction,
	}

	for _, a := range actions {
		if a == "" {
			t.Error("Action constant should not be empty")
		}
	}
}

func TestResult_Constants(t *testing.T) {
	if ResultSuccess == "" {
		t.Error("ResultSuccess should not be empty")
	}
	if ResultFailure == "" {
		t.Error("ResultFailure should not be empty")
	}
	if ResultBlocked == "" {
		t.Error("ResultBlocked should not be empty")
	}
}
