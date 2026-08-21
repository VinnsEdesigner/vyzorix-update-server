package usagestats

import (
	"context"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/featuremgmt"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/storage"
)

func TestService_CollectAndSnapshot(t *testing.T) {
	c, err := storage.Open(storage.DefaultConfig(t.TempDir() + "/usagestats.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = c.Close() }()

	svc := NewService(NewCollector(c.DB()), nil)
	if svc.Snapshot() != nil {
		t.Fatal("expected nil snapshot before first collect")
	}

	svc.Collect(context.Background())
	snap := svc.Snapshot()
	if snap == nil {
		t.Fatal("expected snapshot after collect")
	}
	if snap.Counts.Devices < 0 || snap.Counts.Operators < 0 {
		t.Errorf("unexpected counts: %+v", snap.Counts)
	}
	if snap.CollectedAt.IsZero() {
		t.Error("collected_at should be stamped")
	}
	if len(snap.Toggles) != 0 {
		t.Errorf("no features passed: toggles should stay empty, got %+v", snap.Toggles)
	}
}

func TestService_CollectRecordsToggles(t *testing.T) {
	c, err := storage.Open(storage.DefaultConfig(t.TempDir() + "/usagestats.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = c.Close() }()

	features := featuremgmt.NewManager(map[featuremgmt.Feature]bool{
		featuremgmt.ScopedRBAC:   true,
		featuremgmt.DeviceGroups: true,
	})
	svc := NewService(NewCollector(c.DB()), features)
	svc.Collect(context.Background())

	snap := svc.Snapshot()
	if !snap.Toggles["scoped_rbac"] || !snap.Toggles["device_groups"] {
		t.Errorf("expected declared toggles true: %+v", snap.Toggles)
	}
	if snap.Toggles["server_driven_org"] {
		t.Errorf("undeclared toggle should default false: %+v", snap.Toggles)
	}
}

func TestCollector_Counts(t *testing.T) {
	c, err := storage.Open(storage.DefaultConfig(t.TempDir() + "/usagestats.db"))
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer func() { _ = c.Close() }()

	snap := NewCollector(c.DB()).Query(context.Background())
	// Fresh database: tables exist, counts are zero.
	if snap == nil || snap.Counts.Devices != 0 {
		t.Errorf("fresh counts should be zero: %+v", snap)
	}
}
