// Package featuremgmt provides a feature-toggle system for gradual rollout and
// runtime control of new behavior. Toggles are declared as constants, resolved
// at startup from environment variables, and checked at call sites via
// IsEnabled. A nil Manager (the zero value) reports every toggle as off, so
// callers can safely check features without wiring a manager.
package featuremgmt

import (
	"os"
	"strings"
	"sync"
)

// Feature is a named toggle.
type Feature string

// Declared toggles. Add new constants here as features ship.
const (
	// ScopedRBAC gates the scoped permission engine (action+scope evaluation,
	// custom grants, device groups). When off, the legacy role-tier checks apply.
	ScopedRBAC Feature = "scoped_rbac"

	// DeviceGroups gates team/device-group-based device access.
	DeviceGroups Feature = "device_groups"

	// ServerDrivenOrg gates server-side org resolution (session/key authoritative,
	// conflicting client header rejected).
	ServerDrivenOrg Feature = "server_driven_org"

	// ActionSets gates aggregate-action expansion in the permission evaluator.
	ActionSets Feature = "action_sets"

	// ScopeResolvers gates attribute-scope resolution (devices:imei:X → concrete).
	ScopeResolvers Feature = "scope_resolvers"

	// PermissionCache gates the in-memory permission-resolution cache.
	PermissionCache Feature = "permission_cache"
)

// State is the resolved on/off state of a toggle.
type State int

const (
	StateOff State = iota
	StateOn
)

// Manager resolves feature toggles from environment variables at startup and
// serves them from a lock-free read path. The zero value is a valid manager in
// which every toggle is off.
type Manager struct {
	// states holds the resolved on/off state of each toggle.
	states sync.Map // map[Feature]State — lock-free toggle states.
}

// NewManager builds a Manager from a feature→default map, then applies
// environment overrides. An env var named FEATURE_<UPPER_NAME>=true|false
// overrides the default (e.g. FEATURE_SCOPED_RBAC=true).
func NewManager(defaults map[Feature]bool) *Manager {
	m := &Manager{}
	for f, on := range defaults {
		m.set(f, on)
	}
	m.applyEnv()
	return m
}

func (m *Manager) set(f Feature, on bool) {
	if on {
		m.states.Store(f, StateOn)
	} else {
		m.states.Store(f, StateOff)
	}
}

func (m *Manager) applyEnv() {
	m.states.Range(func(key, value any) bool {
		f, ok := key.(Feature)
		if !ok {
			return true
		}
		envName := "FEATURE_" + strings.ToUpper(strings.ReplaceAll(string(f), "_", "_"))
		if raw, ok := os.LookupEnv(envName); ok {
			m.set(f, strings.EqualFold(raw, "true") || raw == "1")
		}
		return true
	})
}

// IsEnabled reports whether a feature toggle is on.
func (m *Manager) IsEnabled(f Feature) bool {
	if m == nil {
		return false
	}
	v, ok := m.states.Load(f)
	if !ok {
		return false
	}
	state, ok := v.(State)
	return ok && state == StateOn
}

// Set overrides a toggle at runtime (used by tests and admin endpoints).
func (m *Manager) Set(f Feature, on bool) {
	if m != nil {
		m.set(f, on)
	}
}
