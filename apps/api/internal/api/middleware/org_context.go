package middleware

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/session"
	"github.com/gin-gonic/gin"
)

// ContextKeyOrganizationID is the key used to store organization ID in gin context.
const ContextKeyOrganizationID = "organization_id"

// OrganizationContext middleware extracts and validates organization ID from request.
type OrganizationContext struct {
	// SkipIfMissing allows requests without organization context (for optional org scoping).
	SkipIfMissing bool
}

// NewOrganizationContext creates a new OrganizationContext middleware.
func NewOrganizationContext(skipIfMissing *bool) *OrganizationContext {
	c := &OrganizationContext{}
	if skipIfMissing != nil {
		c.SkipIfMissing = *skipIfMissing
	}
	return c
}

// Middleware returns the gin middleware handler.
//
// The organization is resolved from server-side credential state, never blindly
// trusted from the client:
//   - API-key auth: the org bound to the key at creation (set by the tenant API
//     key middleware as ContextKeyOrganizationID with org_source=api_key).
//   - Session auth: the session's SelectedOrganizationID.
//
// A client may still send ?organization_id= or X-Organization-ID. If it
// disagrees with the server-resolved org, the request is rejected (400) rather
// than silently honored — this is how the server, not the client, owns org
// state: the server forces a validated org switch instead of trusting a
// per-request header).
func (c *OrganizationContext) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		clientOrg := ctx.Query("organization_id")
		if clientOrg == "" {
			clientOrg = ctx.GetHeader("X-Organization-ID")
		}

		// Resolve the authoritative org from the credential.
		serverOrg := ""
		if storedOrgID, exists := ctx.Get(ContextKeyOrganizationID); exists {
			if id, ok := storedOrgID.(string); ok {
				serverOrg = id
			}
		}
		if serverOrg == "" {
			if sessVal, exists := ctx.Get("session"); exists {
				if sess, ok := sessVal.(*session.Session); ok {
					serverOrg = sess.SelectedOrganizationID
				}
			}
		}

		// A client-supplied org that disagrees with the resolved org is rejected.
		if clientOrg != "" && serverOrg != "" && clientOrg != serverOrg {
			responses.RespondStructuredAbort(ctx, 400, "organization does not match the authenticated credential; select the organization first")
			return
		}

		orgID := serverOrg
		if orgID == "" {
			orgID = clientOrg
		}

		if orgID == "" && !c.SkipIfMissing {
			responses.RespondStructuredAbort(ctx, 400,

				"organization context required",
			)
			return
		}

		ctx.Set(ContextKeyOrganizationID, orgID)
		ctx.Next()
	}
}

// GetOrganizationID retrieves the organization ID from the gin context.
func GetOrganizationID(c *gin.Context) string {
	orgID, exists := c.Get(ContextKeyOrganizationID)
	if !exists {
		return ""
	}
	if id, ok := orgID.(string); ok {
		return id
	}
	return ""
}

// GetMembership retrieves the organization membership from the gin context.
func GetMembership(c *gin.Context) *organization.OrganizationMember {
	membership, exists := c.Get(ContextKeyMembership)
	if !exists {
		return nil
	}
	if m, ok := membership.(*organization.OrganizationMember); ok {
		return m
	}
	return nil
}
