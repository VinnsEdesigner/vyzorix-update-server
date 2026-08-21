// Package cache provides an in-memory TTL cache with per-section metrics.
// Sections are named scopes (e.g. "search", "dashboard") that share expiry
// but have their own hit/miss counters.
package cache

import (
	"sync"
	"time"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/infrastructure/metrics"
)

// Cache is a thread-safe in-memory TTL cache with per-section metrics.
type Cache struct {
	items      map[string]*item
	metrics    *Metrics
	defaultTTL time.Duration
	mu         sync.RWMutex
}

// item is one cached entry with expiry.
type item struct {
	value     interface{}
	expiresAt time.Time
}

// Metrics tracks cache hit/miss counts per section.
type Metrics struct {
	counts map[string]int64
	mu     sync.Mutex
}

// Emit exports cached counters to the global metrics sink.
func (m *Metrics) Emit(section, outcome string) {
	metrics.Get().RecordCacheOperation(section, outcome)
}

// New creates a cache with the default TTL (seconds). Entries older than the
// TTL are evicted on read (lazy expiry).
func New(defaultTTL time.Duration) *Cache {
	return &Cache{
		items:      make(map[string]*item),
		defaultTTL: defaultTTL,
		metrics:    newMetrics(),
	}
}

// Get returns the value for the key if present and not expired. It returns
// (value, true) on hit and (nil, false) on miss.
func (c *Cache) Get(key string) (interface{}, bool) {
	c.mu.RLock()
	item, ok := c.items[key]
	c.mu.RUnlock()
	if !ok {
		return nil, false
	}
	if time.Now().After(item.expiresAt) {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
		return nil, false
	}
	return item.value, true
}

// Set stores a value with the default TTL.
func (c *Cache) Set(key string, value interface{}) {
	c.SetTTL(key, value, c.defaultTTL)
}

// SetTTL stores a value with a custom TTL.
func (c *Cache) SetTTL(key string, value interface{}, ttl time.Duration) {
	c.mu.Lock()
	c.items[key] = &item{
		value:     value,
		expiresAt: time.Now().Add(ttl),
	}
	c.mu.Unlock()
}

// Delete removes a key.
func (c *Cache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// Flush clears the entire cache.
func (c *Cache) Flush() {
	c.mu.Lock()
	c.items = make(map[string]*item)
	c.mu.Unlock()
}

// Len returns the number of cached entries (including expired-but-not-evicted).
func (c *Cache) Len() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.items)
}

// Section returns a per-section scoped view. All section accesses route through
// the parent cache but carry their own metrics label.
func (c *Cache) Section(name string) *Section {
	return &Section{parent: c, name: name}
}

// Section is a scoped view over a Cache.
type Section struct {
	parent *Cache
	name   string
}

// Get reads from the parent cache with the section label applied to metrics.
func (s *Section) Get(key string) (interface{}, bool) {
	v, ok := s.parent.Get(s.name + ":" + key)
	if !ok {
		s.parent.metrics.Emit(s.name, "miss")
		s.parent.metrics.miss(s.name)
		return nil, false
	}
	s.parent.metrics.Emit(s.name, "hit")
	s.parent.metrics.hit(s.name)
	return v, true
}

// Set stores under the section label.
func (s *Section) Set(key string, value interface{}) {
	s.parent.Set(s.name+":"+key, value)
}

// SetTTL stores with a custom TTL under the section label.
func (s *Section) SetTTL(key string, value interface{}, ttl time.Duration) {
	s.parent.SetTTL(s.name+":"+key, value, ttl)
}

// Delete removes under the section label.
func (s *Section) Delete(key string) {
	s.parent.Delete(s.name + ":" + key)
}

func newMetrics() *Metrics {
	return &Metrics{counts: make(map[string]int64)}
}

// Metrics methods (exported so the collector can read them).
func (m *Metrics) Hit(section string) {
	m.mu.Lock()
	m.counts[section+":hit"]++
	m.mu.Unlock()
}

func (m *Metrics) Miss(section string) {
	m.mu.Lock()
	m.counts[section+":miss"]++
	m.mu.Unlock()
}

func (m *Metrics) Snapshot() map[string]int64 {
	m.mu.Lock()
	defer m.mu.Unlock()
	snapshot := make(map[string]int64, len(m.counts))
	for k, v := range m.counts {
		snapshot[k] = v
	}
	return snapshot
}

func (m *Metrics) hit(section string)  { m.Hit(section) }
func (m *Metrics) miss(section string) { m.Miss(section) }
