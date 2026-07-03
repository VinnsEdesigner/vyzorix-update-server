package command

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the command history routes.
func (h *HistoryHandler) RegisterRoutes(r *gin.RouterGroup, rateLimiter *middleware.DashboardRateLimiterMiddleware) {
	r.GET("/device/:imei/commands",
		rateLimiter.CommandHistoryLimit(),
		h.GetHistory)
}
