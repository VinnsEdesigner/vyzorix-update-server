package controllers

import (
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/config"
	"github.com/VinnsEdesigner/vyzorix/apps/api/security"
)

func TestUpgraderFactory_Create(t *testing.T) {
	// Test that factory creates upgraders with consistent settings
	validator := security.NewOriginValidator([]string{"https://app.example.com"})
	factory := NewUpgraderFactory(validator)

	upgrader := factory.Create()

	// Verify timeout is set
	if upgrader.HandshakeTimeout != 10*time.Second {
		t.Errorf("HandshakeTimeout = %v, want 10s", upgrader.HandshakeTimeout)
	}

	// Verify CheckOrigin is set
	if upgrader.CheckOrigin == nil {
		t.Error("CheckOrigin should be set")
	}
}

func TestUpgraderFactory_SetHandshakeTimeout(t *testing.T) {
	validator := security.NewOriginValidator([]string{"https://app.example.com"})
	factory := NewUpgraderFactory(validator).SetHandshakeTimeout(5 * time.Second)

	upgrader := factory.Create()

	if upgrader.HandshakeTimeout != 5*time.Second {
		t.Errorf("HandshakeTimeout = %v, want 5s", upgrader.HandshakeTimeout)
	}
}

func TestUpgraderFactory_Consistency(t *testing.T) {
	// Multiple calls should return upgraders with same config
	validator := security.NewOriginValidator([]string{"https://app.example.com"})
	factory := NewUpgraderFactory(validator).SetHandshakeTimeout(15 * time.Second)

	upgrader1 := factory.Create()
	upgrader2 := factory.Create()

	// Both should have same CheckOrigin function reference
	if upgrader1.CheckOrigin == nil || upgrader2.CheckOrigin == nil {
		t.Error("CheckOrigin should not be nil")
	}

	if upgrader1.HandshakeTimeout != upgrader2.HandshakeTimeout {
		t.Error("HandshakeTimeout should be consistent")
	}
}

func TestNewWebSocketHandlerWithFactory(t *testing.T) {
	validator := security.NewOriginValidator([]string{"https://app.example.com"})
	factory := NewUpgraderFactory(validator)

	// This should not panic
	handler := NewWebSocketHandlerWithFactory(nil, config.Config{}, nil, security.Verifier{}, factory)
	//nolint:staticcheck // SA5011: handler from factory is never nil here
	if handler == nil {
		t.Error("Handler should not be nil")
	}
	//nolint:staticcheck // SA5011: handler is already checked non-nil above
	if handler.originValidator != validator {
		t.Error("originValidator should be set from factory")
	}
}