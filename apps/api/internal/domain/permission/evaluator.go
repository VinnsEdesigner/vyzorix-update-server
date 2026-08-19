package permission

// Evaluator resolves an operator's effective scoped permissions within an
// organization: the member's role defaults unioned with custom
// ResourcePermissions (operator-direct and team grants), with action sets
// expanded into base actions. Authorization layers evaluate (action, scope)
// against it instead of checking role tier.
type Evaluator struct {
	scopes ScopedPermissions
}

// UnionScopes merges the role's default scopes with custom grants and expands
// action sets into base actions.
func UnionScopes(role string, custom []*ResourcePermission) ScopedPermissions {
	scopes := DefaultScopesForRole(role)
	for _, g := range custom {
		for _, a := range g.Actions {
			scopes = append(scopes, ScopedPermission{Action: a, Scope: g.Scope})
		}
	}
	return ExpandActionSets(scopes)
}

// NewEvaluator builds an Evaluator from a role name and custom grants. The
// grants are unioned with role defaults and action sets are expanded.
func NewEvaluator(role string, custom []*ResourcePermission) *Evaluator {
	return &Evaluator{scopes: UnionScopes(role, custom)}
}

// NewEvaluatorWithScopes builds an Evaluator from a pre-assembled scope set.
// The caller is responsible for action-set expansion (used by tests that build
// scopes directly).
func NewEvaluatorWithScopes(scopes ScopedPermissions) *Evaluator {
	return &Evaluator{scopes: scopes}
}

// Grants reports whether the evaluator grants `action` on `scope`. Because
// action sets are expanded at construction, an aggregate grant (e.g.
// device.manage) satisfies a check for any of its base actions.
func (e *Evaluator) Grants(action Action, scope string) bool {
	return e.scopes.Grants(action, scope)
}
