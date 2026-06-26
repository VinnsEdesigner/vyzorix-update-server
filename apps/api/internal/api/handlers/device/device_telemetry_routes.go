package device

import (
	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/middleware"
	"github.com/gin-gonic/gin"
)

// RegisterTelemetryRoutes registers the device telemetry routes.
func (h *TelemetryHandler) RegisterTelemetryRoutes(r *gin.RouterGroup, rateLimiter *middleware.DashboardRateLimiterMiddleware) {
	r.GET("/device/:id/telemetry",
		rateLimiter.DeviceMetricsLimit(),
		h.GetTelemetry)
}
