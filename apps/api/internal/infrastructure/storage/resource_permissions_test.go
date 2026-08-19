package storage

import (
	"context"
	"database/sql"
	"testing"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
)

// newTestResourcePermissionDB opens an in-memory SQLite database with the real
// resource_permissions + device_group_members schema (no mocks; the repository
// is tested end-to-end, including the team-grant join).
func newTestResourcePermissionDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("open in-memory sqlite: %v", err)
	}
	stmts := []string{
		`CREATE TABLE resource_permissions (
			id TEXT PRIMARY KEY, org_id TEXT NOT NULL, subject_type TEXT NOT NULL,
			subject_id TEXT NOT NULL, actions TEXT NOT NULL, scope TEXT NOT NULL,
			is_managed INTEGER NOT NULL DEFAULT 0, is_inherited INTEGER NOT NULL DEFAULT 0,
			created_at INTEGER NOT NULL, updated_at INTEGER NOT NULL)`,
		`CREATE TABLE device_group_members (
			group_id TEXT NOT NULL, operator_id TEXT NOT NULL, created_at INTEGER NOT NULL,
			PRIMARY KEY (group_id, operator_id))`,
		`CREATE UNIQUE INDEX idx_rp_test_unique ON resource_permissions(org_id, subject_type, subject_id, scope)`,
	}
	for _, q := range stmts {
		if _, err := db.Exec(q); err != nil {
			t.Fatalf("create schema: %v", err)
		}
	}
	return db
}

func TestGrantRepository_SaveAndListEffectiveOperatorGrants(t *testing.T) {
	db := newTestResourcePermissionDB(t)
	r := NewGrantRepository(db)
	ctx := context.Background()

	p := &permission.ResourcePermission{
		ID:          "rp1",
		OrgID:       "org1",
		SubjectType: permission.SubjectOperator,
		SubjectID:   "op-1",
		Actions:     []permission.Action{permission.ActionDeviceRead, permission.ActionDeviceWrite},
		Scope:       permission.WildcardScope(permission.ScopeDevices),
	}
	if err := r.Save(ctx, p); err != nil {
		t.Fatalf("Save: %v", err)
	}

	// Idempotent re-save with expanded actions.
	p.Actions = []permission.Action{permission.ActionDeviceRead}
	if err := r.Save(ctx, p); err != nil {
		t.Fatalf("re-Save: %v", err)
	}

	effective, err := r.ListEffective(ctx, "op-1", "org1")
	if err != nil {
		t.Fatalf("ListEffective: %v", err)
	}
	if len(effective) != 1 {
		t.Fatalf("ListEffective = %d grants, want 1", len(effective))
	}
	if effective[0].SubjectType != permission.SubjectOperator {
		t.Errorf("subject = %v, want operator", effective[0].SubjectType)
	}
	if len(effective[0].Actions) != 1 || effective[0].Actions[0] != permission.ActionDeviceRead {
		t.Errorf("actions = %v, want [device.read]", effective[0].Actions)
	}
}

func TestGrantRepository_TeamGrantsJoinedViaMembership(t *testing.T) {
	db := newTestResourcePermissionDB(t)
	r := NewGrantRepository(db)
	ctx := context.Background()

	// Grant to a team.
	if err := r.Save(ctx, &permission.ResourcePermission{
		ID:          "rp-team",
		OrgID:       "org1",
		SubjectType: permission.SubjectTeam,
		SubjectID:   "team-1",
		Actions:     []permission.Action{permission.ActionCommand},
		Scope:       permission.WildcardScope(permission.ScopeDevices),
	}); err != nil {
		t.Fatalf("Save team grant: %v", err)
	}

	// op-1 is NOT in team-1 → should see no team grant.
	effective, err := r.ListEffective(ctx, "op-1", "org1")
	if err != nil {
		t.Fatalf("ListEffective: %v", err)
	}
	if len(effective) != 0 {
		t.Errorf("ListEffective for non-member = %d, want 0", len(effective))
	}

	// Add op-1 to team-1.
	if _, err := db.Exec(`INSERT INTO device_group_members (group_id, operator_id, created_at) VALUES (?, ?, 0)`, "team-1", "op-1"); err != nil {
		t.Fatalf("insert membership: %v", err)
	}

	// Now op-1 should see the team grant.
	effective, err = r.ListEffective(ctx, "op-1", "org1")
	if err != nil {
		t.Fatalf("ListEffective after join: %v", err)
	}
	if len(effective) != 1 {
		t.Fatalf("ListEffective after join = %d, want 1", len(effective))
	}
	if effective[0].SubjectType != permission.SubjectTeam || effective[0].SubjectID != "team-1" {
		t.Errorf("team grant = %+v, want team-1", effective[0])
	}
}

func TestGrantRepository_RevokeAndRevokeForSubject(t *testing.T) {
	db := newTestResourcePermissionDB(t)
	r := NewGrantRepository(db)
	ctx := context.Background()

	_ = r.Save(ctx, &permission.ResourcePermission{ID: "rp1", OrgID: "org1", SubjectType: permission.SubjectOperator, SubjectID: "op-1", Actions: []permission.Action{permission.ActionDeviceRead}, Scope: "devices:*"})
	_ = r.Save(ctx, &permission.ResourcePermission{ID: "rp2", OrgID: "org1", SubjectType: permission.SubjectTeam, SubjectID: "team-1", Actions: []permission.Action{permission.ActionCommand}, Scope: "devices:*"})

	removed, err := r.Revoke(ctx, "rp1")
	if err != nil || !removed {
		t.Fatalf("Revoke rp1 = %v, %v, want true", removed, err)
	}
	removed, err = r.Revoke(ctx, "rp1")
	if err != nil || removed {
		t.Errorf("Revoke again = %v, %v, want false", removed, err)
	}

	if err := r.RevokeForSubject(ctx, permission.SubjectTeam, "team-1"); err != nil {
		t.Fatalf("RevokeForSubject: %v", err)
	}
	all, _ := r.ListByOrg(ctx, "org1")
	if len(all) != 0 {
		t.Errorf("ListByOrg after revoke-for-subject = %d, want 0", len(all))
	}
}

func TestExpandActionSets_GrantsAggregateToBase(t *testing.T) {
	scopes := permission.ScopedPermissions{
		{Action: permission.ActionSetDeviceManage, Scope: permission.WildcardScope(permission.ScopeDevices)},
	}
	expanded := permission.ExpandActionSets(scopes)
	eval := permission.NewEvaluatorWithScopes(expanded)

	// device.manage should satisfy device.read, device.write, device.delete, device.assign.
	for _, a := range []permission.Action{permission.ActionDeviceRead, permission.ActionDeviceWrite, permission.ActionDeviceDelete, permission.ActionDeviceAssign} {
		if !eval.Grants(a, permission.WildcardScope(permission.ScopeDevices)) {
			t.Errorf("device.manage should grant %s, got false", a)
		}
	}
	// device.manage should NOT grant command.execute.
	if eval.Grants(permission.ActionCommand, permission.WildcardScope(permission.ScopeDevices)) {
		t.Error("device.manage should not grant command.execute")
	}
}
