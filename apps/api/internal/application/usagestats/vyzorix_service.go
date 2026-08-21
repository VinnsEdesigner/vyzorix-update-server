package usagestats

import (
	"context"
	"sync"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/featuremgmt"
)

// Service computes usage stats on demand and periodic snapshots.
type Service struct {
	collector *Collector
	features  *featuremgmt.Manager
	last      *Snapshot
	mu        sync.RWMutex
}

// NewService creates a usage stats Service.
func NewService(collector *Collector, features *featuremgmt.Manager) *Service {
	return &Service{collector: collector, features: features}
}

// Collect performs a snapshot and stores it.
func (s *Service) Collect(ctx context.Context) {
	snap := s.collector.Query(ctx)
	if s.features != nil {
		snap.Toggles = map[string]bool{
			"scoped_rbac":       s.features.IsEnabled(featuremgmt.ScopedRBAC),
			"device_groups":     s.features.IsEnabled(featuremgmt.DeviceGroups),
			"server_driven_org": s.features.IsEnabled(featuremgmt.ServerDrivenOrg),
			"action_sets":       s.features.IsEnabled(featuremgmt.ActionSets),
			"scope_resolvers":   s.features.IsEnabled(featuremgmt.ScopeResolvers),
			"permission_cache":  s.features.IsEnabled(featuremgmt.PermissionCache),
		}
	}
	s.mu.Lock()
	s.last = snap
	s.mu.Unlock()
}

// Snapshot returns the last collected snapshot (nil until first Collect).
func (s *Service) Snapshot() *Snapshot {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.last
}
