package hub

import (
	"context"
	"database/sql"
	"encoding/json"
	"path/filepath"
	"strconv"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/command"
)

// newTestQueue builds a MessageQueue backed by a real SQLite DB, so persistence and
// replay are exercised end to end.
func newTestQueue(t *testing.T) *MessageQueue {
	t.Helper()

	db, err := sql.Open("sqlite3", filepath.Join(t.TempDir(), "q.db"))
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// A file-backed DB is required: ":memory:" gives every pooled connection its
	// own database, so the async persistence worker would not see this table.
	if _, err := db.Exec(`
		CREATE TABLE message_queue (
			id TEXT PRIMARY KEY,
			device_id TEXT NOT NULL,
			frame_json TEXT NOT NULL,
			enqueued_at INTEGER NOT NULL,
			expires_at INTEGER NOT NULL
		)
	`); err != nil {
		t.Fatalf("create table: %v", err)
	}

	return NewMessageQueue(testLogger(), db, DefaultMessageQueueConfig())
}

// persistInMemoryToDB copies whatever is in the in-memory queue into the database
// synchronously, standing in for the async persistence worker so the test does not
// race it. This reproduces Enqueue's dual-write (channel + database).
func persistInMemoryToDB(t *testing.T, q *MessageQueue, deviceID string) {
	t.Helper()

	ch := q.GetOrCreateQueue(deviceID)

	var msgs []*QueuedMessage
	for {
		select {
		case m := <-ch:
			msgs = append(msgs, m)
		default:
			for _, m := range msgs {
				// The async worker may already have persisted this id; OR REPLACE
				// keeps the helper idempotent.
				if err := q.upsertPersistedForTest(m); err != nil {
					t.Fatalf("persist: %v", err)
				}
				ch <- m
			}
			return
		}
	}
}

// upsertPersistedForTest persists a message, tolerating a row the async worker may
// already have written for the same id.
func (q *MessageQueue) upsertPersistedForTest(m *QueuedMessage) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	frameJSON, err := json.Marshal(m.Frame)
	if err != nil {
		return err
	}

	_, err = q.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO message_queue (id, device_id, frame_json, enqueued_at, expires_at)
		 VALUES (?, ?, ?, ?, ?)`,
		m.ID, m.DeviceID, string(frameJSON),
		m.EnqueuedAt.UnixMilli(), m.ExpiresAt.UnixMilli(),
	)
	return err
}

// TestReplayQueue_DeliversEachMessageOnce is the regression guard for a duplicate
// delivery bug: Enqueue writes a message to BOTH the in-memory channel and the
// database, so a replay that drains both sources would deliver a queued command
// twice.
func TestReplayQueue_DeliversEachMessageOnce(t *testing.T) {
	q := newTestQueue(t)

	if !q.Enqueue("dev-1", frame("dispatch-1")) {
		t.Fatal("enqueue should succeed")
	}
	persistInMemoryToDB(t, q, "dev-1")

	if got := len(q.LoadPersistedMessages("dev-1")); got != 1 {
		t.Fatalf("expected 1 persisted message, got %d", got)
	}

	dest := make(chan command.CommandFrame, 10)
	res := q.ReplayQueue("dev-1", dest)

	if res.Count != 1 {
		t.Fatalf("replay must deliver the message exactly once, delivered %d", res.Count)
	}
	if len(dest) != 1 {
		t.Fatalf("expected 1 frame in the destination, got %d", len(dest))
	}
}

// TestReplayQueue_NoDuplicatesWithManyMessages proves the invariant holds for a
// batch, not just a single message.
func TestReplayQueue_NoDuplicatesWithManyMessages(t *testing.T) {
	q := newTestQueue(t)

	const n = 5
	for i := 0; i < n; i++ {
		q.Enqueue("dev-1", frame("dispatch-"+strconv.Itoa(i)))
	}
	persistInMemoryToDB(t, q, "dev-1")

	dest := make(chan command.CommandFrame, 50)
	res := q.ReplayQueue("dev-1", dest)

	if res.Count != n {
		t.Fatalf("expected %d deliveries, got %d", n, res.Count)
	}

	seen := map[string]int{}
	for len(dest) > 0 {
		f := <-dest
		seen[f.DispatchID]++
	}
	for id, count := range seen {
		if count != 1 {
			t.Fatalf("message %s delivered %d times", id, count)
		}
	}
}
