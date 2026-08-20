package permission

// DefaultScopesForRole returns the scoped permissions a role grants by default.
// Mapping is by role name so this package stays dependency-free; the caller
// passes membership.Role (or another role string). Custom grants are added on
// top of these defaults by the operator's permission evaluator.
func DefaultScopesForRole(roleName string) ScopedPermissions {
	devicesAll := WildcardScope(ScopeDevices)
	updatesAll := WildcardScope(ScopeUpdates)
	orgAll := WildcardScope(ScopeOrg)
	telemetryAll := WildcardScope(ScopeTelemetry)
	diagnosticsAll := WildcardScope(ScopeDiagnostics)
	inboxAll := WildcardScope(ScopeInbox)

	var perms ScopedPermissions

	appendViewer := func() {
		perms = ScopedPermissions{
			{ActionDeviceRead, devicesAll},
			{ActionUpdateRead, updatesAll},
			{ActionSettingsRead, orgAll},
			{ActionTelemetryRead, telemetryAll},
			{ActionDiagnosticsRead, diagnosticsAll},
			{ActionEventsRead, devicesAll},
			{ActionLogsRead, devicesAll},
			{ActionConnectionsRead, devicesAll},
			{ActionStatsRead, orgAll},
		}
	}
	appendOperator := func() {
		perms = append(perms,
			ScopedPermission{ActionDeviceWrite, devicesAll},
			ScopedPermission{ActionCommand, devicesAll},
			ScopedPermission{ActionUpdateRead, updatesAll},
			ScopedPermission{ActionTelemetryRead, telemetryAll},
			ScopedPermission{ActionDiagnosticsRead, diagnosticsAll},
			ScopedPermission{ActionInboxRead, inboxAll},
			ScopedPermission{ActionInboxWrite, inboxAll},
		)
	}
	appendAdmin := func() {
		perms = append(perms,
			ScopedPermission{ActionDeviceDelete, devicesAll},
			ScopedPermission{ActionDeviceAssign, devicesAll},
			ScopedPermission{ActionUpdateWrite, updatesAll},
			ScopedPermission{ActionUpdateDelete, updatesAll},
			ScopedPermission{ActionSettingsWrite, orgAll},
			ScopedPermission{ActionKeysManage, orgAll},
			ScopedPermission{ActionAuditRead, orgAll},
			ScopedPermission{ActionTelemetryWrite, telemetryAll},
			ScopedPermission{ActionMembersRead, orgAll},
			ScopedPermission{ActionMembersWrite, orgAll},
			ScopedPermission{ActionMembersDelete, orgAll},
		)
	}
	appendSuperAdmin := func() {
		perms = append(perms,
			ScopedPermission{ActionAdminLockout, orgAll},
			ScopedPermission{ActionAdminImpersonate, orgAll},
		)
	}

	switch roleName {
	case "viewer":
		appendViewer()
	case "operator":
		appendViewer()
		appendOperator()
	case "admin":
		appendViewer()
		appendOperator()
		appendAdmin()
	case "super_admin":
		appendViewer()
		appendOperator()
		appendAdmin()
		appendSuperAdmin()
	default:
		// Unknown role: fall back to viewer-only (fail to the least privilege).
		appendViewer()
	}

	return perms
}
