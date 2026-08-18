package storage

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/device_group"
)

// newTestDeviceGroupDB opens an in-memory SQLite database with the real
// device_groups schema (no mocks; the repository is tested end-to-end).
func newTestDeviceGroupDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE device_groups (
			id TEXT PRIMARY KEY, org_id TEXT NOT NULL, name TEXT NOT NULL, created_at INTEGER NOT NULL)`,
		`CREATE TABLE device_group_members (
			group_id TEXT NOT NULL, operator_id TEXT NOT NULL, created_at INTEGER NOT NULL,
			PRIMARY KEY (group_id, operator_id))`,
		`CREATE TABLE device_group_devices (
			group_id TEXT NOT NULL, device_id TEXT NOT NULL, created_at INTEGER NOT NULL,
			PRIMARY KEY (group_id, device_id))`,
	}
	for _, q := range stmts {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}
	return db
}

func TestDeviceGroupRepository_GroupLifecycleAndMembership(t *testing.T) {
	db := newTestDeviceGroupDB(t)
	r := NewDeviceGroupRepository(db)
	ctx := context.Background()

	g := &device_group.Group{ID: "g1", OrgID: "org1", Name: "field-ops", CreatedAt: time.Now()}
	if err := r.Save(ctx, g); err != nil {
		t.Fatalf("Save group: %v", err)
	}

	got, err := r.GetByID(ctx, "g1")
	if err != nil {
		t.Fatalf("GetByID: %v", err)
	}
	if got.Name != "field-ops" || got.OrgID != "org1" {
		t.Errorf("GetByID = %+v, want field-ops/org1", got)
	}

	if err := r.AddMember(ctx, "g1", "op-1"); err != nil {
		t.Fatalf("AddMember: %v", err)
	}
	member, err := r.IsMember(ctx, "g1", "op-1")
	if err != nil || !member {
		t.Errorf("IsMember after add = %v, %v, want true", member, err)
	}

	// Idempotent re-add.
	if err := r.AddMember(ctx, "g1", "op-1"); err != nil {
		t.Fatalf("re-AddMember: %v", err)
	}

	ok, err := r.IsMember(ctx, "g1", "op-2")
	if err != nil || ok {
		t.Errorf("IsMember op-2 = %v, %v, want false", ok, err)
	}

	removed, err := r.RemoveMember(ctx, "g1", "op-1")
	if err != nil || !removed {
		t.Fatalf("RemoveMember = %v, %v, want true", removed, err)
	}
}

func TestDeviceGroupRepository_DeviceAssignmentAndLookup(t *testing.T) {
	db := newTestDeviceGroupDB(t)
	r := NewDeviceGroupRepository(db)
	ctx := context.Background()

	_ = r.Save(ctx, &device_group.Group{ID: "g1", OrgID: "org1", Name: "noc", CreatedAt: time.Now()})
	_ = r.Save(ctx, &device_group.Group{ID: "g2", OrgID: "org1", Name: "hq", CreatedAt: time.Now()})

	if err := r.AddDevice(ctx, "g1", "dev-1"); err != nil {
		t.Fatalf("AddDevice g1: %v", err)
	}
	if err := r.AddDevice(ctx, "g2", "dev-1"); err != nil {
		t.Fatalf("AddDevice g2: %v", err)
	}

	ids, err := r.GroupIDsForDevice(ctx, "dev-1")
	if err != nil {
		t.Fatalf("GroupIDsForDevice: %v", err)
	}
	if len(ids) != 2 {
		t.Errorf("GroupIDsForDevice = %v, want 2 groups (g1,g2)", ids)
	}
}

func TestDeviceGroupRepository_RemoveDeviceClearsLookup(t *testing.T) {
	db := newTestDeviceGroupDB(t)
	r := NewDeviceGroupRepository(db)
	ctx := context.Background()

	_ = r.Save(ctx, &device_group.Group{ID: "g1", OrgID: "org1", Name: "noc", CreatedAt: time.Now()})
	_ = r.AddDevice(ctx, "g1", "dev-1")

	removed, err := r.RemoveDevice(ctx, "g1", "dev-1")
	if err != nil || !removed {
		t.Fatalf("RemoveDevice = %v, %v, want true", removed, err)
	}

	ids, err := r.GroupIDsForDevice(ctx, "dev-1")
	if err != nil {
		t.Fatalf("GroupIDsForDevice: %v", err)
	}
	if len(ids) != 0 {
		t.Errorf("GroupIDsForDevice after remove = %v, want empty", ids)
	}
}
