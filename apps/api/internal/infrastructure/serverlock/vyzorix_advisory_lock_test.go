package serverlock

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE server_locks (name TEXT PRIMARY KEY, holder TEXT NOT NULL, acquired_at INTEGER NOT NULL, expires_at INTEGER NOT NULL)`)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	return db
}

func TestAcquire_FirstTime(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	got, err := s.Acquire(ctx, "test-lock", "worker-1", 30*time.Second)
	if err != nil {
		t.Fatalf("acquire: %v", err)
	}
	if !got {
		t.Error("should acquire lock on first attempt")
	}
}

func TestAcquire_SecondHolderFails(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 30*time.Second)
	got, _ := s.Acquire(ctx, "test-lock", "worker-2", 30*time.Second)
	if got {
		t.Error("second holder should not acquire while first holds")
	}
}

func TestAcquire_ExpiredLockReacquired(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 1*time.Nanosecond)
	time.Sleep(10 * time.Millisecond)

	got, _ := s.Acquire(ctx, "test-lock", "worker-2", 30*time.Second)
	if !got {
		t.Error("expired lock should be reacquired by new holder")
	}
}

func TestRelease(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 30*time.Second)
	_ = s.Release(ctx, "test-lock", "worker-1")

	held, _ := s.IsHeld(ctx, "test-lock")
	if held {
		t.Error("lock should not be held after release")
	}
}

func TestRelease_WrongHolder(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 30*time.Second)
	_ = s.Release(ctx, "test-lock", "wrong-holder")

	held, _ := s.IsHeld(ctx, "test-lock")
	if !held {
		t.Error("lock should still be held when wrong holder releases")
	}
}

func TestIsHeld_NoLock(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	held, _ := s.IsHeld(ctx, "nonexistent-lock")
	if held {
		t.Error("nonexistent lock should not be held")
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
