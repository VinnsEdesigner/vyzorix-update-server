package outcome_test

import (
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/outcome"
)

func TestOwnedResourcesSuppressRetryability(t *testing.T) {
	f := outcome.Unavailable("dep_down", "unreachable")
	if !f.Retryable() {
		t.Fatal("a transient failure with nothing held should be retryable")
	}

	held := f.WithOwned("device-lock:dev-1")
	if held.Retryable() {
		t.Fatal("a failure that still holds a resource must not be retryable")
	}
}

func TestOwnershipInCauseChainSuppressesRetryability(t *testing.T) {
	inner := outcome.Timeout("inner", "inner").WithOwned("session:abc")
	outer := outcome.Unavailable("outer", "outer").WithCause(inner)
	if outer.Retryable() {
		t.Fatal("ownership anywhere in the cause chain must suppress retryability")
	}
}

func TestRemediateEscalatesWhileResourcesAreHeld(t *testing.T) {
	f := outcome.Unavailable("dep_down", "unreachable").WithOwned("device-lock:dev-1")
	r := outcome.Remediate(f)
	if r.Action != outcome.ActionEscalate {
		t.Fatalf("action = %v; ownership must outrank transient retry", r.Action)
	}
	if r.Guidance == "" {
		t.Fatal("escalation must carry guidance")
	}
}

func TestRemediateUncertainEffectAsksForReconciliation(t *testing.T) {
	f := outcome.Timeout("ack_timeout", "no ack")
	r := outcome.Remediate(f)
	if r.Action != outcome.ActionReconcile {
		t.Fatalf("action = %v; an unconfirmed outcome must reconcile, not retry", r.Action)
	}
}

func TestRemediateCleanTransientFailureRetries(t *testing.T) {
	f := outcome.New(outcome.ClassUnavailable, outcome.BlameEnvironment, outcome.EffectNone, "not_sent", "not sent")
	r := outcome.Remediate(f)
	if r.Action != outcome.ActionRetry {
		t.Fatalf("action = %v; a clean transient failure may retry", r.Action)
	}
}

func TestRemediateRateLimitWaits(t *testing.T) {
	f := outcome.RateLimited("slow_down", "too many")
	r := outcome.Remediate(f)
	if r.Action != outcome.ActionWait {
		t.Fatalf("action = %v; want wait", r.Action)
	}
	if r.RetryAfterSeconds <= 0 {
		t.Fatal("a wait remediation must state a delay")
	}
}

func TestRemediateDeniedEscalates(t *testing.T) {
	f := outcome.Denied("forbidden", "no")
	if r := outcome.Remediate(f); r.Action != outcome.ActionEscalate {
		t.Fatalf("action = %v; want escalate", r.Action)
	}
}

func TestCaptureFlagsTruncation(t *testing.T) {
	short := outcome.Capture("ok")
	if short.Truncated {
		t.Fatal("short capture must not be flagged truncated")
	}

	long := make([]byte, 20*1024)
	for i := range long {
		long[i] = 'x'
	}
	big := outcome.Capture(string(long))
	if !big.Truncated {
		t.Fatal("an oversized capture must be flagged truncated")
	}
	if big.Bytes != len(long) {
		t.Fatalf("Bytes = %d; want the original size %d", big.Bytes, len(long))
	}
	if len(big.Detail) > 8*1024 {
		t.Fatal("retained detail must stay bounded")
	}
}

func TestEvidenceSurvivesOnFailure(t *testing.T) {
	ev := outcome.Capture("connection reset by peer").WithExitCode(1).WithFact("host", "10.0.0.1")
	f := outcome.Unavailable("send_failed", "could not send").WithEvidence(ev)

	got := f.Evidence()
	if got.Detail == "" {
		t.Fatal("captured detail must be retained")
	}
	if got.ExitCode == nil || *got.ExitCode != 1 {
		t.Fatal("exit code must be retained")
	}
	if got.Facts["host"] != "10.0.0.1" {
		t.Fatal("structured facts must be retained")
	}
}

func TestEvidenceIsEmptyByDefault(t *testing.T) {
	if !(outcome.Evidence{}).IsEmpty() {
		t.Fatal("a zero Evidence must report empty")
	}
	if outcome.Capture("x").IsEmpty() {
		t.Fatal("captured evidence must not report empty")
	}
}
