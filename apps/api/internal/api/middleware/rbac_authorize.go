package middleware

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/gin-gonic/gin"
)

// RequireOrgRole returns middleware that requires a minimum role level in the organization.
func RequireOrgRole(minRole organization.OrganizationRole) gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			responses.RespondStructured(c, http.StatusUnauthorized,

				"authentication required",
			)
			c.Abort()
			return
		}

		membership := op.GetMembership(c.GetString("organizationId"))
		if membership == nil || membership.Role.Level() < minRole.Level() {
			responses.RespondStructured(c, http.StatusForbidden,

				"insufficient organization role",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequirePermission returns middleware that requires a specific permission.
func RequirePermission(perm operator.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			responses.RespondStructured(c, http.StatusUnauthorized,

				"authentication required",
			)
			c.Abort()
			return
		}

		if !op.HasPermission(perm) {
			responses.RespondStructured(c, http.StatusForbidden,

				"insufficient permissions",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAnyPermission returns middleware that checks for any of the given permissions.
func RequireAnyPermission(perms ...operator.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			responses.RespondStructured(c, http.StatusUnauthorized,

				"authentication required",
			)
			c.Abort()
			return
		}

		if !op.HasAnyPermission(perms...) {
			responses.RespondStructured(c, http.StatusForbidden,

				"insufficient permissions",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireAllPermissions returns middleware that checks for all of the given permissions.
func RequireAllPermissions(perms ...operator.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			responses.RespondStructured(c, http.StatusUnauthorized,

				"authentication required",
			)
			c.Abort()
			return
		}

		if !op.HasAllPermissions(perms...) {
			responses.RespondStructured(c, http.StatusForbidden,

				"insufficient permissions",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}
