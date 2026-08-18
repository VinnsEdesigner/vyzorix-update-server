package middleware

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/gin-gonic/gin"
)

type SuperAdminAuth struct{}

// NewSuperAdminAuth creates a new SuperAdminAuth middleware.
func NewSuperAdminAuth() *SuperAdminAuth {
	return &SuperAdminAuth{}
}

// Middleware returns the Gin middleware function for super admin authorization.
// This middleware checks if the operator is a super_admin in the organization context.
// IMPORTANT: Requires organization ID in context - no global fallback.
func (s *SuperAdminAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			responses.RespondStructuredAbort(c, http.StatusUnauthorized,

				"Authentication required",
			)
			return
		}

		// Require organization context.
		orgID := GetOrganizationID(c)
		if orgID == "" {
			responses.RespondStructuredAbort(c, http.StatusBadRequest,

				"Organization ID is required for super admin access",
			)
			return
		}

		// Check if operator is super_admin in this specific organization.
		if !op.IsSuperAdminIn(orgID) {
			responses.RespondStructuredAbort(c, http.StatusForbidden,

				"Super admin access required in this organization",
			)
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin is a convenience function that returns the super admin middleware.
// IMPORTANT: This now requires organization-scoped super_admin access (not global).
// The operator must be a super_admin in the current organization context.
func RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			responses.RespondStructuredAbort(c, http.StatusUnauthorized,

				"Authentication required",
			)
			return
		}

		// Require organization context for super admin access.
		orgID := GetOrganizationID(c)
		if orgID == "" {
			responses.RespondStructuredAbort(c, http.StatusBadRequest,

				"Organization ID is required for admin access",
			)
			return
		}

		// Check if operator is super_admin in this specific organization.
		if !op.IsSuperAdminIn(orgID) {
			responses.RespondStructuredAbort(c, http.StatusForbidden,

				"Super admin access required in this organization",
			)
			return
		}

		c.Next()
	}
}
