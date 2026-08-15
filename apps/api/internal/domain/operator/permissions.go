package operator

// Permission represents a specific action an operator can perform.
// Note: Permissions are now derived from organization membership roles.
// This file provides constants for permission checking but the actual.
// permission resolution is done through membership-based role checks.
type Permission string

const (
	// Device permissions.
	PermissionDeviceRead   Permission = "device:read"
	PermissionDeviceWrite  Permission = "device:write"
	PermissionDeviceDelete Permission = "device:delete"

	// Operator management permissions.
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
