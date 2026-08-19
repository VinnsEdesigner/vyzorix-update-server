package permission

import (
	"sync"
	"time"
)

// Cache is a TTL cache for permission-resolution results. It stores the
// ResourcePermissions returned by ListEffective keyed by (operatorID, orgID),
// so repeated authorization checks within the TTL window skip the database. An
// invalidation call (on grant mutation) clears entries for the affected
// operator or org. The cache is lock-free on the hot read path: a
// sync.Map holds entries, each entry holds a deadline, and a stale entry is
// simply refetched.
type Cache struct {
	entries sync.Map // map[cacheKey]cacheEntry — lock-free cached grants.
	ttl     time.Duration
}

type cacheKey struct {
	operatorID string
	orgID      string
}

type cacheEntry struct {
	deadline time.Time
	perms    []*ResourcePermission
}

// NewCache builds a permission cache with the given TTL. A TTL of 0 disables
// caching (every read misses).
func NewCache(ttl time.Duration) *Cache {
	return &Cache{ttl: ttl}
}

// Get returns the cached grants for (operatorID, orgID) if present and fresh.
func (c *Cache) Get(operatorID, orgID string) []*ResourcePermission {
	if c == nil || c.ttl == 0 {
		return nil
	}
	v, ok := c.entries.Load(cacheKey{operatorID, orgID})
	if !ok {
		return nil
	}
	e, ok := v.(cacheEntry)
	if !ok {
		return nil
	}
	if time.Now().After(e.deadline) {
		return nil
	}
	return e.perms
}

// Put stores grants for (operatorID, orgID) with the cache's TTL.
func (c *Cache) Put(operatorID, orgID string, perms []*ResourcePermission) {
	if c == nil || c.ttl == 0 {
		return
	}
	c.entries.Store(cacheKey{operatorID, orgID}, cacheEntry{
		perms:    perms,
		deadline: time.Now().Add(c.ttl),
	})
}

// InvalidateOperator removes all cache entries for an operator (call on
// operator-grant mutation).
func (c *Cache) InvalidateOperator(operatorID string) {
	if c == nil {
		return
	}
	c.entries.Range(func(key, _ any) bool {
		if k, ok := key.(cacheKey); ok && k.operatorID == operatorID {
			c.entries.Delete(k)
		}
		return true
	})
}

// InvalidateOrg removes all cache entries for an org (call on team-grant
// mutation, since team grants affect every member of the team).
func (c *Cache) InvalidateOrg(orgID string) {
	if c == nil {
		return
	}
	c.entries.Range(func(key, _ any) bool {
		if k, ok := key.(cacheKey); ok && k.orgID == orgID {
			c.entries.Delete(k)
		}
		return true
	})
}

// InvalidateAll clears every cache entry.
func (c *Cache) InvalidateAll() {
	if c == nil {
		return
	}
	c.entries.Range(func(key, _ any) bool {
		c.entries.Delete(key)
		return true
	})
}
