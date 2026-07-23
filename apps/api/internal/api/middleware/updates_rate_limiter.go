// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// UpdatesRateLimits defines rate limits for updates endpoints per the spec.
type UpdatesRateLimits struct {
	// Status endpoint: 60 requests per minute.
	StatusLimit int
	StatusRefill time.Duration

	// Versions endpoint: 30 requests per minute.
	VersionsLimit int
	VersionsRefill time.Duration

	// Changelog endpoint: 30 requests per minute.
	ChangelogLimit int
	ChangelogRefill time.Duration

	// Push endpoint: 10 requests per minute.
	PushLimit int
	PushRefill time.Duration

	// History endpoint: 30 requests per minute.
	HistoryLimit int
	HistoryRefill time.Duration

	// Cancel endpoint: 10 requests per minute.
	CancelLimit int
	CancelRefill time.Duration

	// Sync endpoint: 5 requests per hour.
	SyncLimit int
	SyncRefill time.Duration
}

// DefaultUpdatesRateLimits returns the default rate limits per the spec.
func DefaultUpdatesRateLimits() *UpdatesRateLimits {
	return &UpdatesRateLimits{
		// GET /v1/updates/status: 60 per minute.
		StatusLimit:  60,
		StatusRefill: time.Minute,

		// GET /v1/updates/versions: 30 per minute.
		VersionsLimit:  30,
		VersionsRefill: time.Minute,

		// GET /v1/updates/changelog: 30 per minute.
		ChangelogLimit:  30,
		ChangelogRefill: time.Minute,

		// POST /v1/updates/push: 10 per minute.
		PushLimit:  10,
		PushRefill: time.Minute,

		// GET /v1/updates/history: 30 per minute.
		HistoryLimit:  30,
		HistoryRefill: time.Minute,

		// POST /v1/updates/history/:id/cancel: 10 per minute.
		CancelLimit:  10,
		CancelRefill: time.Minute,

		// POST /v1/updates/sync: 5 per hour.
		SyncLimit:  5,
		SyncRefill: time.Hour,
	}
}

// UpdatesRateLimiterMiddleware creates rate limiters for updates endpoints.
type UpdatesRateLimiterMiddleware struct {
	statusLimiter    *RateLimiter
	versionsLimiter  *RateLimiter
	changelogLimiter *RateLimiter
	pushLimiter      *RateLimiter
	historyLimiter   *RateLimiter
	cancelLimiter    *RateLimiter
	syncLimiter      *RateLimiter
}

// NewUpdatesRateLimiterMiddleware creates a new updates rate limiter middleware.
func NewUpdatesRateLimiterMiddleware(limits *UpdatesRateLimits) *UpdatesRateLimiterMiddleware {
	if limits == nil {
		limits = DefaultUpdatesRateLimits()
	}

	return &UpdatesRateLimiterMiddleware{
		statusLimiter:    NewRateLimiter(limits.StatusLimit, limits.StatusRefill),
		versionsLimiter:  NewRateLimiter(limits.VersionsLimit, limits.VersionsRefill),
		changelogLimiter: NewRateLimiter(limits.ChangelogLimit, limits.ChangelogRefill),
		pushLimiter:      NewRateLimiter(limits.PushLimit, limits.PushRefill),
		historyLimiter:   NewRateLimiter(limits.HistoryLimit, limits.HistoryRefill),
		cancelLimiter:    NewRateLimiter(limits.CancelLimit, limits.CancelRefill),
		syncLimiter:      NewRateLimiter(limits.SyncLimit, limits.SyncRefill),
	}
}

// StatusLimit returns the rate limiter middleware for GET /status.
func (m *UpdatesRateLimiterMiddleware) StatusLimit() gin.HandlerFunc {
	return rateLimitMiddleware(m.statusLimiter)
}

// VersionsLimit returns the rate limiter middleware for GET /versions.
func (m *UpdatesRateLimiterMiddleware) VersionsLimit() gin.HandlerFunc {
	return rateLimitMiddleware(m.versionsLimiter)
}

// ChangelogLimit returns the rate limiter middleware for GET /changelog.
func (m *UpdatesRateLimiterMiddleware) ChangelogLimit() gin.HandlerFunc {
	return rateLimitMiddleware(m.changelogLimiter)
}

// PushLimit returns the rate limiter middleware for POST /push.
func (m *UpdatesRateLimiterMiddleware) PushLimit() gin.HandlerFunc {
	return rateLimitMiddleware(m.pushLimiter)
}

// HistoryLimit returns the rate limiter middleware for GET /history.
func (m *UpdatesRateLimiterMiddleware) HistoryLimit() gin.HandlerFunc {
	return rateLimitMiddleware(m.historyLimiter)
}

// CancelLimit returns the rate limiter middleware for POST /history/:id/cancel.
func (m *UpdatesRateLimiterMiddleware) CancelLimit() gin.HandlerFunc {
	return rateLimitMiddleware(m.cancelLimiter)
}

// SyncLimit returns the rate limiter middleware for POST /sync.
func (m *UpdatesRateLimiterMiddleware) SyncLimit() gin.HandlerFunc {
	return rateLimitMiddleware(m.syncLimiter)
}

// Stop stops all rate limiter cleanup goroutines.
func (m *UpdatesRateLimiterMiddleware) Stop() {
	m.statusLimiter.Stop()
	m.versionsLimiter.Stop()
	m.changelogLimiter.Stop()
	m.pushLimiter.Stop()
	m.historyLimiter.Stop()
	m.cancelLimiter.Stop()
	m.syncLimiter.Stop()
}

// Stats returns statistics for all rate limiters.
func (m *UpdatesRateLimiterMiddleware) Stats() map[string]RateLimiterStats {
	return map[string]RateLimiterStats{
		"status":    m.statusLimiter.Stats(),
		"versions":  m.versionsLimiter.Stats(),
		"changelog": m.changelogLimiter.Stats(),
		"push":      m.pushLimiter.Stats(),
		"history":   m.historyLimiter.Stats(),
		"cancel":    m.cancelLimiter.Stats(),
		"sync":      m.syncLimiter.Stats(),
	}
}

// rateLimitMiddleware creates a Gin middleware for rate limiting.
func rateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use operator ID if available, otherwise use IP.
		key := c.ClientIP()
		if op, exists := c.Get("operator_id"); exists {
			if opID, ok := op.(string); ok && opID != "" {
				key = "op:" + opID
			}
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
