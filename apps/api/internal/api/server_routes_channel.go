package api

import (
	"github.com/gin-gonic/gin"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"
)

// setupChannelRoutes registers org-scoped live channel subscription endpoints.
func (s *Server) setupChannelRoutes(r *gin.RouterGroup) {
	if s.channelHandler == nil {
		return
	}
	group := r.Group("/channels")
	group.Use(middleware.NewOrganizationContext(nil).Middleware())
	group.Use(middleware.NewOrganizationMembership(s.memberHandler.MembershipChecker()).Middleware())
	{
		group.GET("/status",
			s.requireScope(permission.ActionAlertRead, permission.WildcardScope(permission.ScopeOrg)),
			s.channelHandler.Status)
		group.POST("/subscribe",
			s.requireScope(permission.ActionAlertRead, permission.WildcardScope(permission.ScopeOrg)),
			s.channelHandler.Subscribe)
		group.POST("/unsubscribe",
			s.requireScope(permission.ActionAlertRead, permission.WildcardScope(permission.ScopeOrg)),
			s.channelHandler.Unsubscribe)
	}
}
