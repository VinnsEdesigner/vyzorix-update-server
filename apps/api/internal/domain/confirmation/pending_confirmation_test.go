
package confirmation

import (
	"testing"
	"time"
)

func TestPendingConfirmation_IsExpired(t *testing.T) {
	past := &PendingConfirmation{ExpiresAt: time.Now().Add(-time.Minute)}
	if !past.IsExpired() {
		t.Error("expected past expiry to be expired")
	}
	future := &PendingConfirmation{ExpiresAt: time.Now().Add(time.Minute)}
	if future.IsExpired() {
		t.Error("expected future expiry to not be expired")
	}
}

func TestPendingConfirmation_IsConsumed(t *testing.T) {
	now := time.Now()
	if (&PendingConfirmation{}).IsConsumed() {
		t.Error("fresh confirmation should not be consumed")
	}
	if !(&PendingConfirmation{ConsumedAt: &now}).IsConsumed() {
		t.Error("confirmation with ConsumedAt should be consumed")
	}
}

func TestPendingConfirmation_Matches(t *testing.T) {
	c := &PendingConfirmation{OperatorID: "op-1", Command: "device.reboot", DeviceID: "imei-1"}
	tests := []struct {
		name                        string
		operatorID, command, device string
		want                        bool
	}{
		{"exact match", "op-1", "device.reboot", "imei-1", true},
		{"wrong operator", "op-2", "device.reboot", "imei-1", false},
		{"wrong command", "op-1", "device.reset", "imei-1", false},
		{"wrong device", "op-1", "device.reboot", "imei-2", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.Matches(tt.operatorID, tt.command, tt.device); got != tt.want {
				t.Errorf("Matches = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPendingConfirmation_MatchesAnyDeviceWhenUnscoped(t *testing.T) {
	// A confirmation issued without a specific device matches any device for
	// the same operator+command (e.g. org-wide confirmations).
	c := &PendingConfirmation{OperatorID: "op-1", Command: "device.reboot", DeviceID: ""}
	if !c.Matches("op-1", "device.reboot", "any-device") {
		t.Error("unscoped confirmation should match any device")
	}
}