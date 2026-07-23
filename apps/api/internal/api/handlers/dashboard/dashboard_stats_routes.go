package dashboard

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the dashboard routes.
func (h *StatsHandler) RegisterRoutes(r *gin.RouterGroup, membershipChecker middleware.OrganizationMembershipChecker) {
	// Dashboard routes require organization context for multi-tenant isolation.
	dashboardGroup := r.Group("/dashboard")
	dashboardGroup.Use(middleware.NewOrganizationContext(nil).Middleware())
	dashboardGroup.Use(middleware.NewOrganizationMembership(membershipChecker).Middleware())
	dashboardGroup.GET("/stats", h.GetStats)
}
