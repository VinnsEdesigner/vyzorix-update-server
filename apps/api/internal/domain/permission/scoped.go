// Package permission provides scoped, resource-aware authorization. A
// permission is an action granted on a resource scope, where the scope is a
// hierarchical resource path (for example "devices:imei:<id>") that may end in
// a trailing wildcard ("devices:*"). Evaluation matches an operator's action
// and scope against its granted permissions rather than relying on role tier
// alone. This is the scoped-RBAC model the coarse role-bundle replaced.
package permission

import "strings"

// Action is a resource-scoped operation, for example "device.command".
type Action string

const (
	// Device resource actions.
	ActionDeviceRead   Action = "device.read"
	ActionDeviceWrite  Action = "device.write"
	ActionDeviceDelete Action = "device.delete"
	ActionDeviceAssign Action = "device.assign"

	// Command execution on a device.
	ActionCommand Action = "command.execute"

	// Telemetry: query device telemetry history, latest, stats, export.
	ActionTelemetryRead  Action = "telemetry.read"
	ActionTelemetryWrite Action = "telemetry.write"

	// Diagnostics: inspect device state and view diagnostic timelines.
	ActionDiagnosticsRead Action = "diagnostics.read"

	// Device events and logs: view the event/log timeline for a device.
	ActionEventsRead Action = "events.read"
	ActionLogsRead   Action = "logs.read"

	// Connection status: view device WebSocket/FCM connection state and metrics.
	ActionConnectionsRead Action = "connections.read"

	// Cluster stats: aggregate dashboard/cluster health and utilization metrics.
	ActionStatsRead Action = "stats.read"

	// Device inbox: view/approve device registration inbox entries.
	ActionInboxRead  Action = "inbox.read"
	ActionInboxWrite Action = "inbox.write"

	// Update management actions.
	ActionUpdateRead   Action = "update.read"
	ActionUpdateWrite  Action = "update.write"
	ActionUpdateDelete Action = "update.delete"

	// Settings, keys, members, audit, and admin actions.
	ActionSettingsRead     Action = "settings.read"
	ActionSettingsWrite    Action = "settings.write"
	ActionMembersRead      Action = "members.read"
	ActionMembersWrite     Action = "members.write"
	ActionMembersDelete    Action = "members.delete"
	ActionAuditRead        Action = "audit.read"
	ActionKeysManage       Action = "keys.manage"
	ActionAdminLockout     Action = "admin.lockout"
	ActionAdminImpersonate Action = "admin.impersonate"
)

// Well-known resource scope prefixes. A scope is a sequence of `prefix:id`
// segments optionally ending with a trailing wildcard `:*`.
const (
	ScopeDevices     = "devices"
	ScopeOrg         = "org"
	ScopeUpdates     = "updates"
	ScopeTelemetry   = "telemetry"
	ScopeDiagnostics = "diagnostics"
	ScopeInbox       = "inbox"
)

// ScopedPermission is one action granted on one resource scope.
type ScopedPermission struct {
	Action Action
	Scope  string
}

// ScopedPermissions is the set an operator may act with in an organization.
type ScopedPermissions []ScopedPermission

// Grants reports whether the set grants `action` on the resource `scope`. A
// grant matches when the requested scope equals the granted scope or sits under
// the granted scope's trailing wildcard.
func (s ScopedPermissions) Grants(action Action, scope string) bool {
	for _, p := range s {
		if p.Action == action && MatchScope(p.Scope, scope) {
			return true
		}
	}
	return false
}

// MatchScope reports whether a granted scope covers a target scope. A trailing
// ":*" in the grant is a prefix wildcard, so "devices:*" covers every device.
// Wildcards anywhere else are invalid and never match (fail closed).
func MatchScope(grant, target string) bool {
	if grant == "" || target == "" {
		return false
	}
	if !ValidScope(grant) || !ValidScope(target) {
		return false
	}
	if grant == target {
		return true
	}
	// Trailing wildcard: match as a prefix.
	if strings.HasSuffix(grant, ":*") {
		prefix := grant[:len(grant)-1] // Strip the wildcard itself.
		return strings.HasPrefix(target, prefix)
	}
	return false
}

// ValidScope reports whether a scope string is well-formed: non-empty segments,
// and a wildcard only permitted as the final character. Wildcards in any other
// position or a bare empty segment make the scope invalid.
func ValidScope(scope string) bool {
	if scope == "" {
		return false
	}
	star := strings.Index(scope, "*")
	if star >= 0 && star != len(scope)-1 {
		return false // Wildcard only allowed as the trailing character.
	}
	for _, seg := range strings.Split(scope, ":") {
		if seg == "" {
			return false
		}
	}
	return true
}

// BuildScope joins a resource prefix and id into a scope string.
func BuildScope(prefix, id string) string {
	if id == "" {
		return prefix
	}
	return prefix + ":" + id
}

// WildcardScope returns the wildcard form of a resource prefix, e.g. "devices:*".
func WildcardScope(prefix string) string {
	return prefix + ":*"
}
