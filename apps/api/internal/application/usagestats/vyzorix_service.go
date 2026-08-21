package usagestats

import (
	"context"

	"github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/featuremgmt"
)

// Service computes usage stats on demand and periodic snapshots.
// Service computes usage stats on demand and periodic snapshots.
type Service struct {
	collector *Collector
	features  *featuremgmt.Manager
	last      *Snapshot
}

// NewService creates a usage stats Service.
func NewService(collector *Collector, features *featuremgmt.Manager) *Service {
	return &Service{collector: collector, features: features}
}

// Snapshot returns the last collected snapshot (nil if none exists).
func (s *Service) Snapshot() *Snapshot {
	return s.last
}

// Collect performs a snapshot and stores it.
func (s *Service) Collect(ctx context.Context) {
	snap := s.collector.Query(ctx)
	if s.features != nil {
		snap.Toggles = map[string]bool{
			"scoped_rbac":           s.features.IsEnabled(featuremgmt.ScopedRBAC),
			"device_groups":         s.features.IsEnabled(featuremgmt.DeviceGroups),
			"server_driven_org":     s.features.IsEnabled(featuremgmt.ServerDrivenOrg),
			"action_sets":           s.features.IsEnabled(featuremgmt.ActionSets),
			"scope_resolvers":       s.features.IsEnabled(featuremgmt.ScopeResolvers),
			"permission_cache":      s.features.IsEnabled(featuremgmt.PermissionCache),
		}
	}
	s.last = snap
}

// ServiceName returns the service name for the update-checker response.
func (s *Service) Name() string {
	return "vyzorix-usagestats"
}
