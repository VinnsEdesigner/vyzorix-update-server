package search

import (
	"context"
	"database/sql"
	"testing"

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
