package configversion

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/configversion"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

func setupConfigVersionTestDB(t *testing.T) *storage.SQLite {
	t.Helper()
	cfg := storage.DefaultConfig(filepath.Join(t.TempDir(), "cv-test.db"))
	s, err := storage.Open(cfg)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestService_SnapshotAndList(t *testing.T) {
	s := setupConfigVersionTestDB(t)
	service := NewService(storage.NewConfigVersionRepository(s.DB()))
	ctx := context.Background()

	settings := map[string]interface{}{"timezone": "UTC", "alert_cooldown": 30}
	v1, err := service.Snapshot(ctx, "org-1", configversion.ResourceTypeSettings, settings, "operator-1")
	if err != nil {
		t.Fatalf("Snapshot v1: %v", err)
	}
	if v1 != 1 {
		t.Errorf("v1 = %d, want 1", v1)
	}

	v2, err := service.Snapshot(ctx, "org-1", configversion.ResourceTypeSettings, map[string]interface{}{"timezone": "Europe/Berlin"}, "operator-1")
	if err != nil {
		t.Fatalf("Snapshot v2: %v", err)
	}
	if v2 != 2 {
		t.Errorf("v2 = %d, want 2", v2)
	}

	items, err := service.List(ctx, "org-1", configversion.ResourceTypeSettings, 10)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 versions, got %d", len(items))
	}
	if items[0].Version != 2 || items[1].Version != 1 {
		t.Errorf("ordering wrong: %v", []int64{items[0].Version, items[1].Version})
	}

	got, err := service.Get(ctx, "org-1", configversion.ResourceTypeSettings, 1)
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got.Snapshot == "" {
		t.Error("expected snapshot content")
	}

	latest, err := service.Latest(ctx, "org-1", configversion.ResourceTypeSettings)
	if err != nil {
		t.Fatalf("Latest: %v", err)
	}
	if latest != 2 {
		t.Errorf("latest = %d, want 2", latest)
	}
}

func TestDeviceSettingsHook_SnapshotOnWrite(t *testing.T) {
	s := setupConfigVersionTestDB(t)
	service := NewService(storage.NewConfigVersionRepository(s.DB()))
	ctx := context.Background()

	// Snapshot the type used by device settings (ResourceTypeThresholds).
	deviceSettings := map[string]interface{}{
		"thresholds": map[string]int{"risk_warn": 70, "risk_crit": 90},
		"metadata":   map[string]string{"source": "device"},
	}
	version, err := service.Snapshot(ctx, "org-1", configversion.ResourceTypeThresholds, deviceSettings, "device")
	if err != nil {
		t.Fatalf("Snapshot: %v", err)
	}
	if version != 1 {
		t.Errorf("version = %d, want 1", version)
	}
}

// recordingVersionSvc captures Snapshot calls for hook tests.
type recordingVersionSvc struct {
	service *Service
	calls   []snapshotCall
}

type snapshotCall struct {
	orgID      string
	targetPtr  interface{}
	resource   configversion.ResourceType
	changedBy  string
}

func (r *recordingVersionSvc) Snapshot(ctx context.Context, orgID string, resourceType configversion.ResourceType, snapshot interface{}, changedBy string) (int64, error) {
	r.calls = append(r.calls, snapshotCall{orgID: orgID, targetPtr: snapshot, resource: resourceType, changedBy: changedBy})
	return r.service.Snapshot(ctx, orgID, resourceType, snapshot, changedBy)
}


func ptr[T any](v T) *T { return &v }

func TestService_ResourceTypeIsolation(t *testing.T) {
	s := setupConfigVersionTestDB(t)
	service := NewService(storage.NewConfigVersionRepository(s.DB()))
	ctx := context.Background()

	settings := map[string]string{"timezone": "UTC"}
	thresholds := map[string]int{"risk_warn": 60}

	_, err := service.Snapshot(ctx, "org-1", configversion.ResourceTypeSettings, settings, "op")
	if err != nil {
		t.Fatalf("Snapshot settings: %v", err)
	}
	_, err = service.Snapshot(ctx, "org-1", configversion.ResourceTypeThresholds, thresholds, "op")
	if err != nil {
		t.Fatalf("Snapshot thresholds: %v", err)
	}

	settingsList, _ := service.List(ctx, "org-1", configversion.ResourceTypeSettings, 10)
	thresholdsList, _ := service.List(ctx, "org-1", configversion.ResourceTypeThresholds, 10)

	if len(settingsList) != 1 || len(thresholdsList) != 1 {
		t.Errorf("isolation failed: settings=%d, thresholds=%d", len(settingsList), len(thresholdsList))
	}
	if settingsList[0].ResourceType != configversion.ResourceTypeSettings {
		t.Errorf("settings resource type wrong")
	}
	if thresholdsList[0].ResourceType != configversion.ResourceTypeThresholds {
		t.Errorf("thresholds resource type wrong")
	}
}
