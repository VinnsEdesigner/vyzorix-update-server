// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/gin-gonic/gin"
)

// SettingsRateLimits defines rate limits for settings endpoints per the spec.
type SettingsRateLimits struct {
	// GET /v1/auth/me/settings: 120 requests per minute.
	SettingsGetLimit  int
	SettingsGetRefill time.Duration

	// PATCH /v1/auth/me/settings: 30 requests per minute.
	SettingsUpdateLimit  int
	SettingsUpdateRefill time.Duration

	// GET /v1/auth/me/thresholds: 60 requests per minute.
	ThresholdsGetLimit  int
	ThresholdsGetRefill time.Duration

	// PATCH /v1/auth/me/thresholds: 30 requests per minute.
	ThresholdsUpdateLimit  int
	ThresholdsUpdateRefill time.Duration

	// GET /v1/auth/me/notifications: 60 requests per minute.
	NotificationsGetLimit  int
	NotificationsGetRefill time.Duration

	// PATCH /v1/auth/me/notifications: 30 requests per minute.
	NotificationsUpdateLimit  int
	NotificationsUpdateRefill time.Duration

	// POST /v1/auth/me/notifications/webhook/test: 10 requests per minute.
	WebhookTestLimit  int
	WebhookTestRefill time.Duration

	// POST /v1/auth/me/notifications/webhook/rotate: 5 requests per minute.
	WebhookRotateLimit  int
	WebhookRotateRefill time.Duration
}

// DefaultSettingsRateLimits returns the default rate limits.
func DefaultSettingsRateLimits() *SettingsRateLimits {
	return &SettingsRateLimits{
		// GET /v1/auth/me/settings: 120 per minute.
		SettingsGetLimit:  120,
		SettingsGetRefill: time.Minute,

		// PATCH /v1/auth/me/settings: 30 per minute.
		SettingsUpdateLimit:  30,
		SettingsUpdateRefill: time.Minute,

		// GET /v1/auth/me/thresholds: 60 per minute.
		ThresholdsGetLimit:  60,
		ThresholdsGetRefill: time.Minute,

		// PATCH /v1/auth/me/thresholds: 30 per minute.
		ThresholdsUpdateLimit:  30,
		ThresholdsUpdateRefill: time.Minute,

		// GET /v1/auth/me/notifications: 60 per minute.
		NotificationsGetLimit:  60,
		NotificationsGetRefill: time.Minute,

		// PATCH /v1/auth/me/notifications: 30 per minute.
		NotificationsUpdateLimit:  30,
		NotificationsUpdateRefill: time.Minute,

		// POST /v1/auth/me/notifications/webhook/test: 10 per minute.
		WebhookTestLimit:  10,
		WebhookTestRefill: time.Minute,

		// POST /v1/auth/me/notifications/webhook/rotate: 5 per minute.
		WebhookRotateLimit:  5,
		WebhookRotateRefill: time.Minute,
	}
}

// SettingsRateLimiterMiddleware creates rate limiters for settings endpoints.
type SettingsRateLimiterMiddleware struct {
	settingsGetLimiter         *RateLimiter
	settingsUpdateLimiter      *RateLimiter
	thresholdsGetLimiter       *RateLimiter
	thresholdsUpdateLimiter    *RateLimiter
	notificationsGetLimiter    *RateLimiter
	notificationsUpdateLimiter *RateLimiter
	webhookTestLimiter         *RateLimiter
	webhookRotateLimiter       *RateLimiter
}

// NewSettingsRateLimiterMiddleware creates a new settings rate limiter middleware.
func NewSettingsRateLimiterMiddleware(limits *SettingsRateLimits) *SettingsRateLimiterMiddleware {
	if limits == nil {
		limits = DefaultSettingsRateLimits()
	}

	return &SettingsRateLimiterMiddleware{
		settingsGetLimiter:         NewRateLimiter(limits.SettingsGetLimit, limits.SettingsGetRefill),
		settingsUpdateLimiter:      NewRateLimiter(limits.SettingsUpdateLimit, limits.SettingsUpdateRefill),
		thresholdsGetLimiter:       NewRateLimiter(limits.ThresholdsGetLimit, limits.ThresholdsGetRefill),
		thresholdsUpdateLimiter:    NewRateLimiter(limits.ThresholdsUpdateLimit, limits.ThresholdsUpdateRefill),
		notificationsGetLimiter:    NewRateLimiter(limits.NotificationsGetLimit, limits.NotificationsGetRefill),
		notificationsUpdateLimiter: NewRateLimiter(limits.NotificationsUpdateLimit, limits.NotificationsUpdateRefill),
		webhookTestLimiter:         NewRateLimiter(limits.WebhookTestLimit, limits.WebhookTestRefill),
		webhookRotateLimiter:       NewRateLimiter(limits.WebhookRotateLimit, limits.WebhookRotateRefill),
	}
}

// SettingsGetLimit returns the rate limiter middleware for GET /v1/auth/me/settings.
func (m *SettingsRateLimiterMiddleware) SettingsGetLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.settingsGetLimiter)
}

// SettingsUpdateLimit returns the rate limiter middleware for PATCH /v1/auth/me/settings.
func (m *SettingsRateLimiterMiddleware) SettingsUpdateLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.settingsUpdateLimiter)
}

// ThresholdsGetLimit returns the rate limiter middleware for GET /v1/auth/me/thresholds.
func (m *SettingsRateLimiterMiddleware) ThresholdsGetLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.thresholdsGetLimiter)
}

// ThresholdsUpdateLimit returns the rate limiter middleware for PATCH /v1/auth/me/thresholds.
func (m *SettingsRateLimiterMiddleware) ThresholdsUpdateLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.thresholdsUpdateLimiter)
}

// NotificationsGetLimit returns the rate limiter middleware for GET /v1/auth/me/notifications.
func (m *SettingsRateLimiterMiddleware) NotificationsGetLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.notificationsGetLimiter)
}

// NotificationsUpdateLimit returns the rate limiter middleware for PATCH /v1/auth/me/notifications.
func (m *SettingsRateLimiterMiddleware) NotificationsUpdateLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.notificationsUpdateLimiter)
}

// WebhookTestLimit returns the rate limiter middleware for POST /v1/auth/me/notifications/webhook/test.
func (m *SettingsRateLimiterMiddleware) WebhookTestLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.webhookTestLimiter)
}

// WebhookRotateLimit returns the rate limiter middleware for POST /v1/auth/me/notifications/webhook/rotate.
func (m *SettingsRateLimiterMiddleware) WebhookRotateLimit() gin.HandlerFunc {
	return settingsRateLimitMiddleware(m.webhookRotateLimiter)
}

// Stop stops all rate limiter cleanup goroutines.
func (m *SettingsRateLimiterMiddleware) Stop() {
	m.settingsGetLimiter.Stop()
	m.settingsUpdateLimiter.Stop()
	m.thresholdsGetLimiter.Stop()
	m.thresholdsUpdateLimiter.Stop()
	m.notificationsGetLimiter.Stop()
	m.notificationsUpdateLimiter.Stop()
	m.webhookTestLimiter.Stop()
	m.webhookRotateLimiter.Stop()
}

// Stats returns statistics for all rate limiters.
func (m *SettingsRateLimiterMiddleware) Stats() map[string]RateLimiterStats {
	return map[string]RateLimiterStats{
		"settings_get":         m.settingsGetLimiter.Stats(),
		"settings_update":      m.settingsUpdateLimiter.Stats(),
		"thresholds_get":       m.thresholdsGetLimiter.Stats(),
		"thresholds_update":    m.thresholdsUpdateLimiter.Stats(),
		"notifications_get":    m.notificationsGetLimiter.Stats(),
		"notifications_update": m.notificationsUpdateLimiter.Stats(),
		"webhook_test":         m.webhookTestLimiter.Stats(),
		"webhook_rotate":       m.webhookRotateLimiter.Stats(),
	}
}

// settingsRateLimitMiddleware creates a Gin middleware for rate limiting.
func settingsRateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use operator ID if available, otherwise use IP.
		key := c.ClientIP()
		if op, exists := c.Get("operator_id"); exists {
			if opID, ok := op.(string); ok && opID != "" {
				key = "op:" + opID
			}
		}

		if !limiter.Allow(key) {
			responses.RespondStructured(c, http.StatusTooManyRequests,

				"Too many requests. Please try again later.",
			)
			c.Header("Retry-After", strconv.Itoa(int(limiter.Refill.Seconds())))
			c.Abort()
			return
		}

		c.Next()
	}
}
