// Package auth provides authentication utilities.
// Deprecated: Use github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security instead.
package auth

import (
"time"

"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/security/ratelimit"
)

// Re-export from new location for backward compatibility.
type (
RateLimitConfig    = ratelimit.Config
RateLimiter        = ratelimit.Limiter
MultiWindowLimiter = ratelimit.MultiWindowLimiter
)

var AuthRateLimiter = ratelimit.AuthLimiter

func NewRateLimiter(window time.Duration, maxRequests int) *ratelimit.Limiter {
return ratelimit.New(window, maxRequests)
}

func NewMultiWindowLimiter(limits map[string]struct {
Window time.Duration
Max    int
}) *ratelimit.MultiWindowLimiter {
return ratelimit.NewMultiWindow(limits)
}
