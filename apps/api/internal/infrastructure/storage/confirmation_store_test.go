
package storage

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/confirmation"
)

func newConfirmationTestRepo(t *testing.T) *ConfirmationRepository {
	t.Helper()
	s, err := Open(DefaultConfig(t.TempDir() + "/confirm_test.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return NewConfirmationRepository(s.DB())
}

func TestConfirmationRepository_CreateAndGet(t *testing.T) {
	repo := newConfirmationTestRepo(t)
	ctx := context.Background()
	c := &confirmation.PendingConfirmation{
		Token: "tok-1", OperatorID: "op-1", OrgID: "org-1",
		Command: "device.reboot", DeviceID: "imei-1", RiskTier: "high",
		CreatedAt: time.Now(), ExpiresAt: time.Now().Add(5 * time.Minute),
	}
	if err := repo.Create(ctx, c); err != nil {
		t.Fatalf("Create: %v", err)
	}
	got, err := repo.Get(ctx, "tok-1")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.OperatorID != "op-1" || got.Command != "device.reboot" || got.DeviceID != "imei-1" {
		t.Errorf("unexpected confirmation: %+v", got)
	}
	if got.IsConsumed() {
		t.Error("fresh confirmation should not be consumed")
	}
}

func TestConfirmationRepository_GetNotFound(t *testing.T) {
	repo := newConfirmationTestRepo(t)
	if _, err := repo.Get(context.Background(), "missing"); !errors.Is(err, confirmation.ErrNotFound) {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestConfirmationRepository_Consume_Success(t *testing.T) {
	repo := newConfirmationTestRepo(t)
	ctx := context.Background()
	_ = repo.Create(ctx, &confirmation.PendingConfirmation{
		Token: "tok-2", OperatorID: "op-1", Command: "device.reboot", DeviceID: "imei-1",
		CreatedAt: time.Now(), ExpiresAt: time.Now().Add(5 * time.Minute),
	})

	consumed, err := repo.Consume(ctx, "tok-2", time.Now())
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if !consumed.IsConsumed() {
		t.Error("confirmation should be consumed after Consume")
	}
}

func TestConfirmationRepository_Consume_AlreadyConsumed(t *testing.T) {
	repo := newConfirmationTestRepo(t)
	ctx := context.Background()
	_ = repo.Create(ctx, &confirmation.PendingConfirmation{
		Token: "tok-3", OperatorID: "op-1", Command: "device.reboot", DeviceID: "imei-1",
		CreatedAt: time.Now(), ExpiresAt: time.Now().Add(5 * time.Minute),
	})
	if _, err := repo.Consume(ctx, "tok-3", time.Now()); err != nil {
		t.Fatalf("first consume: %v", err)
	}
	_, err := repo.Consume(ctx, "tok-3", time.Now())
	if !errors.Is(err, confirmation.ErrAlreadyConsumed) {
		t.Errorf("expected ErrAlreadyConsumed, got %v", err)
	}
}

func TestConfirmationRepository_Consume_Expired(t *testing.T) {
	repo := newConfirmationTestRepo(t)
	ctx := context.Background()
	_ = repo.Create(ctx, &confirmation.PendingConfirmation{
		Token: "tok-4", OperatorID: "op-1", Command: "device.reboot", DeviceID: "imei-1",
		CreatedAt: time.Now().Add(-time.Hour), ExpiresAt: time.Now().Add(-time.Minute),
	})
	_, err := repo.Consume(ctx, "tok-4", time.Now())
	if !errors.Is(err, confirmation.ErrExpired) {
		t.Errorf("expected ErrExpired, got %v", err)
	}
}

func TestConfirmationRepository_DeleteExpired(t *testing.T) {
	repo := newConfirmationTestRepo(t)
	ctx := context.Background()
	_ = repo.Create(ctx, &confirmation.PendingConfirmation{
		Token: "tok-exp", OperatorID: "op-1", Command: "device.reboot", DeviceID: "imei-1",
		CreatedAt: time.Now().Add(-time.Hour), ExpiresAt: time.Now().Add(-time.Minute),
	})
	n, err := repo.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if n < 1 {
		t.Errorf("expected at least 1 row deleted, got %d", n)
	}
	if _, err := repo.Get(ctx, "tok-exp"); !errors.Is(err, confirmation.ErrNotFound) {
		t.Errorf("expected expired confirmation to be gone, got %v", err)
	}
}