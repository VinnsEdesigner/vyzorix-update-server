package permission

import "context"

// Grant is a persisted custom scoped permission assigned to a member within an
// organization. Custom grants are unioned on top of the member's role defaults
// at evaluation time.
type Grant struct {
	ID         string
	OperatorID string
	OrgID      string
	Action     Action
	Scope      string
	CreatedAt  int64
}

// GrantRepository persists custom scoped permission grants.
type GrantRepository interface {
	// Save upserts a grant (idempotent on operator+org+action+scope).
	Save(ctx context.Context, g *Grant) error
	// ListByOperatorOrg returns all custom grants for an operator within an org.
	ListByOperatorOrg(ctx context.Context, operatorID, orgID string) ([]*Grant, error)
	// Revoke deletes a grant by id, returning whether a grant was removed.
	Revoke(ctx context.Context, id string) (bool, error)
}
