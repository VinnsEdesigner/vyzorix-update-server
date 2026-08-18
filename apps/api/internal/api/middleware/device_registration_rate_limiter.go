// Package middleware provides HTTP middleware.
package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/api/responses"
	"github.com/gin-gonic/gin"
)

// DeviceRegistrationRateLimits defines rate limits for device registration endpoints per the spec.
type DeviceRegistrationRateLimits struct {
	// GET /v1/device/inbox: 60 requests per minute.
	InboxListLimit  int
	InboxListRefill time.Duration

	// GET /v1/device/inbox/:imei: 60 requests per minute.
	InboxGetLimit  int
	InboxGetRefill time.Duration

	// POST /v1/device/inbox/:imei/ack: 10 requests per minute.
	InboxAckLimit  int
	InboxAckRefill time.Duration

	// GET /v1/devices: 60 requests per minute.
	DevicesListLimit  int
	DevicesListRefill time.Duration

	// GET /v1/devices/:imei: 60 requests per minute.
	DevicesGetLimit  int
	DevicesGetRefill time.Duration

	// DELETE /v1/devices/:imei: 10 requests per minute.
	DevicesDeleteLimit  int
	DevicesDeleteRefill time.Duration
}

// DefaultDeviceRegistrationRateLimits returns the default rate limits per the spec.
func DefaultDeviceRegistrationRateLimits() *DeviceRegistrationRateLimits {
	return &DeviceRegistrationRateLimits{
		// GET /v1/device/inbox: 60 per minute.
		InboxListLimit:  60,
		InboxListRefill: time.Minute,

		// GET /v1/device/inbox/:imei: 60 per minute.
		InboxGetLimit:  60,
		InboxGetRefill: time.Minute,

		// POST /v1/device/inbox/:imei/ack: 10 per minute.
		InboxAckLimit:  10,
		InboxAckRefill: time.Minute,

		// GET /v1/devices: 60 per minute.
		DevicesListLimit:  60,
		DevicesListRefill: time.Minute,

		// GET /v1/devices/:imei: 60 per minute.
		DevicesGetLimit:  60,
		DevicesGetRefill: time.Minute,

		// DELETE /v1/devices/:imei: 10 per minute.
		DevicesDeleteLimit:  10,
		DevicesDeleteRefill: time.Minute,
	}
}

// DeviceRegistrationRateLimiterMiddleware creates rate limiters for device registration endpoints.
type DeviceRegistrationRateLimiterMiddleware struct {
	inboxListLimiter     *RateLimiter
	inboxGetLimiter      *RateLimiter
	inboxAckLimiter      *RateLimiter
	devicesListLimiter   *RateLimiter
	devicesGetLimiter    *RateLimiter
	devicesDeleteLimiter *RateLimiter
}

// NewDeviceRegistrationRateLimiterMiddleware creates a new device registration rate limiter middleware.
func NewDeviceRegistrationRateLimiterMiddleware(limits *DeviceRegistrationRateLimits) *DeviceRegistrationRateLimiterMiddleware {
	if limits == nil {
		limits = DefaultDeviceRegistrationRateLimits()
	}

	return &DeviceRegistrationRateLimiterMiddleware{
		inboxListLimiter:     NewRateLimiter(limits.InboxListLimit, limits.InboxListRefill),
		inboxGetLimiter:      NewRateLimiter(limits.InboxGetLimit, limits.InboxGetRefill),
		inboxAckLimiter:      NewRateLimiter(limits.InboxAckLimit, limits.InboxAckRefill),
		devicesListLimiter:   NewRateLimiter(limits.DevicesListLimit, limits.DevicesListRefill),
		devicesGetLimiter:    NewRateLimiter(limits.DevicesGetLimit, limits.DevicesGetRefill),
		devicesDeleteLimiter: NewRateLimiter(limits.DevicesDeleteLimit, limits.DevicesDeleteRefill),
	}
}

// InboxListLimit returns the rate limiter middleware for GET /v1/device/inbox.
func (m *DeviceRegistrationRateLimiterMiddleware) InboxListLimit() gin.HandlerFunc {
	return deviceRegRateLimitMiddleware(m.inboxListLimiter)
}

// InboxGetLimit returns the rate limiter middleware for GET /v1/device/inbox/:imei.
func (m *DeviceRegistrationRateLimiterMiddleware) InboxGetLimit() gin.HandlerFunc {
	return deviceRegRateLimitMiddleware(m.inboxGetLimiter)
}

// InboxAckLimit returns the rate limiter middleware for POST /v1/device/inbox/:imei/ack.
func (m *DeviceRegistrationRateLimiterMiddleware) InboxAckLimit() gin.HandlerFunc {
	return deviceRegRateLimitMiddleware(m.inboxAckLimiter)
}

// DevicesListLimit returns the rate limiter middleware for GET /v1/devices.
func (m *DeviceRegistrationRateLimiterMiddleware) DevicesListLimit() gin.HandlerFunc {
	return deviceRegRateLimitMiddleware(m.devicesListLimiter)
}

// DevicesGetLimit returns the rate limiter middleware for GET /v1/devices/:imei.
func (m *DeviceRegistrationRateLimiterMiddleware) DevicesGetLimit() gin.HandlerFunc {
	return deviceRegRateLimitMiddleware(m.devicesGetLimiter)
}

// DevicesDeleteLimit returns the rate limiter middleware for DELETE /v1/devices/:imei.
func (m *DeviceRegistrationRateLimiterMiddleware) DevicesDeleteLimit() gin.HandlerFunc {
	return deviceRegRateLimitMiddleware(m.devicesDeleteLimiter)
}

// Stop stops all rate limiter cleanup goroutines.
func (m *DeviceRegistrationRateLimiterMiddleware) Stop() {
	m.inboxListLimiter.Stop()
	m.inboxGetLimiter.Stop()
	m.inboxAckLimiter.Stop()
	m.devicesListLimiter.Stop()
	m.devicesGetLimiter.Stop()
	m.devicesDeleteLimiter.Stop()
}

// Stats returns statistics for all rate limiters.
func (m *DeviceRegistrationRateLimiterMiddleware) Stats() map[string]RateLimiterStats {
	return map[string]RateLimiterStats{
		"inbox_list":     m.inboxListLimiter.Stats(),
		"inbox_get":      m.inboxGetLimiter.Stats(),
		"inbox_ack":      m.inboxAckLimiter.Stats(),
		"devices_list":   m.devicesListLimiter.Stats(),
		"devices_get":    m.devicesGetLimiter.Stats(),
		"devices_delete": m.devicesDeleteLimiter.Stats(),
	}
}

// deviceRegRateLimitMiddleware creates a Gin middleware for rate limiting.
func deviceRegRateLimitMiddleware(limiter *RateLimiter) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use operator ID if available, otherwise use IP.
		key := c.ClientIP()
		if op, exists := c.Get("operator_id"); exists {
			if opID, ok := op.(string); ok && opID != "" {
				key = "op:" + opID
			}
		}

		// Include IMEI in key for device-specific endpoints.
		if imei := c.Param("imei"); imei != "" {
			key = key + ":imei:" + imei
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
