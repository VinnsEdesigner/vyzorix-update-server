package permission

import "context"

// SubjectType identifies whom a ResourcePermission is granted to.
type SubjectType string

const (
	SubjectOperator SubjectType = "operator"
	SubjectTeam     SubjectType = "team"
	SubjectBuiltin  SubjectType = "builtin"
)

// ResourcePermission is a first-class persisted permission grant: one or more
// actions on a scope, granted to a subject (operator, team, or built-in role)
// within an organization. This mirrors Grafana's ResourcePermission, which is
// the unit of grant in its RBAC system — assignable to users OR teams OR
// built-in roles, with managed/inherited lifecycle flags.
type ResourcePermission struct {
	OrgID       string
	ID          string
	SubjectType SubjectType
	SubjectID   string
	Scope       string
	Actions     []Action
	IsManaged   bool
	IsInherited bool
	CreatedAt   int64
	UpdatedAt   int64
}

// Grant is a compatibility alias for ResourcePermission restricted to a single
// action granted to an operator. Existing call sites that build a Grant continue
// to work; new code should construct ResourcePermission directly.
type Grant = ResourcePermission

// GrantRepository persists ResourcePermissions. The canonical loader is
// ListEffective, which assembles an operator's effective grants (operator
// grants UNION team grants) so the evaluator sees the full set in one query.
type GrantRepository interface {
	// Save upserts a ResourcePermission (idempotent on subject+scope+actions).
	Save(ctx context.Context, p *ResourcePermission) error
	// ListEffective returns all grants an operator can act with in an org:
	// operator-direct grants UNION grants on teams the operator belongs to.
	ListEffective(ctx context.Context, operatorID, orgID string) ([]*ResourcePermission, error)
	// ListByOrg returns all ResourcePermissions in an org (admin view).
	ListByOrg(ctx context.Context, orgID string) ([]*ResourcePermission, error)
	// Revoke deletes a grant by id, returning whether one was removed.
	Revoke(ctx context.Context, id string) (bool, error)
	// RevokeForSubject removes all grants for a subject (e.g. when a team is deleted).
	RevokeForSubject(ctx context.Context, subjectType SubjectType, subjectID string) error
}
