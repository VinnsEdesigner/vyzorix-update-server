// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/operator"
	"github.com/gin-gonic/gin"
)

// UpdatesAdminAuth provides admin authorization for updates endpoints.
// Per the spec: "Only admins can push updates".
type UpdatesAdminAuth struct{}

// NewUpdatesAdminAuth creates a new UpdatesAdminAuth middleware.
func NewUpdatesAdminAuth() *UpdatesAdminAuth {
	return &UpdatesAdminAuth{}
}

// RequireAdmin returns a Gin middleware that requires admin role.
func (m *UpdatesAdminAuth) RequireAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		if !op.IsAdmin() {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Admin privileges required for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin returns a Gin middleware that requires super_admin role.
func (m *UpdatesAdminAuth) RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		op := GetOperatorFromContext(c)
		if op == nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			c.Abort()
			return
		}

		if op.Role != operator.RoleSuperAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Super admin privileges required for this operation",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}
