package featuremgmt

import (
	"testing"
)

func TestNewManager_AllOn(t *testing.T) {
	m := NewManager(map[Feature]bool{
		ScopedRBAC:      true,
		DeviceGroups:    true,
		ServerDrivenOrg: true,
	})
	if !m.IsEnabled(ScopedRBAC) {
		t.Error("ScopedRBAC should be enabled")
	}
	if !m.IsEnabled(DeviceGroups) {
		t.Error("DeviceGroups should be enabled")
	}
	if m.IsEnabled(ActionSets) {
		t.Error("ActionSets should be off (not in defaults)")
	}
}

func TestNewManager_AllOff(t *testing.T) {
	m := NewManager(map[Feature]bool{})
	for _, f := range []Feature{ScopedRBAC, DeviceGroups, ActionSets} {
		if m.IsEnabled(f) {
			t.Errorf("%s should be off", f)
		}
	}
}

func TestNilManager_AllOff(t *testing.T) {
	var m *Manager
	if m.IsEnabled(ScopedRBAC) {
		t.Error("nil manager should report all features off")
	}
}

func TestManager_EnvOverride(t *testing.T) {
	t.Setenv("FEATURE_SCOPED_RBAC", "true")
	t.Setenv("FEATURE_DEVICE_GROUPS", "false")

	m := NewManager(map[Feature]bool{
		ScopedRBAC:   false,
		DeviceGroups: true,
	})
	if !m.IsEnabled(ScopedRBAC) {
		t.Error("env should have enabled ScopedRBAC")
	}
	if m.IsEnabled(DeviceGroups) {
		t.Error("env should have disabled DeviceGroups")
	}
}

func TestManager_EnvOverride_One(t *testing.T) {
	t.Setenv("FEATURE_PERMISSION_CACHE", "1")
	m := NewManager(map[Feature]bool{PermissionCache: false})
	if !m.IsEnabled(PermissionCache) {
		t.Error("FEATURE_PERMISSION_CACHE=1 should enable")
	}
}

func TestManager_SetRuntime(t *testing.T) {
	m := NewManager(map[Feature]bool{ActionSets: false})
	if m.IsEnabled(ActionSets) {
		t.Fatal("should start off")
	}
	m.Set(ActionSets, true)
	if !m.IsEnabled(ActionSets) {
		t.Error("Set(true) should enable")
	}
	m.Set(ActionSets, false)
	if m.IsEnabled(ActionSets) {
		t.Error("Set(false) should disable")
	}
}

func TestManager_IsEnabled_UnknownFeature(t *testing.T) {
	m := NewManager(map[Feature]bool{ScopedRBAC: true})
	if m.IsEnabled("nonexistent_feature") {
		t.Error("unknown feature should be off")
	}
}

func TestManager_EnvOnlyAffectsKnownFeatures(t *testing.T) {
	t.Setenv("FEATURE_UNKNOWN_THING", "true")
	m := NewManager(map[Feature]bool{ScopedRBAC: true})
	if m.IsEnabled("unknown_thing") {
		t.Error("unknown feature should not be enabled by env")
	}
}
