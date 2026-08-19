package permission

import (
	"context"
	"strings"
	"sync"
)

// ScopeResolver expands a requested scope into the concrete grantable scopes
// that cover it. A resource type registers a resolver for its scope prefix;
// when a request checks "devices:imei:356938...", the devices resolver turns
// that into ["devices:imei:356938...", "devices:group:field-ops", "devices:*"]
// — the set of scopes whose grants would satisfy the request. This is the
// attribute-resolution layer: it maps a runtime resource identifier to the
// scope paths an operator might be granted on.
type ScopeResolver interface {
	// Resolve returns the concrete scopes that grant access to the requested
	// scope. The requested scope is always included as the first element so an
	// exact-match grant still works without resolver logic.
	Resolve(ctx context.Context, scope string) ([]string, error)
}

// ScopeResolverFunc adapts a function to the ScopeResolver interface.
type ScopeResolverFunc func(ctx context.Context, scope string) ([]string, error)

func (f ScopeResolverFunc) Resolve(ctx context.Context, scope string) ([]string, error) {
	return f(ctx, scope)
}

// resolverRegistry holds scope-prefix → resolver registrations. It is safe for
// concurrent read/write — resolvers register at startup, reads happen per
// request.
type resolverRegistry struct {
	resolvers map[string]ScopeResolver
	mu        sync.RWMutex
}

var globalResolvers = &resolverRegistry{resolvers: make(map[string]ScopeResolver)}

// RegisterScopeResolver registers a resolver for a scope prefix. When a
// requested scope starts with the prefix, the resolver expands it into the
// concrete grantable scopes. Registering a resolver for an existing prefix
// replaces it.
func RegisterScopeResolver(prefix string, resolver ScopeResolver) {
	globalResolvers.mu.Lock()
	defer globalResolvers.mu.Unlock()
	globalResolvers.resolvers[prefix] = resolver
}

// ResetScopeResolvers clears all registered resolvers (used in tests).
func ResetScopeResolvers() {
	globalResolvers.mu.Lock()
	defer globalResolvers.mu.Unlock()
	globalResolvers.resolvers = make(map[string]ScopeResolver)
}

// ResolveScope expands a requested scope into the concrete grantable scopes
// that cover it. If no resolver is registered for the scope's prefix, the
// requested scope is returned as-is (exact match + trailing-wildcard matching
// still applies).
func ResolveScope(ctx context.Context, scope string) []string {
	if scope == "" || !ValidScope(scope) {
		return nil
	}
	prefix := scopePrefix(scope)
	globalResolvers.mu.RLock()
	resolver, ok := globalResolvers.resolvers[prefix]
	globalResolvers.mu.RUnlock()
	if !ok || resolver == nil {
		return []string{scope}
	}
	expanded, err := resolver.Resolve(ctx, scope)
	if err != nil || len(expanded) == 0 {
		return []string{scope}
	}
	// Always include the original scope so exact-match grants still work.
	out := make([]string, 0, len(expanded)+1)
	out = append(out, scope)
	for _, s := range expanded {
		if s != scope && ValidScope(s) {
			out = append(out, s)
		}
	}
	return out
}

// scopePrefix returns the first segment of a colon-delimited scope.
func scopePrefix(scope string) string {
	if idx := strings.Index(scope, ":"); idx >= 0 {
		return scope[:idx]
	}
	return scope
}

// GrantsResolved reports whether the set grants `action` on `scope`, resolving
// the scope through registered resolvers first. This is the evaluation entry
// point that resolvers plug into: instead of checking only the literal scope,
// it checks every concrete scope the requested scope resolves to.
func (s ScopedPermissions) GrantsResolved(ctx context.Context, action Action, scope string) bool {
	for _, target := range ResolveScope(ctx, scope) {
		if s.Grants(action, target) {
			return true
		}
	}
	return false
}
