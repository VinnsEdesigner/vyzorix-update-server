package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// SuperAdminAuth provides super admin authorization middleware.
type SuperAdminAuth struct{}

// NewSuperAdminAuth creates a new SuperAdminAuth middleware.
func NewSuperAdminAuth() *SuperAdminAuth {
	return &SuperAdminAuth{}
}

// Middleware returns the Gin middleware function for super admin authorization.
func (s *SuperAdminAuth) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if user is authenticated via session
		operatorVal, exists := c.Get("operator_id")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Authentication required",
			})
			return
		}

		operatorID, ok := operatorVal.(string)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error":   "unauthorized",
				"message": "Invalid operator ID",
			})
			return
		}

		// Check if operator has super_admin role
		roleVal, exists := c.Get("operator_role")
		if !exists {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Super admin access required",
			})
			return
		}

		role, ok := roleVal.(string)
		if !ok || role != "super_admin" {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error":   "forbidden",
				"message": "Super admin access required",
			})
			return
		}

		// Store operator ID for audit logging
		c.Set("super_admin_operator_id", operatorID)
		c.Next()
	}
}

// RequireSuperAdmin is a convenience function that returns the super admin middleware.
func RequireSuperAdmin() gin.HandlerFunc {
	return NewSuperAdminAuth().Middleware()
}
