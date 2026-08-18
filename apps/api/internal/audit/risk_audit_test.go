<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
package audit

import (
	"context"
	"testing"
)

func TestEntry_RiskFields(t *testing.T) {
	entry := &Entry{
		Action:   ActionCommandExecuted,
		Result:   ResultSuccess,
		TraceID:  "trace-abc",
		RiskTier: "high",
	}
	if entry.TraceID != "trace-abc" {
		t.Errorf("TraceID = %q, want trace-abc", entry.TraceID)
	}
	if entry.RiskTier != "high" {
		t.Errorf("RiskTier = %q, want high", entry.RiskTier)
	}
}

func TestEntry_ChangeTrackingFields(t *testing.T) {
	entry := &Entry{
		Action:     ActionCommandExecuted,
		Result:     ResultSuccess,
		ActorType:  "operator",
		ActorEmail: "admin@vyzorix.test",
		OldValue:   "pending",
		NewValue:   "delivered",
	}
	if entry.ActorType != "operator" {
		t.Errorf("ActorType = %q, want operator", entry.ActorType)
	}
	if entry.ActorEmail != "admin@vyzorix.test" {
		t.Errorf("ActorEmail = %q, want admin@vyzorix.test", entry.ActorEmail)
	}
	if entry.OldValue != "pending" || entry.NewValue != "delivered" {
		t.Errorf("change tracking = %q→%q, want pending→delivered", entry.OldValue, entry.NewValue)
	}
}

func TestAction_CommandExecutedConstant(t *testing.T) {
	if ActionCommandExecuted != "command_executed" {
		t.Errorf("ActionCommandExecuted = %q, want command_executed", ActionCommandExecuted)
	}
}

func TestNoOpLogger_CommandExecuted(t *testing.T) {
	l := NewNoOpLogger()
	if l == nil {
		t.Fatal("NewNoOpLogger returned nil")
	}
	l.CommandExecuted(context.Background(), CommandExecutedEvent{
		OperatorID: "op-1",
		DeviceID:   "dev-1",
		Command:    "device.reboot",
		Result:     ResultBlocked,
		RiskTier:   "high",
		TraceID:    "trace-1",
	})
}

func TestNoOpLogger_CommandExecutedWithChangeTracking(t *testing.T) {
	l := NewNoOpLogger()
	l.CommandExecuted(context.Background(), CommandExecutedEvent{
		OperatorID: "op-1",
		DeviceID:   "dev-1",
		Command:    "device.reboot",
		Result:     ResultSuccess,
		ActorType:  "operator",
		ActorEmail: "admin@vyzorix.test",
		OldValue:   "pending",
		NewValue:   "delivered",
	})
}

func TestNoOpLogger_ImplementsInterface(t *testing.T) {
	var _ interface {
		CommandExecuted(ctx context.Context, e CommandExecutedEvent)
	} = (*NoOpLogger)(nil)
	var _ interface {
		CommandExecuted(ctx context.Context, e CommandExecutedEvent)
	} = (*Logger)(nil)
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
