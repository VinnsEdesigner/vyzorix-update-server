package api

import (
	"net/http"
	"strings"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
	"github.com/gin-gonic/gin"
)

// requireScope returns gin middleware that enforces a scoped permission: the
// caller must hold `action` on `scope` in the current organization, resolved
// through the scoped permission engine (role defaults unioned with custom
// grants). This mirrors the standard scoped-RBAC pattern `authorize(EvalPermission(action, scope))`
// at the route level. The scope is normally a resource-type wildcard
// ("devices:*"); per-resource narrowing happens in the handler (e.g. the
// command path's owner-or-group check).
func (s *Server) requireScope(action permission.Action, scope string) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Service accounts authenticate without an operator; evaluate scope
		// based on the token's own scope list instead.
		if c.GetString("auth_type") == "service_account" {
			scopes := c.GetStringSlice("service_account_scopes")
			if !serviceAccountGrants(scopes, action) {
				responses.RespondStructured(c, http.StatusForbidden, "insufficient permissions")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		op := middleware.GetOperatorFromContext(c)
		if op == nil {
			responses.RespondStructured(c, http.StatusUnauthorized, "authentication required")
			c.Abort()
			return
		}

		orgID := middleware.GetOrganizationID(c)
		if orgID == "" {
			responses.RespondStructured(c, http.StatusBadRequest, "organization context required")
			c.Abort()
			return
		}

		m := op.GetMembership(orgID)
		if m == nil || !m.IsActive() {
			responses.RespondStructured(c, http.StatusForbidden, "not a member of this organization")
			c.Abort()
			return
		}

		var grants []*permission.Grant
		if s.grantRepo != nil {
			grants, _ = s.grantRepo.ListEffective(c.Request.Context(), op.ID, orgID)
		}
		eval := permission.NewEvaluator(string(m.Role), grants)
		if !eval.Grants(action, scope) {
			responses.RespondStructured(c, http.StatusForbidden, "insufficient permissions")
			c.Abort()
			return
		}

		c.Next()
	}
}

// serviceAccountGrants maps token scopes (read/write/admin) to the required
// action class: read satisfies *.read, write adds *.write, admin adds every
// remaining scope prefix.
func serviceAccountGrants(scopes []string, action permission.Action) bool {
	highest := "read"
	for _, s := range scopes {
		switch s {
		case "admin":
			return true
		case "write":
			highest = "write"
		}
	}
	switch highest {
	case "write":
		return strings.HasSuffix(string(action), ".read") || strings.HasSuffix(string(action), ".write")
	default:
		return strings.HasSuffix(string(action), ".read")
	}
}
