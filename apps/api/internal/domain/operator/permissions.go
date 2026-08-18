package operator

import "github.com/VinnsEdesigner/vyzorix/apps/api/internal/domain/permission"

// Permission represents a specific action an operator can perform.
// Note: Permissions are now derived from organization membership roles.
// The scoped permission engine (defaults + custom grants) performs evaluation.
type Permission string

const (
	// Device permissions.
	PermissionDeviceRead   Permission = "device:read"
	PermissionDeviceWrite  Permission = "device:write"
	PermissionDeviceDelete Permission = "device:delete"

	// Operator (member) management permissions.
	PermissionOperatorRead   Permission = "operator:read"
	PermissionOperatorWrite  Permission = "operator:write"
	PermissionOperatorDelete Permission = "operator:delete"

	// Update management permissions.
	PermissionUpdateRead   Permission = "update:read"
	PermissionUpdateWrite  Permission = "update:write"
	PermissionUpdateDelete Permission = "update:delete"

	// Audit log permissions.
	PermissionAuditRead Permission = "audit:read"

	// Settings permissions.
	PermissionSettingsRead  Permission = "settings:read"
	PermissionSettingsWrite Permission = "settings:write"

	// Admin permissions.
	PermissionAdminLockout     Permission = "admin:lockout"
	PermissionAdminImpersonate Permission = "admin:impersonate"
)

// scopedFromLegacy maps a legacy coarse Permission to the scoped (action, scope)
// form evaluated by the permission engine. Unknown permissions default to a
// no-match scope, so evaluation fails closed rather than granting everything.
func scopedFromLegacy(perm Permission) (action, scope string) {
	switch perm {
	case PermissionDeviceRead:
		return string(permission.ActionDeviceRead), permission.WildcardScope(permission.ScopeDevices)
	case PermissionDeviceWrite:
		return string(permission.ActionDeviceWrite), permission.WildcardScope(permission.ScopeDevices)
	case PermissionDeviceDelete:
		return string(permission.ActionDeviceDelete), permission.WildcardScope(permission.ScopeDevices)
	case PermissionOperatorRead:
		return string(permission.ActionMembersRead), permission.WildcardScope(permission.ScopeOrg)
	case PermissionOperatorWrite:
		return string(permission.ActionMembersWrite), permission.WildcardScope(permission.ScopeOrg)
	case PermissionOperatorDelete:
		return string(permission.ActionMembersDelete), permission.WildcardScope(permission.ScopeOrg)
	case PermissionUpdateRead:
		return string(permission.ActionUpdateRead), permission.WildcardScope(permission.ScopeUpdates)
	case PermissionUpdateWrite:
		return string(permission.ActionUpdateWrite), permission.WildcardScope(permission.ScopeUpdates)
	case PermissionUpdateDelete:
		return string(permission.ActionUpdateDelete), permission.WildcardScope(permission.ScopeUpdates)
	case PermissionAuditRead:
		return string(permission.ActionAuditRead), permission.WildcardScope(permission.ScopeOrg)
	case PermissionSettingsRead:
		return string(permission.ActionSettingsRead), permission.WildcardScope(permission.ScopeOrg)
	case PermissionSettingsWrite:
		return string(permission.ActionSettingsWrite), permission.WildcardScope(permission.ScopeOrg)
	case PermissionAdminLockout:
		return string(permission.ActionAdminLockout), permission.WildcardScope(permission.ScopeOrg)
	case PermissionAdminImpersonate:
		return string(permission.ActionAdminImpersonate), permission.WildcardScope(permission.ScopeOrg)
	default:
		return string(perm), ""
	}
}

// DefaultPermissions returns the default permissions for a standard operator.
// Deprecated: Use membership-based role checks instead.
func DefaultPermissions() []Permission {
	return []Permission{
		PermissionDeviceRead,
		PermissionDeviceWrite,
		PermissionUpdateRead,
		PermissionUpdateWrite,
		PermissionSettingsRead,
		PermissionSettingsWrite,
	}
}

// AdminPermissions returns all permissions for an admin.
// Deprecated: Use membership-based role checks instead.
func AdminPermissions() []Permission {
	return []Permission{
		PermissionDeviceRead,
		PermissionDeviceWrite,
		PermissionDeviceDelete,
		PermissionOperatorRead,
		PermissionOperatorWrite,
		PermissionOperatorDelete,
		PermissionUpdateRead,
		PermissionUpdateWrite,
		PermissionUpdateDelete,
		PermissionAuditRead,
		PermissionSettingsRead,
		PermissionSettingsWrite,
		PermissionAdminLockout,
	}
}

// SuperAdminPermissions returns all permissions.
// Deprecated: Use membership-based role checks instead.
func SuperAdminPermissions() []Permission {
	perms := AdminPermissions()
	perms = append(perms, PermissionAdminImpersonate)
	return perms
}
