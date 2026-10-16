package confirmation

import (
	"context"
	"errors"
	"testing"
	"time"

	domainconfirmation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/outcome"
)

type failingConsumeRepo struct {
	*fakeRepo
	consumeErr error
}

func (r *failingConsumeRepo) Consume(_ context.Context, _ string, _ time.Time) (*domainconfirmation.PendingConfirmation, error) {
	return nil, r.consumeErr
}

func seedToken(t *testing.T, repo *fakeRepo) *domainconfirmation.PendingConfirmation {
	t.Helper()
	svc := NewService(repo)
	c, err := svc.RequestConfirmation(context.Background(), "op-1", "org-1", "device.reboot", "dev-1")
	if err != nil {
		t.Fatalf("seeding token: %v", err)
	}
	return c
}

func TestTypedConsumeUnknownTokenIsNotFound(t *testing.T) {
	svc := NewService(newFakeRepo())
	result := svc.ConsumeForCommandTyped(context.Background(), "nope", "op-1", "device.reboot", "dev-1")
	if result.Failure == nil {
		t.Fatal("expected a failure")
	}
	if !result.Failure.HasClass(outcome.ClassNotFound) {
		t.Fatalf("class = %v; want not_found", result.Failure.Class())
	}
}

func TestTypedConsumeSuccessHasNoFailure(t *testing.T) {
	repo := newFakeRepo()
	c := seedToken(t, repo)
	svc := NewService(repo)

	result := svc.ConsumeForCommandTyped(context.Background(), c.Token, "op-1", "device.reboot", "dev-1")
	if result.Failure != nil {
		t.Fatalf("unexpected failure: %v", result.Failure)
	}
	if result.Confirmation == nil || result.Confirmation.ConsumedAt == nil {
		t.Fatal("token must be marked consumed")
	}
}

func TestTypedConsumeAlreadyConsumedIsCleanRejection(t *testing.T) {
	repo := newFakeRepo()
	c := seedToken(t, repo)
	svc := NewService(repo)

	if r := svc.ConsumeForCommandTyped(context.Background(), c.Token, "op-1", "device.reboot", "dev-1"); r.Failure != nil {
		t.Fatalf("first consume failed: %v", r.Failure)
	}

	second := svc.ConsumeForCommandTyped(context.Background(), c.Token, "op-1", "device.reboot", "dev-1")
	if second.Failure == nil {
		t.Fatal("expected a failure on reuse")
	}
	if !second.Failure.HasClass(outcome.ClassConflict) {
		t.Fatalf("class = %v; want conflict", second.Failure.Class())
	}
	if second.Failure.Effect() != outcome.EffectNone {
		t.Fatalf("effect = %v; a rejected reuse changes nothing", second.Failure.Effect())
	}
	if second.Failure.Effect().NeedsReconciliation() {
		t.Fatal("a rejected reuse must not require reconciliation")
	}
}

func TestTypedConsumeStoreFailureIsUncertain(t *testing.T) {
	repo := &failingConsumeRepo{fakeRepo: newFakeRepo(), consumeErr: errors.New("disk error")}
	c := seedToken(t, repo.fakeRepo)
	svc := NewService(repo)

	result := svc.ConsumeForCommandTyped(context.Background(), c.Token, "op-1", "device.reboot", "dev-1")
	if result.Failure == nil {
		t.Fatal("expected a failure")
	}
	if result.Failure.Effect() != outcome.EffectUnknown {
		t.Fatalf("effect = %v; want unknown", result.Failure.Effect())
	}
	if !outcome.DispositionFor(result.Failure).NeedsReconciliation() {
		t.Fatal("an uncertain consume must require reconciliation")
	}
}

func TestTypedConsumeMismatchIsDenied(t *testing.T) {
	repo := newFakeRepo()
	c := seedToken(t, repo)
	svc := NewService(repo)

	result := svc.ConsumeForCommandTyped(context.Background(), c.Token, "op-2", "device.reboot", "dev-1")
	if result.Failure == nil || !result.Failure.HasClass(outcome.ClassDenied) {
		t.Fatalf("want denied, got %v", result.Failure)
	}
}
