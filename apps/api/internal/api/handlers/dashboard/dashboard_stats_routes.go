package dashboard

import (
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the dashboard routes.
func (h *StatsHandler) RegisterRoutes(r *gin.RouterGroup) {
	r.GET("/dashboard/stats", h.GetStats)
}
