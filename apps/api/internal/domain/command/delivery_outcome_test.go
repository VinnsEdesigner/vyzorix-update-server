package command_test

import (
	"errors"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/outcome"
)

var errUnreachable = errors.New("connection refused")

func TestAckTimeoutIsUncertain(t *testing.T) {
	f := command.ClassifyDelivery(command.TransportResult{Sent: true, Acked: false})
	if f == nil {
		t.Fatal("a sent-but-unacked frame must classify as a failure")
	}
	if f.Effect() != outcome.EffectUnknown {
		t.Fatalf("effect = %v; a delivered frame with no ack must be unknown", f.Effect())
	}
}

func TestDestructiveCommandIsNotRetriedOnUncertainEffect(t *testing.T) {
	f := command.ClassifyDelivery(command.TransportResult{Sent: true, Acked: false})
	plan := command.PlanDelivery(f, command.RiskTierCritical, 1, 5)

	if plan.Retry {
		t.Fatal("a critical-tier command with an uncertain effect must not be retried")
	}
	if !plan.Reconcile {
		t.Fatal("it must be flagged for reconciliation instead")
	}
}

func TestCleanFailureIsRetried(t *testing.T) {
	f := command.ClassifyDelivery(command.TransportResult{Sent: false, Acked: false})
	if f == nil {
		t.Fatal("expected a failure")
	}
	if f.Effect() != outcome.EffectNone {
		t.Fatalf("effect = %v; want none", f.Effect())
	}
	plan := command.PlanDelivery(f, command.RiskTierHigh, 1, 5)

	if !plan.Retry {
		t.Fatalf("a clean failure must be retried; got %+v", plan)
	}
	if plan.Reconcile {
		t.Fatal("a clean failure needs no reconciliation")
	}
}

func TestSendErrorIsUncertainEvenForDestructive(t *testing.T) {
	f := command.ClassifyDelivery(command.TransportResult{Sent: false, Err: errUnreachable})
	if f.Effect() == outcome.EffectNone {
		t.Fatal("a send error must not be treated as a proven no-op")
	}
	plan := command.PlanDelivery(f, command.RiskTierCritical, 1, 5)
	if plan.Retry {
		t.Fatal("a critical command must not be retried on an ambiguous send error")
	}
	if !plan.Reconcile {
		t.Fatal("it must be escalated for reconciliation")
	}
}

func TestUncertainLowTierStillRetries(t *testing.T) {
	f := command.ClassifyDelivery(command.TransportResult{Sent: true, Acked: false})
	plan := command.PlanDelivery(f, command.RiskTierLow, 1, 5)

	if !plan.Retry {
		t.Fatal("a low-tier command with an uncertain effect is idempotent and may retry")
	}
}

func TestOwnedResourceBlocksRetryEvenForLowTier(t *testing.T) {
	f := command.ClassifyDelivery(command.TransportResult{Sent: false, Err: errUnreachable}).
		WithOwned("device-lock:dev-1")

	if f.Retryable() {
		t.Fatal("a failure that still owns a resource must not report retryable")
	}
	plan := command.PlanDelivery(f, command.RiskTierLow, 1, 5)
	if plan.Retry {
		t.Fatal("a failure that still owns a resource must not be retried")
	}
}

func TestOwnedResourceEscalatesForDestructiveTier(t *testing.T) {
	f := command.ClassifyDelivery(command.TransportResult{Sent: false, Err: errUnreachable}).
		WithOwned("device-lock:dev-1")

	plan := command.PlanDelivery(f, command.RiskTierCritical, 1, 5)
	if plan.Retry {
		t.Fatal("a critical command holding a resource must not be retried")
	}
	if !plan.Reconcile {
		t.Fatal("it must be surfaced for reconciliation")
	}
}
