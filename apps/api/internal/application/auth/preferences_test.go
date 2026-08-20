package auth

import (
	"context"
	"database/sql"
	"encoding/json"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
	_ "github.com/mattn/go-sqlite3"
)

func setupPrefsTestDB(t *testing.T) (*storage.OperatorRepository, *sql.DB) {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	_, err = db.Exec(`CREATE TABLE operator_settings (operator_id TEXT PRIMARY KEY, risk_warn INTEGER DEFAULT 70, risk_crit INTEGER DEFAULT 85, thermal_warn INTEGER DEFAULT 45, thermal_crit INTEGER DEFAULT 50, buffer_warn INTEGER DEFAULT 30, buffer_crit INTEGER DEFAULT 15, preferences TEXT DEFAULT '{}')`)
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	repo := storage.NewOperatorRepository(db)
	return repo, db
}

func TestGetPreferences_NoData_ReturnsDefaults(t *testing.T) {
	repo, db := setupPrefsTestDB(t)
	defer db.Close()
	svc := NewClientSettingsService(repo)
	ctx := context.Background()

	prefs, err := svc.GetPreferences(ctx, "op-nonexistent")
	if err != nil {
		t.Fatalf("expected nil error for missing operator, got: %v", err)
	}
	if prefs["theme"] != "dark" {
		t.Errorf("expected default theme 'dark', got %v", prefs["theme"])
	}
	if prefs["language"] != "en" {
		t.Errorf("expected default language 'en', got %v", prefs["language"])
	}
}

func TestUpdatePreferences_ThenGet(t *testing.T) {
	repo, db := setupPrefsTestDB(t)
	defer db.Close()
	svc := NewClientSettingsService(repo)
	ctx := context.Background()

	updates := map[string]any{
		"theme":            "light",
		"refresh_interval": 60,
	}
	_, err := svc.UpdatePreferences(ctx, "op-1", updates)
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	prefs, err := svc.GetPreferences(ctx, "op-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if prefs["theme"] != "light" {
		t.Errorf("expected theme 'light', got %v", prefs["theme"])
	}
	if prefs["refresh_interval"] != float64(60) {
		t.Errorf("expected refresh_interval float64(60), got %v (%T)", prefs["refresh_interval"], prefs["refresh_interval"])
	}
	if prefs["language"] != "en" {
		t.Errorf("expected default language 'en' to be merged in, got %v", prefs["language"])
	}
}

func TestUpdatePreferences_DeleteCustomKey(t *testing.T) {
	repo, db := setupPrefsTestDB(t)
	defer db.Close()
	svc := NewClientSettingsService(repo)
	ctx := context.Background()

	_, _ = svc.UpdatePreferences(ctx, "op-1", map[string]any{"custom_key": "custom_value"})
	_, err := svc.UpdatePreferences(ctx, "op-1", map[string]any{"custom_key": nil})
	if err != nil {
		t.Fatalf("update with nil: %v", err)
	}

	raw, _ := repo.GetPreferencesRaw(ctx, "op-1")
	if raw == "" {
		t.Fatal("expected raw preferences to be saved")
	}
	var stored map[string]any
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if _, exists := stored["custom_key"]; exists {
		t.Error("custom_key should be absent in stored JSON after nil update")
	}
}
