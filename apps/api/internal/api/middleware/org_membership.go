package middleware

import (
	"context"
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/gin-gonic/gin"
)

// ContextKeyMembership is the key used to store organization membership in gin context.
const ContextKeyMembership = "org_membership"

// OrganizationMembershipChecker defines the interface for checking organization membership.
type OrganizationMembershipChecker interface {
	GetMembership(ctx context.Context, operatorID, orgID string) (interface{}, error)
}

// OrganizationMembership middleware validates operator's membership in the organization.
type OrganizationMembership struct {
	membershipChecker OrganizationMembershipChecker
}

// NewOrganizationMembership creates a new OrganizationMembership middleware.
func NewOrganizationMembership(checker OrganizationMembershipChecker) *OrganizationMembership {
	return &OrganizationMembership{
		membershipChecker: checker,
	}
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

		// Get operator from context (set by auth middleware).
		operator := GetOperatorFromContext(c)
		if operator == nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "authentication required",
			})
			return
		}

		// If no membership checker is configured, deny access (fail secure).
		if m.membershipChecker == nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error":   "server_error",
				"message": "membership validation not configured",
			})
			return
		}

		// Check if operator is a member of the organization.
		membership, err := m.membershipChecker.GetMembership(c.Request.Context(), operator.ID, orgID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "not a member of this organization",
			})
			return
		}

		// Store actual membership object in context for downstream use.
		// GetMembership returns interface{}, so we need to assert the type.
		if member, ok := membership.(*organization.OrganizationMember); ok {
			c.Set(ContextKeyMembership, member)
		} else {
			// Fallback: store the raw value if it's not the expected type.
			c.Set(ContextKeyMembership, membership)
		}

		c.Next()
	}
}

// RequireOrganizationMembership is a convenience function for requiring organization membership.
// Note: This should only be used when the membership checker has been set via NewOrganizationMembership.
func RequireOrganizationMembership() gin.HandlerFunc {
	return NewOrganizationMembership(nil).Middleware()
}
