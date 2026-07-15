package middleware

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/organization"
	"github.com/gin-gonic/gin"
)

// =============================================================================
// Organization-Scoped RBAC (No Global Fallback)
// =============================================================================

// RequireOrgRole returns middleware that requires a minimum role level in the organization.
func RequireOrgRole(minRole organization.OrganizationRole) gin.HandlerFunc {
func RequirePermission(perm operator.Permission) gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		if !op.HasPermission(perm) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		if !op.HasAnyPermission(perms...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
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
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "authentication required",
			})
			c.Abort()
			return
		}

		if !op.HasAllPermissions(perms...) {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "insufficient permissions",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
