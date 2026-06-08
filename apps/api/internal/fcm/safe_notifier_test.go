package fcm

import (
	"context"
	"errors"
	"testing"
)

// mockNotifier implements Notifier for testing
type mockNotifier struct {
	err error
}

func (m *mockNotifier) SendSilentWake(ctx context.Context, wake SilentWake) error {
	return m.err
}

func TestSafeNotifier_NilNotifier(t *testing.T) {
	sn := &SafeNotifier{Notifier: nil}

	err := sn.SendSilentWake(context.Background(), SilentWake{
		DeviceID: "test-device",
	})

	if err != nil {
		t.Errorf("SafeNotifier with nil notifier should return nil, got %v", err)
	}
}

func TestSafeNotifier_DisabledError(t *testing.T) {
	sn := &SafeNotifier{Notifier: &mockNotifier{err: ErrDisabled}}

	err := sn.SendSilentWake(context.Background(), SilentWake{
		DeviceID: "test-device",
	})

	if err != nil {
		t.Errorf("SafeNotifier should swallow ErrDisabled, got %v", err)
	}
}

func TestSafeNotifier_OtherError(t *testing.T) {
	sn := &SafeNotifier{Notifier: &mockNotifier{err: errors.New("network error")}}

	err := sn.SendSilentWake(context.Background(), SilentWake{
		DeviceID: "test-device",
	})

	// SafeNotifier should swallow all errors for graceful degradation
	if err != nil {
		t.Errorf("SafeNotifier should swallow errors, got %v", err)
	}
}

func TestSafeNotifier_Success(t *testing.T) {
	sn := &SafeNotifier{Notifier: &mockNotifier{err: nil}}

	err := sn.SendSilentWake(context.Background(), SilentWake{
		DeviceID: "test-device",
	})

	if err != nil {
		t.Errorf("SafeNotifier should return nil on success, got %v", err)
	}
}

func TestErrUnavailable(t *testing.T) {
	err := ErrUnavailable
	if err.Error() != "fcm: temporarily unavailable" {
		t.Errorf("ErrUnavailable message = %s, want 'fcm: temporarily unavailable'", err.Error())
	}
}

func TestSilentWake_Fields_SafeNotifier(t *testing.T) {
	wake := SilentWake{
		Token:      "test-token",
		Command:    "restart",
		DispatchID: "dispatch-123",
		DeviceID:   "device-456",
	}

	if wake.Token != "test-token" {
		t.Error("Token field mismatch")
	}
	if wake.Command != "restart" {
		t.Error("Command field mismatch")
	}
	if wake.DispatchID != "dispatch-123" {
		t.Error("DispatchID field mismatch")
	}
	if wake.DeviceID != "device-456" {
		t.Error("DeviceID field mismatch")
	}
}
