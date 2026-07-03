package device

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterRoutes registers the device logs routes.
func (h *LogsHandler) RegisterRoutes(r *gin.RouterGroup, rateLimiter *middleware.DashboardRateLimiterMiddleware) {
	r.GET("/device/:imei/logs",
		rateLimiter.DeviceLogsLimit(),
		h.GetLogs)
}
