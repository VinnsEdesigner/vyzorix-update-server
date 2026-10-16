package command

import "sync"

//nolint:govet // fieldalignment: field order chosen for readability, not packing
type DispatchGuard struct {
	mu    sync.Mutex
	held  map[string]int
	limit int
}

func NewDispatchGuard(limit int) *DispatchGuard {
	if limit < 1 {
		limit = 1
	}
	return &DispatchGuard{held: map[string]int{}, limit: limit}
}

func (g *DispatchGuard) Acquire(deviceID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.held[deviceID] >= g.limit {
		return false
	}
	g.held[deviceID]++
	return true
}

func (g *DispatchGuard) Release(deviceID string) {
	g.mu.Lock()
	defer g.mu.Unlock()
	if g.held[deviceID] <= 1 {
		delete(g.held, deviceID)
		return
	}
	g.held[deviceID]--
}

func (g *DispatchGuard) Held(deviceID string) bool {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.held[deviceID] > 0
}

func DispatchResource(deviceID string) string { return "dispatch:" + deviceID }
