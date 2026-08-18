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
func (c *OrganizationContext) Middleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// Try to get organization ID from query parameter first.
		orgID := ctx.Query("organization_id")

		// Then check header.
		if orgID == "" {
			orgID = ctx.GetHeader("X-Organization-ID")
		}

		// Then check context (set by auth middleware).
		if orgID == "" {
			if storedOrgID, exists := ctx.Get(ContextKeyOrganizationID); exists {
				if id, ok := storedOrgID.(string); ok {
					orgID = id
				}
			}
		}

		// Fall back to session's SelectedOrganizationID if no explicit org ID.
		if orgID == "" {
			if sessVal, exists := ctx.Get("session"); exists {
				if sess, ok := sessVal.(*session.Session); ok {
					orgID = sess.SelectedOrganizationID
				}
			}
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
