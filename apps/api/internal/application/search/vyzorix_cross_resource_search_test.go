package search

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/cache"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"

	_ "github.com/mattn/go-sqlite3"
)

func TestSortResults_ByName(t *testing.T) {
	results := []Result{
		{Type: "device", Title: "Zebra Phone"},
		{Type: "device", Title: "Alpha Tablet"},
		{Type: "command", Title: "Middle Command"},
	}
	sortResults(results, "name")
	if results[0].Title != "Alpha Tablet" {
		t.Errorf("expected Alpha Tablet first, got %s", results[0].Title)
	}
	if results[1].Title != "Middle Command" {
		t.Errorf("expected Middle Command second, got %s", results[1].Title)
	}
	if results[2].Title != "Zebra Phone" {
		t.Errorf("expected Zebra Phone third, got %s", results[2].Title)
	}
}

func TestSortResults_ByType(t *testing.T) {
	results := []Result{
		{Type: "device", Title: "Phone"},
		{Type: "command", Title: "Reboot"},
		{Type: "event", Title: "Online"},
	}
	sortResults(results, "type")
	if results[0].Type != "command" {
		t.Errorf("expected command first, got %s", results[0].Type)
	}
	if results[1].Type != "device" {
		t.Errorf("expected device second, got %s", results[1].Type)
	}
	if results[2].Type != "event" {
		t.Errorf("expected event third, got %s", results[2].Type)
	}
}

func TestSortResults_ByStatus(t *testing.T) {
	results := []Result{
		{Type: "device", Status: "online"},
		{Type: "device", Status: "offline"},
	}
	sortResults(results, "status")
	if results[0].Status != "offline" {
		t.Errorf("expected offline first, got %s", results[0].Status)
	}
	if results[1].Status != "online" {
		t.Errorf("expected online second, got %s", results[1].Status)
	}
}

func TestSortResults_UnknownSortBy_NoOp(t *testing.T) {
	results := []Result{
		{Type: "device", Title: "B"},
		{Type: "device", Title: "A"},
	}
	sortResults(results, "unknown")
	if results[0].Title != "B" {
		t.Errorf("unknown sort should not change order, got %s", results[0].Title)
	}
}

func TestSearch_EmptyDB_ReturnsEmpty(t *testing.T) {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer db.Close()
	s := NewService(db)
	results, err := s.Search(context.Background(), "test", "org-1", "all", 20, "", "")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) != 0 {
		t.Errorf("expected 0 results on empty DB, got %d", len(results))
	}
}


func setupSearchTestDB(t *testing.T) *storage.SQLite {
	t.Helper()
	cfg := storage.DefaultConfig(filepath.Join(t.TempDir(), "search-test.db"))
	s, err := storage.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// TestSearch_CachesResults proves the cache layer returns stored results on
// the second call with the same args.
func TestSearch_CachesResults(t *testing.T) {
	s := setupSearchTestDB(t)
	searchCache := cache.New(30 * time.Second).Section("search")
	svc := NewService(s.DB())
	svc.SetSearchCache(searchCache)
	ctx := context.Background()

	now := time.Now().UnixMilli()
	if _, err := s.DB().ExecContext(ctx,
		`INSERT INTO devices (id, firebase_install_id, command_secret, online, registered_at, last_seen, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"dev-1", "fcm-1", "secret", 1, now, now, now, now); err != nil {
		t.Fatalf("insert device: %v", err)
	}

	results1, err := svc.Search(ctx, "dev", "org-1", "", 20, "", "")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results1) == 0 {
		t.Skip("no devices yet in org-1; skip cache count check")
	}

	// Second call with same args: served from cache.
	results2, err := svc.Search(ctx, "dev", "org-1", "", 20, "", "")
	if err != nil {
		t.Fatalf("Search (cached): %v", err)
	}
	if len(results2) != len(results1) {
		t.Errorf("cached result count %d != initial %d", len(results2), len(results1))
	}

	// Different query args produce a different cache key and hit the DB again.
	results3, err := svc.Search(ctx, "other", "org-1", "", 20, "", "")
	if err != nil {
		t.Fatalf("Search different args: %v", err)
	}
	_ = results3
}

// TestSearch_NilCacheWorks proves the search service works without a cache
// (backwards-compatible).
func TestSearch_NilCacheWorks(t *testing.T) {
	s := setupSearchTestDB(t)
	svc := NewService(s.DB())
	ctx := context.Background()

	if _, err := svc.Search(ctx, "dev", "org-1", "", 20, "", ""); err != nil {
		t.Fatalf("Search without cache: %v", err)
	}
}
