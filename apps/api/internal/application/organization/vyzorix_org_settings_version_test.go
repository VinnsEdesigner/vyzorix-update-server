package organization

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	configversionapp "github.com/VinnsEdesigner/vyzorix/apps/api/internal/application/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

func setupOrgSettingsTestDB(t *testing.T) *storage.SQLite {
	t.Helper()
	cfg := storage.DefaultConfig(filepath.Join(t.TempDir(), "org-settings-test.db"))
	s, err := storage.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func seedOrg(t *testing.T, s *storage.SQLite) {
	t.Helper()
	ctx := context.Background()
	now := time.Now().UnixMilli()

	if _, err := s.DB().ExecContext(ctx,
		`INSERT INTO operators (id, email, password_hash, name, role, email_verified, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"op-1", "op@test.local", "hash", "Operator", "admin", 1, now, now); err != nil {
		t.Fatalf("insert operator: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx,
		`INSERT INTO organizations (id, name, created_by, created_at, updated_at, is_active, max_members) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		"org-1", "Test Org", "op-1", now, now, 1, 100); err != nil {
		t.Fatalf("insert org: %v", err)
	}
	if _, err := s.DB().ExecContext(ctx,
		`INSERT INTO organization_settings (id, organization_id, timezone, date_format, alert_cooldown_minutes, default_thresholds, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		"settings-1", "org-1", "UTC", "YYYY-MM-DD", 30, "{}", now, now); err != nil {
		t.Fatalf("insert settings: %v", err)
	}
}

func TestOrgSettingsService_VersionsOnWrite(t *testing.T) {
	s := setupOrgSettingsTestDB(t)
	versionSvc := configversionapp.NewService(storage.NewConfigVersionRepository(s.DB()))
	orgSvc := NewOrganizationSettingsService(storage.NewOrganizationSettingsRepository(s.DB()), nil, nil)
	orgSvc.SetVersionService(versionSvc)
	ctx := context.Background()

	seedOrg(t, s)

	tz := "UTC"
	if _, err := orgSvc.UpdateSettings(ctx, "org-1", &organization.UpdateOrganizationSettingsRequest{Timezone: &tz}); err != nil {
		t.Fatalf("UpdateSettings: %v", err)
	}

	versions, err := versionSvc.List(ctx, "org-1", configversion.ResourceTypeSettings, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}

	// Thresholds produce their own type.
	hookThresholds := organization.DefaultThresholds()

	if _, err := versionSvc.Snapshot(ctx, "org-1", configversion.ResourceTypeThresholds, hookThresholds, "operator"); err != nil {
		t.Fatalf("Snapshot thresholds: %v", err)
	}
	thresholdsVersions, err := versionSvc.List(ctx, "org-1", configversion.ResourceTypeThresholds, 10)
	if err != nil {
		t.Fatalf("List thresholds: %v", err)
	}
	if len(thresholdsVersions) != 1 {
		t.Fatalf("expected 1 thresholds version, got %d", len(thresholdsVersions))
	}
	if versions[0].ResourceType == thresholdsVersions[0].ResourceType {
		t.Error("settings and thresholds versions should be distinct types")
	}
}

func TestOrgSettingsService_RestoreRoundTrip(t *testing.T) {
	s := setupOrgSettingsTestDB(t)
	versionSvc := configversionapp.NewService(storage.NewConfigVersionRepository(s.DB()))
	orgSvc := NewOrganizationSettingsService(storage.NewOrganizationSettingsRepository(s.DB()), nil, nil)
	orgSvc.SetVersionService(versionSvc)
	ctx := context.Background()

	seedOrg(t, s)

	// Write initial state.
	tz := "UTC"
	if _, err := orgSvc.UpdateSettings(ctx, "org-1", &organization.UpdateOrganizationSettingsRequest{Timezone: &tz}); err != nil {
		t.Fatalf("write v1: %v", err)
	}

	// Change it.
	tz2 := "Europe/Berlin"
	if _, err := orgSvc.UpdateSettings(ctx, "org-1", &organization.UpdateOrganizationSettingsRequest{Timezone: &tz2}); err != nil {
		t.Fatalf("write v2: %v", err)
	}

	// Restore v1.
	restored, err := orgSvc.RestoreSettings(ctx, "org-1", `{"Timezone":"UTC"}`)
	if err != nil {
		t.Fatalf("Restore: %v", err)
	}
	if restored.Timezone != "UTC" {
		t.Errorf("restored timezone = %q, want UTC", restored.Timezone)
	}

	// Restore creates a new version (round-trip).
	versions, err := versionSvc.List(ctx, "org-1", configversion.ResourceTypeSettings, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(versions) < 3 {
		t.Errorf("expected at least 3 versions (v1 write, v2 write, restore), got %d", len(versions))
	}
}
