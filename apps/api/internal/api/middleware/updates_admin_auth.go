// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"

	gqlcontext "github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/graphql/context"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
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
		op, exists := gqlcontext.GetOperator(c.Request.Context())
		if !exists || op == nil {
			responses.RespondStructured(c, http.StatusUnauthorized,

				"Authentication required",
			)
			c.Abort()
			return
		}

		if !op.IsAdmin() {
			responses.RespondStructured(c, http.StatusForbidden,

				"Admin privileges required for this operation",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}

// RequireSuperAdmin returns a Gin middleware that requires super_admin role.
func (m *UpdatesAdminAuth) RequireSuperAdmin() gin.HandlerFunc {
	return func(c *gin.Context) {
		op, exists := gqlcontext.GetOperator(c.Request.Context())
		if !exists || op == nil {
			responses.RespondStructured(c, http.StatusUnauthorized,

				"Authentication required",
			)
			c.Abort()
			return
		}

		if !op.IsSuperAdmin() {
			responses.RespondStructured(c, http.StatusForbidden,

				"Super admin privileges required for this operation",
			)
			c.Abort()
			return
		}

		c.Next()
	}
}
