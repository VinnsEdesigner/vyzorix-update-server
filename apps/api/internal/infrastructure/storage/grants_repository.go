package storage

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
)

// GrantRepository is the persistence implementation for ResourcePermissions.
// It optionally caches ListEffective results in a TTL cache to skip the
// database on repeated authorization checks within the TTL window.
type GrantRepository struct {
	db    *sql.DB
	cache *permission.Cache
}

// NewGrantRepository creates a new GrantRepository without caching.
func NewGrantRepository(db *sql.DB) *GrantRepository {
	return &GrantRepository{db: db}
}

// NewGrantRepositoryWithCache creates a GrantRepository with a permission cache.
func NewGrantRepositoryWithCache(db *sql.DB, cache *permission.Cache) *GrantRepository {
	return &GrantRepository{db: db, cache: cache}
}

func actionsColumn(actions []permission.Action) string {
	parts := make([]string, len(actions))
	for i, a := range actions {
		parts[i] = string(a)
	}
	return strings.Join(parts, ",")
}

func parseActions(col string) []permission.Action {
	if col == "" {
		return nil
	}
	parts := strings.Split(col, ",")
	out := make([]permission.Action, 0, len(parts))
	for _, p := range parts {
		out = append(out, permission.Action(strings.TrimSpace(p)))
	}
	return out
}

func scanPermission(scanner interface{ Scan(...any) error }) (*permission.ResourcePermission, error) {
	var p permission.ResourcePermission
	var actionsCol string
	var managed, inherited int
	if err := scanner.Scan(&p.ID, &p.OrgID, &p.SubjectType, &p.SubjectID, &actionsCol, &p.Scope, &managed, &inherited, &p.CreatedAt, &p.UpdatedAt); err != nil {
		return nil, err
	}
	p.Actions = parseActions(actionsCol)
	p.IsManaged = managed != 0
	p.IsInherited = inherited != 0
	return &p, nil
}

// Save upserts a ResourcePermission (idempotent on org+subject+scope).
func (r *GrantRepository) Save(ctx context.Context, p *permission.ResourcePermission) error {
	now := time.Now().UnixMilli()
	if p.CreatedAt == 0 {
		p.CreatedAt = now
	}
	p.UpdatedAt = now
	managed := 0
	if p.IsManaged {
		managed = 1
	}
	inherited := 0
	if p.IsInherited {
		inherited = 1
	}
	query := `
		INSERT INTO resource_permissions (id, org_id, subject_type, subject_id, actions, scope, is_managed, is_inherited, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(org_id, subject_type, subject_id, scope) DO UPDATE SET actions = excluded.actions, is_managed = excluded.is_managed, is_inherited = excluded.is_inherited, updated_at = excluded.updated_at
	`
	if _, err := r.db.ExecContext(ctx, query, p.ID, p.OrgID, string(p.SubjectType), p.SubjectID, actionsColumn(p.Actions), p.Scope, managed, inherited, p.CreatedAt, p.UpdatedAt); err != nil {
		return err
	}
	r.invalidateAfterMutation(p, ctx)
	return nil
}

// ListEffective returns all grants an operator can act with in an org: their
// operator-direct grants UNION grants on teams they belong to. A single query
// joins through device_group_members so the evaluator sees the full effective
// set without N+1 lookups. Results are cached within the TTL window.
func (r *GrantRepository) ListEffective(ctx context.Context, operatorID, orgID string) ([]*permission.ResourcePermission, error) {
	if cached := r.cache.Get(operatorID, orgID); cached != nil {
		return cached, nil
	}
	query := `
		SELECT id, org_id, subject_type, subject_id, actions, scope, is_managed, is_inherited, created_at, updated_at
		FROM resource_permissions
		WHERE org_id = ? AND subject_type = 'operator' AND subject_id = ?
		UNION ALL
		SELECT rp.id, rp.org_id, rp.subject_type, rp.subject_id, rp.actions, rp.scope, rp.is_managed, rp.is_inherited, rp.created_at, rp.updated_at
		FROM resource_permissions rp
		INNER JOIN device_group_members dgm ON dgm.group_id = rp.subject_id
		WHERE rp.org_id = ? AND rp.subject_type = 'team' AND dgm.operator_id = ?
	`
	rows, err := r.db.QueryContext(ctx, query, orgID, operatorID, orgID, operatorID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []*permission.ResourcePermission
	for rows.Next() {
		p, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	r.cache.Put(operatorID, orgID, out)
	return out, nil
}

// ListByOrg returns all ResourcePermissions in an org (admin view).
func (r *GrantRepository) ListByOrg(ctx context.Context, orgID string) ([]*permission.ResourcePermission, error) {
	query := `SELECT id, org_id, subject_type, subject_id, actions, scope, is_managed, is_inherited, created_at, updated_at FROM resource_permissions WHERE org_id = ?`
	rows, err := r.db.QueryContext(ctx, query, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var out []*permission.ResourcePermission
	for rows.Next() {
		p, err := scanPermission(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

// Revoke deletes a grant by id, returning whether one was removed.
func (r *GrantRepository) Revoke(ctx context.Context, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM resource_permissions WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if n > 0 {
		r.cache.InvalidateAll()
	}
	return n > 0, err
}

// RevokeForSubject removes all grants for a subject (e.g. when a team is deleted).
func (r *GrantRepository) RevokeForSubject(ctx context.Context, subjectType permission.SubjectType, subjectID string) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM resource_permissions WHERE subject_type = ? AND subject_id = ?`, string(subjectType), subjectID)
	if err == nil {
		r.cache.InvalidateAll()
	}
	return err
}

// invalidateAfterMutation clears cache entries affected by a grant mutation.
// An operator-direct grant invalidates that operator's entries; a team grant
// invalidates the entire org (since team grants affect every team member).
func (r *GrantRepository) invalidateAfterMutation(p *permission.ResourcePermission, _ context.Context) {
	if r.cache == nil {
		return
	}
	if p.SubjectType == permission.SubjectTeam {
		r.cache.InvalidateOrg(p.OrgID)
		return
	}
	r.cache.InvalidateAll()
}
