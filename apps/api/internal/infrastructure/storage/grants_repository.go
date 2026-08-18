package storage

import (
	"context"
	"database/sql"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
)

// GrantRepository is the persistence implementation for custom scoped grants.
type GrantRepository struct {
	db *sql.DB
}

// NewGrantRepository creates a new GrantRepository.
func NewGrantRepository(db *sql.DB) *GrantRepository {
	return &GrantRepository{db: db}
}

// Save upserts a grant (idempotent on operator+org+action+scope).
func (r *GrantRepository) Save(ctx context.Context, g *permission.Grant) error {
	if g.CreatedAt == 0 {
		g.CreatedAt = time.Now().UnixMilli()
	}
	query := `
		INSERT INTO permission_grants (id, operator_id, org_id, action, scope, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
		ON CONFLICT(operator_id, org_id, action, scope) DO UPDATE SET created_at = excluded.created_at
	`
	_, err := r.db.ExecContext(ctx, query, g.ID, g.OperatorID, g.OrgID, string(g.Action), g.Scope, g.CreatedAt)
	return err
}

// ListByOperatorOrg returns all custom grants for an operator within an org.
func (r *GrantRepository) ListByOperatorOrg(ctx context.Context, operatorID, orgID string) ([]*permission.Grant, error) {
	query := `SELECT id, operator_id, org_id, action, scope, created_at FROM permission_grants WHERE operator_id = ? AND org_id = ?`
	rows, err := r.db.QueryContext(ctx, query, operatorID, orgID)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var grants []*permission.Grant
	for rows.Next() {
		g := &permission.Grant{}
		if err := rows.Scan(&g.ID, &g.OperatorID, &g.OrgID, &g.Action, &g.Scope, &g.CreatedAt); err != nil {
			return nil, err
		}
		grants = append(grants, g)
	}
	return grants, rows.Err()
}

// Revoke deletes a grant by id, returning whether a grant was removed.
func (r *GrantRepository) Revoke(ctx context.Context, id string) (bool, error) {
	res, err := r.db.ExecContext(ctx, `DELETE FROM permission_grants WHERE id = ?`, id)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}
