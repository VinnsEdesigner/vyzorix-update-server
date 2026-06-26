package device

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterMetricsRoutes registers the device metrics routes.
func (h *MetricsHandler) RegisterMetricsRoutes(r *gin.RouterGroup, rateLimiter *middleware.DashboardRateLimiterMiddleware) {
	r.GET("/device/:id/metrics",
		rateLimiter.DeviceMetricsLimit(),
		h.GetMetrics)

	r.GET("/device/:id/metrics/export",
		rateLimiter.MetricsExportLimit(),
		h.ExportMetrics)
}
