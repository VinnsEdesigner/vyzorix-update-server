package resolver

import (
	"context"
	"testing"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/graphql-go/graphql"
)

// newOrgScopeParams builds ResolveParams whose context carries the authenticated
// organization ID, mimicking what the middleware chain sets from the URL :org.
func newOrgScopeParams(ctxOrgID string, argOrgID string) graphql.ResolveParams {
	ctx := gqlcontext.WithOrganizationID(context.Background(), ctxOrgID)
	return graphql.ResolveParams{
		Context: ctx,
		Args: map[string]interface{}{
			"organizationId": argOrgID,
		},
	}
}

func TestResolveOrgID_MatchesArgument(t *testing.T) {
	r := &Resolver{}
	orgID, err := r.resolveOrgID(newOrgScopeParams("org-1", "org-1"))
	if err != nil {
		t.Fatalf("matching org should be allowed, got error: %v", err)
	}
	if orgID != "org-1" {
		t.Fatalf("expected org-1, got %q", orgID)
	}
}

func TestResolveOrgID_RejectsCrossOrgArgument(t *testing.T) {
	// This is the IDOR regression: a caller authenticated for org-1 passing
	// org-2 as the organizationId argument must be rejected, not served.
	r := &Resolver{}
	if _, err := r.resolveOrgID(newOrgScopeParams("org-1", "org-2")); err == nil {
		t.Fatal("expected cross-organization access to be rejected, but it was allowed")
	}
}

func TestResolveOrgID_AllowsOmittedArgument(t *testing.T) {
	// Omitting the argument is fine - the context org is authoritative.
	r := &Resolver{}
	orgID, err := r.resolveOrgID(newOrgScopeParams("org-1", ""))
	if err != nil {
		t.Fatalf("omitted argument should fall back to context org, got error: %v", err)
	}
	if orgID != "org-1" {
		t.Fatalf("expected org-1, got %q", orgID)
	}
}

func TestResolveOrgID_RejectsMissingContext(t *testing.T) {
	r := &Resolver{}
	params := graphql.ResolveParams{
		Context: context.Background(),
		Args:    map[string]interface{}{"organizationId": "org-1"},
	}
	if _, err := r.resolveOrgID(params); err == nil {
		t.Fatal("expected missing organization context to be rejected")
	}
}
