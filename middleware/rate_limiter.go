package middleware

import (
	"net/http"
	"sync"
	"time"
)

type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*bucket
	Capacity int
	Refill   time.Duration
}
type bucket struct {
	tokens int
	last   time.Time
}

func NewRateLimiter(capacity int, refill time.Duration) *RateLimiter {
	return &RateLimiter{buckets: map[string]*bucket{}, Capacity: capacity, Refill: refill}
}
func (l *RateLimiter) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.Allow(r.RemoteAddr) {
			http.Error(w, "rate limited", http.StatusTooManyRequests)
			return
		}
		next.ServeHTTP(w, r)
	})
}
func (l *RateLimiter) Allow(key string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	now := time.Now()
	b := l.buckets[key]
	if b == nil {
		b = &bucket{tokens: l.Capacity, last: now}
		l.buckets[key] = b
	}
	if elapsed := int(now.Sub(b.last) / l.Refill); elapsed > 0 {
		b.tokens += elapsed
		if b.tokens > l.Capacity {
			b.tokens = l.Capacity
		}
		b.last = now
	}
	if b.tokens <= 0 {
		return false
	}
	b.tokens--
	return true
}
