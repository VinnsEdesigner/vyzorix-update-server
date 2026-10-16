package outcome_test

import (
	"errors"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/outcome"
)

func TestOkCarriesValueAndReportsSuccess(t *testing.T) {
	r := outcome.Ok(42)
	if !r.IsOk() || r.IsFailed() {
		t.Fatalf("Ok(42) must be successful")
	}
	if v, ok := r.Value(); !ok || v != 42 {
		t.Fatalf("Value() = %d, %v; want 42, true", v, ok)
	}
	if r.Failure() != nil {
		t.Fatalf("successful Result must have no failure")
	}
}

func TestFailedResultReturnsZeroValueNotPartial(t *testing.T) {
	r := outcome.Fail[int](outcome.NotFound("device_missing", "no such device"))
	v, ok := r.Value()
	if ok {
		t.Fatalf("failed Result must not report a value")
	}
	if v != 0 {
		t.Fatalf("failed Result must yield the zero value, got %d", v)
	}
}

func TestFailNormalisesNilFailure(t *testing.T) {
	r := outcome.Fail[int](nil)
	f := r.Failure()
	if f == nil {
		t.Fatalf("Fail(nil) must still carry a failure")
	}
	if f.Class() != outcome.ClassFault {
		t.Fatalf("Fail(nil) class = %v; want fault", f.Class())
	}
}

func TestFlatMapPropagatesFailureAndPreservesCause(t *testing.T) {
	first := outcome.Ok(2)
	chained := outcome.FlatMap(first, func(int) outcome.Result[int] {
		return outcome.Fail[int](outcome.Unavailable("dep_down", "dependency unreachable"))
	})
	if chained.IsOk() {
		t.Fatalf("failure must propagate through FlatMap")
	}
	if !chained.Failure().HasClass(outcome.ClassUnavailable) {
		t.Fatalf("cause class lost: %v", chained.Failure())
	}
}

func TestMapDoesNotTouchFailure(t *testing.T) {
	original := outcome.Fail[int](outcome.Denied("forbidden", "nope"))
	mapped := outcome.Map(original, func(int) int { return 99 })
	if mapped.IsOk() {
		t.Fatalf("Map must not convert failure to success")
	}
	if mapped.Failure().Code() != "forbidden" {
		t.Fatalf("Map changed the failure: %v", mapped.Failure())
	}
}

func TestCollectFailsFastWithoutPartialResult(t *testing.T) {
	items := []outcome.Result[int]{
		outcome.Ok(1),
		outcome.Fail[int](outcome.Invalid("bad", "bad")),
		outcome.Ok(3),
	}
	r := outcome.Collect(items)
	if r.IsOk() {
		t.Fatalf("Collect must fail when any element fails")
	}
	if _, ok := r.Value(); ok {
		t.Fatalf("a failed Collect must not expose a partial slice")
	}
}

func TestPartitionKeepsPartialProgress(t *testing.T) {
	items := []outcome.Result[int]{
		outcome.Ok(1),
		outcome.Fail[int](outcome.Invalid("bad", "bad")),
		outcome.Ok(3),
	}
	values, failures := outcome.Partition(items)
	if len(values) != 2 || len(failures) != 1 {
		t.Fatalf("Partition = %d values, %d failures; want 2, 1", len(values), len(failures))
	}
}

func TestRetryableOnlyForTransientClasses(t *testing.T) {
	cases := map[outcome.Class]bool{
		outcome.ClassUnavailable: true,
		outcome.ClassTimeout:     true,
		outcome.ClassRateLimited: true,
		outcome.ClassInvalid:     false,
		outcome.ClassNotFound:    false,
		outcome.ClassConflict:    false,
		outcome.ClassDenied:      false,
		outcome.ClassFault:       false,
	}
	for class, want := range cases {
		if got := class.Retryable(); got != want {
			t.Errorf("%v.Retryable() = %v; want %v", class, got, want)
		}
	}
}

func TestTimeoutAndUnavailableLeaveEffectUnknown(t *testing.T) {
	for _, f := range []*outcome.Failure{
		outcome.Timeout("timed_out", "deadline exceeded"),
		outcome.Unavailable("dep_down", "unreachable"),
	} {
		if f.Effect() != outcome.EffectUnknown {
			t.Errorf("%s effect = %v; want unknown", f.Code(), f.Effect())
		}
		if !f.Effect().NeedsReconciliation() {
			t.Errorf("%s must require reconciliation", f.Code())
		}
	}
}

func TestFailureHasClassMatchesThroughCauseChain(t *testing.T) {
	inner := outcome.NotFound("device_missing", "missing")
	outer := outcome.Unavailable("lookup_failed", "could not look up").WithCause(inner)
	if !outer.HasClass(outcome.ClassNotFound) {
		t.Fatalf("HasClass must walk the cause chain to find the class")
	}
	if !outer.HasClass(outcome.ClassUnavailable) {
		t.Fatalf("HasClass must also match the outer class")
	}
}

func TestFailureInteroperatesWithErrorsIs(t *testing.T) {
	inner := outcome.NotFound("device_missing", "missing")
	outer := outcome.Unavailable("lookup_failed", "wrapped").WithCause(inner)
	if !errors.Is(outer, inner) {
		t.Fatalf("errors.Is must traverse the failure chain")
	}
}

func TestInternalDetailIsNotInClientMessage(t *testing.T) {
	f := outcome.Fault(errors.New("nil map write at runner.go:42"))
	if f.Message() == "" {
		t.Fatalf("client message must be present")
	}
	if f.Message() == "nil map write at runner.go:42" {
		t.Fatalf("internal diagnostic must not leak into the client message")
	}
	if f.Internal() == "" {
		t.Fatalf("internal diagnostic must be retained for logs")
	}
}

func TestAttemptDispositionMatchesEffect(t *testing.T) {
	clean := outcome.CleanFailure[int](outcome.Invalid("bad", "bad"))
	if clean.Disposition() != outcome.DispositionFailedClean {
		t.Fatalf("clean failure disposition = %v", clean.Disposition())
	}
	if !clean.Disposition().SafeToRetry() {
		t.Fatalf("a clean failure must be safe to retry")
	}

	uncertain := outcome.UncertainFailure[int](outcome.Timeout("timed_out", "deadline"))
	if uncertain.Disposition() != outcome.DispositionUncertain {
		t.Fatalf("uncertain failure disposition = %v", uncertain.Disposition())
	}
	if uncertain.Disposition().SafeToRetry() {
		t.Fatalf("an uncertain failure must NOT be safe to retry")
	}
	if !uncertain.Disposition().NeedsReconciliation() {
		t.Fatalf("an uncertain failure must require reconciliation")
	}
}

func TestCleanFailureCannotClaimAppliedEffect(t *testing.T) {
	f := outcome.Unavailable("x", "x").WithEffect(outcome.EffectApplied)
	attempt := outcome.CleanFailure[int](f)
	if attempt.Failure().Effect() != outcome.EffectNone {
		t.Fatalf("CleanFailure must force effect=none, got %v", attempt.Failure().Effect())
	}
}

func TestClaimTryValueEnforcesCertaintyFloor(t *testing.T) {
	prov := outcome.Provenance{Source: "device_report", Method: "heartbeat", ObservedAtUnix: 1}
	reported := outcome.ReportedClaim("v1.2.0", prov)

	if _, ok := reported.TryValue(outcome.Reported); !ok {
		t.Fatalf("reported claim must satisfy a reported floor")
	}
	if _, ok := reported.TryValue(outcome.Observed); ok {
		t.Fatalf("reported claim must NOT satisfy an observed floor")
	}
}

func TestWeakenNeverStrengthens(t *testing.T) {
	prov := outcome.Provenance{Source: "x", Method: "y"}
	observed := outcome.Observable(1, prov)

	if got := outcome.Weaken(observed, outcome.Inferred).Certainty(); got != outcome.Inferred {
		t.Fatalf("Weaken to inferred = %v", got)
	}
	if got := outcome.Weaken(observed, outcome.Observed).Certainty(); got != outcome.Observed {
		t.Fatalf("Weaken to observed changed the grade: %v", got)
	}
}

func TestCombineTakesWeakestCertainty(t *testing.T) {
	prov := outcome.Provenance{Source: "a", Method: "m"}
	combined := outcome.Combine(
		outcome.Observable(1, prov),
		outcome.Inference(1, prov),
		outcome.ReportedClaim(1, prov),
	)
	if combined.Certainty() != outcome.Inferred {
		t.Fatalf("Combine must take the weakest premise, got %v", combined.Certainty())
	}
}

func TestCombineKeepsOldestObservation(t *testing.T) {
	combined := outcome.Combine(
		outcome.Observable(1, outcome.Provenance{ObservedAtUnix: 500}),
		outcome.Observable(1, outcome.Provenance{ObservedAtUnix: 100}),
	)
	if combined.Provenance().ObservedAtUnix != 100 {
		t.Fatalf("Combine must keep the stalest timestamp, got %d", combined.Provenance().ObservedAtUnix)
	}
}

func TestUnknownIsTheZeroCertainty(t *testing.T) {
	var zero outcome.Certainty
	if zero != outcome.Unknown {
		t.Fatalf("zero Certainty = %v; must be Unknown", zero)
	}
	if zero.AtLeast(outcome.Inferred) {
		t.Fatalf("Unknown must not satisfy any floor above it")
	}
}

func TestMapClaimPreservesGrade(t *testing.T) {
	prov := outcome.Provenance{Source: "x", Method: "y"}
	mapped := outcome.MapClaim(outcome.Inference(2, prov), func(i int) int { return i * 2 })
	if mapped.Certainty() != outcome.Inferred {
		t.Fatalf("MapClaim changed certainty to %v", mapped.Certainty())
	}
	if mapped.Raw() != 4 {
		t.Fatalf("MapClaim value = %d; want 4", mapped.Raw())
	}
}
