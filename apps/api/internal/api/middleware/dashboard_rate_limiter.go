// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// DashboardRateLimits defines rate limits for dashboard command endpoints per the spec.
type DashboardRateLimits struct {
	// GET /v1/device/:imei/commands: 60 requests per minute.
	CommandHistoryLimit int
	CommandHistoryRefill time.Duration

	// GET /v1/device/:imei/logs: 60 requests per minute.
	DeviceLogsLimit int
	DeviceLogsRefill time.Duration

	// GET /v1/device/:imei/metrics: 30 requests per minute.
	DeviceMetricsLimit int
	DeviceMetricsRefill time.Duration

	// GET /v1/device/:imei/metrics/export: 10 requests per minute.
	MetricsExportLimit int
	MetricsExportRefill time.Duration

	// POST /v1/device/:imei/command: 10 requests per minute.
	SendCommandLimit int
	SendCommandRefill time.Duration

	// Diagnostics endpoints: 30 requests per minute.
	DeviceInspectLimit int
	DeviceInspectRefill time.Duration
	DeviceTimelineLimit int
	DeviceTimelineRefill time.Duration
}

// DefaultDashboardRateLimits returns the default rate limits per the spec.
func DefaultDashboardRateLimits() *DashboardRateLimits {
	return &DashboardRateLimits{
		// GET /v1/device/:imei/commands: 60 per minute.
		CommandHistoryLimit:  60,
		CommandHistoryRefill: time.Minute,

		// GET /v1/device/:imei/logs: 60 per minute.
		DeviceLogsLimit:  60,
		DeviceLogsRefill: time.Minute,

		// GET /v1/device/:imei/metrics: 30 per minute.
		DeviceMetricsLimit:  30,
		DeviceMetricsRefill: time.Minute,

		// GET /v1/device/:imei/metrics/export: 10 per minute.
		MetricsExportLimit:  10,
		MetricsExportRefill: time.Minute,

		// POST /v1/device/:imei/command: 10 per minute.
		SendCommandLimit:  10,
		SendCommandRefill: time.Minute,

		// Diagnostics: 30 per minute.
		DeviceInspectLimit:   30,
		DeviceInspectRefill:  time.Minute,
		DeviceTimelineLimit:  30,
		DeviceTimelineRefill: time.Minute,
	}
}

// DashboardRateLimiterMiddleware creates rate limiters for dashboard endpoints.
type DashboardRateLimiterMiddleware struct {
	commandHistoryLimiter *RateLimiter
	deviceLogsLimiter     *RateLimiter
	deviceMetricsLimiter  *RateLimiter
	metricsExportLimiter  *RateLimiter
	sendCommandLimiter    *RateLimiter
	deviceInspectLimiter   *RateLimiter
	deviceTimelineLimiter *RateLimiter
}

// NewDashboardRateLimiterMiddleware creates a new dashboard rate limiter middleware.
func NewDashboardRateLimiterMiddleware(limits *DashboardRateLimits) *DashboardRateLimiterMiddleware {
	if limits == nil {
		limits = DefaultDashboardRateLimits()
	}

	return &DashboardRateLimiterMiddleware{
		commandHistoryLimiter: NewRateLimiter(limits.CommandHistoryLimit, limits.CommandHistoryRefill),
		deviceLogsLimiter:     NewRateLimiter(limits.DeviceLogsLimit, limits.DeviceLogsRefill),
		deviceMetricsLimiter:  NewRateLimiter(limits.DeviceMetricsLimit, limits.DeviceMetricsRefill),
		metricsExportLimiter:  NewRateLimiter(limits.MetricsExportLimit, limits.MetricsExportRefill),
		sendCommandLimiter:    NewRateLimiter(limits.SendCommandLimit, limits.SendCommandRefill),
		deviceInspectLimiter:  NewRateLimiter(limits.DeviceInspectLimit, limits.DeviceInspectRefill),
		deviceTimelineLimiter: NewRateLimiter(limits.DeviceTimelineLimit, limits.DeviceTimelineRefill),
	}
}

// CommandHistoryLimit returns the rate limiter middleware for GET /v1/device/:imei/commands.
func (m *DashboardRateLimiterMiddleware) CommandHistoryLimit() gin.HandlerFunc {
	return dashboardRateLimitMiddleware(m.commandHistoryLimiter)
}

// DeviceLogsLimit returns the rate limiter middleware for GET /v1/device/:imei/logs.
func (m *DashboardRateLimiterMiddleware) DeviceLogsLimit() gin.HandlerFunc {
	return dashboardRateLimitMiddleware(m.deviceLogsLimiter)
}

// DeviceMetricsLimit returns the rate limiter middleware for GET /v1/device/:imei/metrics.
func (m *DashboardRateLimiterMiddleware) DeviceMetricsLimit() gin.HandlerFunc {
	return dashboardRateLimitMiddleware(m.deviceMetricsLimiter)
}

// MetricsExportLimit returns the rate limiter middleware for GET /v1/device/:imei/metrics/export.
func (m *DashboardRateLimiterMiddleware) MetricsExportLimit() gin.HandlerFunc {
	return dashboardRateLimitMiddleware(m.metricsExportLimiter)
}

// SendCommandLimit returns the rate limiter middleware for POST /v1/device/:imei/command.
func (m *DashboardRateLimiterMiddleware) SendCommandLimit() gin.HandlerFunc {
	return dashboardRateLimitMiddleware(m.sendCommandLimiter)
}

// DeviceInspectLimit returns the rate limiter middleware for GET /v1/device/:imei/inspect.
func (m *DashboardRateLimiterMiddleware) DeviceInspectLimit() gin.HandlerFunc {
	return dashboardRateLimitMiddleware(m.deviceInspectLimiter)
}

// DeviceTimelineLimit returns the rate limiter middleware for GET /v1/device/:imei/timeline.
func (m *DashboardRateLimiterMiddleware) DeviceTimelineLimit() gin.HandlerFunc {
	return dashboardRateLimitMiddleware(m.deviceTimelineLimiter)
}

// Stop stops all rate limiter cleanup goroutines.
func (m *DashboardRateLimiterMiddleware) Stop() {
	m.commandHistoryLimiter.Stop()
	m.deviceLogsLimiter.Stop()
	m.deviceMetricsLimiter.Stop()
	m.metricsExportLimiter.Stop()
	m.sendCommandLimiter.Stop()
	m.deviceInspectLimiter.Stop()
	m.deviceTimelineLimiter.Stop()
}

// Stats returns statistics for all rate limiters.
func (m *DashboardRateLimiterMiddleware) Stats() map[string]RateLimiterStats {
	return map[string]RateLimiterStats{
		"command_history":  m.commandHistoryLimiter.Stats(),
		"device_logs":      m.deviceLogsLimiter.Stats(),
		"device_metrics":   m.deviceMetricsLimiter.Stats(),
		"metrics_export":   m.metricsExportLimiter.Stats(),
		"send_command":     m.sendCommandLimiter.Stats(),
		"device_inspect":    m.deviceInspectLimiter.Stats(),
		"device_timeline":   m.deviceTimelineLimiter.Stats(),
	}
}

// dashboardRateLimitMiddleware creates a Gin middleware for rate limiting.
func dashboardRateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use operator ID if available, otherwise use IP.
		key := c.ClientIP()
		if op, exists := c.Get("operator_id"); exists {
			if opID, ok := op.(string); ok && opID != "" {
				key = "op:" + opID
			}
		}

		// Also include device ID in key for device-specific endpoints.
		if deviceID := c.Param("id"); deviceID != "" {
			key = key + ":device:" + deviceID
		}

		// Include IMEI if present.
		if imei := c.Param("imei"); imei != "" {
			key = key + ":imei:" + imei
		}

		if !limiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "rate_limited",
				"message": "Too many requests. Please try again later.",
				"details": gin.H{
					"retry_after_seconds": int(limiter.Refill.Seconds()),
				},
			})
			c.Header("Retry-After", strconv.Itoa(int(limiter.Refill.Seconds())))
			c.Abort()
			return
		}

		c.Next()
	}
}
