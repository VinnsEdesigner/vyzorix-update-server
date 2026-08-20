package serverlock

import (
	"context"
	"database/sql"
	"os"
	"strconv"
	"sync"
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

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	got, _ := s.Acquire(ctx, "test-lock", "worker-2", 30*time.Second)
	if !got {
		t.Error("expired lock should be reacquired by new holder")
	}
}

func TestAcquire_SameHolderReacquiresAfterRelease(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 30*time.Second)
	_ = s.Release(ctx, "test-lock", "worker-1")

	got, err := s.Acquire(ctx, "test-lock", "worker-1", 30*time.Second)
	if err != nil {
		t.Fatalf("re-acquire after release: %v", err)
	}
	if !got {
		t.Error("same holder should reacquire after releasing")
	}
}

func TestAcquire_SameHolderRefreshesWhileHolding(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 1*time.Second)
	got, err := s.Acquire(ctx, "test-lock", "worker-1", 60*time.Second)
	if err != nil {
		t.Fatalf("refresh: %v", err)
	}
	if !got {
		t.Error("same holder should be able to refresh while holding")
	}

	held, _ := s.IsHeld(ctx, "test-lock")
	if !held {
		t.Error("lock should still be held after refresh")
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

func TestRelease_NonexistentLock_NoError(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	err := s.Release(ctx, "nonexistent", "worker-1")
	if err != nil {
		t.Errorf("releasing nonexistent lock should not error, got: %v", err)
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

func TestIsHeld_ActiveLock_ReturnsTrue(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 30*time.Second)
	held, err := s.IsHeld(ctx, "test-lock")
	if err != nil {
		t.Fatalf("IsHeld: %v", err)
	}
	if !held {
		t.Error("active lock should report as held")
	}
}

func TestIsHeld_ExpiredLock_ReturnsFalse(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	_, _ = s.Acquire(ctx, "test-lock", "worker-1", 1*time.Millisecond)
	time.Sleep(10 * time.Millisecond)

	held, _ := s.IsHeld(ctx, "test-lock")
	if held {
		t.Error("expired lock should report as not held")
	}
}

func TestMultipleLocks_Independent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	got1, _ := s.Acquire(ctx, "lock-a", "worker-1", 30*time.Second)
	got2, _ := s.Acquire(ctx, "lock-b", "worker-2", 30*time.Second)

	if !got1 || !got2 {
		t.Error("different named locks should be independently acquirable")
	}

	heldA, _ := s.IsHeld(ctx, "lock-a")
	heldB, _ := s.IsHeld(ctx, "lock-b")
	if !heldA || !heldB {
		t.Error("both locks should be held")
	}

	_ = s.Release(ctx, "lock-a", "worker-1")
	heldA2, _ := s.IsHeld(ctx, "lock-a")
	heldB2, _ := s.IsHeld(ctx, "lock-b")
	if heldA2 {
		t.Error("lock-a should be released")
	}
	if !heldB2 {
		t.Error("lock-b should still be held after lock-a released")
	}
}

func TestConcurrentAcquire_OnlyOneWins(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	var wg sync.WaitGroup
	winners := make(chan bool, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			holder := "worker-" + strconv.Itoa(id)
			got, _ := s.Acquire(ctx, "race-lock", holder, 5*time.Second)
			winners <- got
		}(i)
	}
	wg.Wait()
	close(winners)

	winCount := 0
	for got := range winners {
		if got {
			winCount++
		}
	}

	if winCount != 1 {
		t.Errorf("expected exactly 1 winner in concurrent race, got %d", winCount)
	}
}

func TestWorkerSimulation_LockPreventsDoubleExecution(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	execCount := 0
	var mu sync.Mutex

	// simulateTick acquires the lock, does work, then releases.
	// If a second tick runs while the first holds the lock, it should be blocked.
	simulateTick := func(holder string) bool {
		acquired, _ := s.Acquire(ctx, "worker-lock", holder, 5*time.Second)
		if !acquired {
			return false
		}
		mu.Lock()
		execCount++
		mu.Unlock()
		_ = s.Release(ctx, "worker-lock", holder)
		return true
	}

	// First tick acquires and releases.
	simulateTick("instance-1")

	// Second tick should succeed because first already released.
	simulateTick("instance-2")

	if execCount != 2 {
		t.Errorf("expected 2 sequential executions, got %d", execCount)
	}
}

func TestWorkerSimulation_OverlappingTicks_BlockSecond(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()
	s := NewService(db)
	ctx := context.Background()

	execCount := 0
	var mu sync.Mutex

	// Acquire lock for instance-1 (simulating a running tick).
	got, _ := s.Acquire(ctx, "worker-lock", "instance-1", 5*time.Second)
	if !got {
		t.Fatal("instance-1 should acquire lock")
	}

	// instance-2 tries while instance-1 is still holding (not yet released).
	got2, _ := s.Acquire(ctx, "worker-lock", "instance-2", 5*time.Second)
	if got2 {
		t.Error("instance-2 should NOT acquire while instance-1 holds")
	}
	if got2 {
		mu.Lock()
		execCount++
		mu.Unlock()
	}

	// instance-1 releases.
	_ = s.Release(ctx, "worker-lock", "instance-1")

	// instance-3 should now succeed.
	got3, _ := s.Acquire(ctx, "worker-lock", "instance-3", 5*time.Second)
	if !got3 {
		t.Error("instance-3 should acquire after instance-1 released")
	}
	mu.Lock()
	execCount++
	mu.Unlock()
	_ = s.Release(ctx, "worker-lock", "instance-3")

	if execCount != 1 {
		t.Errorf("expected 1 execution (only instance-1 or instance-3), got %d", execCount)
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
