package permission

// Action sets are aggregate permissions that expand into base actions. An admin
// can grant "device.manage" to a team rather than listing device.read,
// device.write, device.delete, device.assign individually. ExpandActionSets
// expands any aggregate action in a permission set into its members before
// evaluation, so a grant of an action set satisfies a check for any of its
// base actions before evaluation.

// Aggregate action names. These are intentionally distinct from the base
// Action constants so a grant of an aggregate is recognizable as such.
const (
	ActionSetDeviceManage   Action = "device.manage"
	ActionSetUpdateManage    Action = "update.manage"
	ActionSetMembersManage   Action = "members.manage"
	ActionSetSettingsManage  Action = "settings.manage"
	ActionSetOrgAdmin        Action = "org.admin"
	ActionSetCommandManage   Action = "command.manage"
)

// actionSets maps an aggregate action to the base actions it implies.
var actionSets = map[Action][]Action{
	ActionSetDeviceManage:  {ActionDeviceRead, ActionDeviceWrite, ActionDeviceDelete, ActionDeviceAssign},
	ActionSetUpdateManage:  {ActionUpdateRead, ActionUpdateWrite, ActionUpdateDelete},
	ActionSetMembersManage:  {ActionMembersRead, ActionMembersWrite, ActionMembersDelete},
	ActionSetSettingsManage: {ActionSettingsRead, ActionSettingsWrite},
	ActionSetOrgAdmin: {
		ActionSettingsRead, ActionSettingsWrite,
		ActionMembersRead, ActionMembersWrite, ActionMembersDelete,
		ActionAuditRead, ActionKeysManage,
	},
	ActionSetCommandManage: {ActionCommand},
}

// IsActionSet reports whether an action is an aggregate action set.
func IsActionSet(a Action) bool {
	_, ok := actionSets[a]
	return ok
}

// ExpandAction returns the base actions an action expands to. A base action
// expands to itself; an aggregate expands to its members.
func ExpandAction(a Action) []Action {
	if members, ok := actionSets[a]; ok {
		return members
	}
	return []Action{a}
}

// ExpandActionSets returns the permission set with every aggregate action
// expanded into its base members. The result contains only base actions; an
// aggregate grant therefore satisfies a check for any of its members.
func ExpandActionSets(perms ScopedPermissions) ScopedPermissions {
	if len(perms) == 0 {
		return perms
	}
	seen := make(map[ScopedPermission]struct{}, len(perms))
	out := make(ScopedPermissions, 0, len(perms))
	for _, p := range perms {
		for _, base := range ExpandAction(p.Action) {
			sp := ScopedPermission{Action: base, Scope: p.Scope}
			if _, ok := seen[sp]; !ok {
				seen[sp] = struct{}{}
				out = append(out, sp)
			}
		}
	}
	return out
}

// ActionSets returns the catalog of aggregate actions and their members, for
// introspection (e.g. an admin API listing grantable action sets).
func ActionSets() map[Action][]Action {
	out := make(map[Action][]Action, len(actionSets))
	for k, v := range actionSets {
		out[k] = append([]Action(nil), v...)
	}
	return out
}
