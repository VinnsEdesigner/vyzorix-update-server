package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// ContextKeyMembership is the key used to store organization membership in gin context.
const ContextKeyMembership = "org_membership"

// OrganizationMembership middleware validates operator's membership in the organization.
// Note: This is a simplified version that works with global roles until org-scoped memberships are implemented.
type OrganizationMembership struct{}

// NewOrganizationMembership creates a new OrganizationMembership middleware.
func NewOrganizationMembership(_ interface{}) *OrganizationMembership {
	return &OrganizationMembership{}
}

// Middleware returns the gin middleware handler.
func (m *OrganizationMembership) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		orgID := GetOrganizationID(c)
		if orgID == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"error":   "bad_request",
				"message": "organization context required",
			})
			return
		}

		// Get operator from context (set by auth middleware)
		operator := GetOperatorFromContext(c)
		if operator == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "authentication required",
			})
			return
		}

		// TODO: Implement org-scoped membership check
		// For now, any authenticated operator with a valid orgID can access

		c.Next()
	}
}

// RequireOrganizationMembership is a convenience function for requiring organization membership.
func RequireOrganizationMembership() gin.HandlerFunc {
	return NewOrganizationMembership(nil).Middleware()
}
