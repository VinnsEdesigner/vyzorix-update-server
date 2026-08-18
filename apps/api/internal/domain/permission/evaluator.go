package permission

// Evaluator resolves an operator's effective scoped permissions within an
// organization: the member's role defaults unioned with any custom scoped
// grants. Authorization layers evaluate (action, scope) against it instead of
// checking role tier, enabling per-resource scoping.
type Evaluator struct {
	scopes ScopedPermissions
}

// UnionScopes merges the role's default scopes with custom grants as plain
// ScopedPermission tuples.
func UnionScopes(role string, custom []*Grant) ScopedPermissions {
	scopes := DefaultScopesForRole(role)
	for _, g := range custom {
		scopes = append(scopes, ScopedPermission{Action: g.Action, Scope: g.Scope})
	}
	return scopes
}

// NewEvaluator builds an Evaluator from a role name and custom grants.
func NewEvaluator(role string, custom []*Grant) *Evaluator {
	return &Evaluator{scopes: UnionScopes(role, custom)}
}

// Grants reports whether the evaluator grants `action` on `scope`.
func (e *Evaluator) Grants(action Action, scope string) bool {
	return e.scopes.Grants(action, scope)
}
