<<<<<<< HEAD

=======
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
package confirmation

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	domainconfirmation "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
)

// fakeRepo is an in-memory confirmation.Repository for service tests.
type fakeRepo struct {
	mu            sync.Mutex
	confirmations map[string]*domainconfirmation.PendingConfirmation
	consumeCalls  int
}

func newFakeRepo() *fakeRepo {
	return &fakeRepo{confirmations: make(map[string]*domainconfirmation.PendingConfirmation)}
}

func (r *fakeRepo) Create(_ context.Context, c *domainconfirmation.PendingConfirmation) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.confirmations[c.Token] = c
	return nil
}

func (r *fakeRepo) Get(_ context.Context, token string) (*domainconfirmation.PendingConfirmation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	c, ok := r.confirmations[token]
	if !ok {
		return nil, domainconfirmation.ErrNotFound
	}
	return c, nil
}

func (r *fakeRepo) Consume(ctx context.Context, token string, at time.Time) (*domainconfirmation.PendingConfirmation, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.consumeCalls++
	c, ok := r.confirmations[token]
	if !ok {
		return nil, domainconfirmation.ErrNotFound
	}
	if c.ConsumedAt != nil {
		return c, domainconfirmation.ErrAlreadyConsumed
	}
	if c.ExpiresAt.Before(at) || c.ExpiresAt.Equal(at) {
		return c, domainconfirmation.ErrExpired
	}
	now := at
	c.ConsumedAt = &now
	return c, nil
}

func (r *fakeRepo) DeleteExpired(_ context.Context) (int64, error) { return 0, nil }

func TestService_RequestConfirmation_IssuesTokenWithTTL(t *testing.T) {
	svc := NewService(newFakeRepo())
	c, err := svc.RequestConfirmation(context.Background(), "op-1", "org-1", "device.reboot", "imei-1")
	if err != nil {
		t.Fatalf("RequestConfirmation: %v", err)
	}
	if c.Token == "" {
		t.Error("token must be set")
	}
	if c.OperatorID != "op-1" || c.Command != "device.reboot" || c.DeviceID != "imei-1" {
		t.Errorf("unexpected confirmation: %+v", c)
	}
	if c.RiskTier != "high" {
		t.Errorf("risk_tier = %s, want high", c.RiskTier)
	}
	if !c.ExpiresAt.After(c.CreatedAt) {
		t.Error("expires_at must be after created_at")
	}
}

func TestService_RequestConfirmation_RejectsEmptyInputs(t *testing.T) {
	svc := NewService(newFakeRepo())
	if _, err := svc.RequestConfirmation(context.Background(), "", "org", "cmd", "dev"); err == nil {
		t.Error("expected error for empty operatorID")
	}
	if _, err := svc.RequestConfirmation(context.Background(), "op", "org", "", "dev"); err == nil {
		t.Error("expected error for empty command")
	}
}

func TestService_ConsumeForCommand_Success(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	c, _ := svc.RequestConfirmation(context.Background(), "op-1", "org-1", "device.reboot", "imei-1")

	consumed, err := svc.ConsumeForCommand(context.Background(), c.Token, "op-1", "device.reboot", "imei-1")
	if err != nil {
		t.Fatalf("ConsumeForCommand: %v", err)
	}
	if !consumed.IsConsumed() {
		t.Error("confirmation should be marked consumed")
	}
	if repo.consumeCalls != 1 {
		t.Errorf("expected 1 consume call, got %d", repo.consumeCalls)
	}
}

func TestService_ConsumeForCommand_RejectsMismatch(t *testing.T) {
	svc := NewService(newFakeRepo())
	c, _ := svc.RequestConfirmation(context.Background(), "op-1", "org-1", "device.reboot", "imei-1")

	_, err := svc.ConsumeForCommand(context.Background(), c.Token, "op-2", "device.reboot", "imei-1")
	if !errors.Is(err, domainconfirmation.ErrMismatch) {
		t.Errorf("expected ErrMismatch, got %v", err)
	}
}

func TestService_ConsumeForCommand_RejectsUnknownToken(t *testing.T) {
	svc := NewService(newFakeRepo())
	_, err := svc.ConsumeForCommand(context.Background(), "no-such-token", "op-1", "device.reboot", "imei-1")
	if !errors.Is(err, domainconfirmation.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestService_ConsumeForCommand_RejectsExpired(t *testing.T) {
	repo := newFakeRepo()
	svc := NewService(repo)
	// Manually insert an already-expired confirmation.
	c := &domainconfirmation.PendingConfirmation{
		Token: "expired", OperatorID: "op-1", Command: "device.reboot", DeviceID: "imei-1",
		CreatedAt: time.Now().Add(-time.Hour), ExpiresAt: time.Now().Add(-time.Minute),
	}
	_ = repo.Create(context.Background(), c)

	_, err := svc.ConsumeForCommand(context.Background(), "expired", "op-1", "device.reboot", "imei-1")
	if !errors.Is(err, domainconfirmation.ErrExpired) {
		t.Errorf("expected ErrExpired, got %v", err)
	}
}

func TestService_ConsumeForCommand_SingleUse(t *testing.T) {
	svc := NewService(newFakeRepo())
	c, _ := svc.RequestConfirmation(context.Background(), "op-1", "org-1", "device.reboot", "imei-1")

	if _, err := svc.ConsumeForCommand(context.Background(), c.Token, "op-1", "device.reboot", "imei-1"); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	_, err := svc.ConsumeForCommand(context.Background(), c.Token, "op-1", "device.reboot", "imei-1")
	if !errors.Is(err, domainconfirmation.ErrAlreadyConsumed) {
		t.Errorf("second consume expected ErrAlreadyConsumed, got %v", err)
	}
<<<<<<< HEAD
}
=======
}
>>>>>>> 34b853d (feat: production hardening — structured errors, risk/audit, confirmation flow, validation, security hardening)
