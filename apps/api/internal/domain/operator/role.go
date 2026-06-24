package operator

// OperatorRole represents the role of an operator in the system.
type OperatorRole string

const (
	RoleViewer     OperatorRole = "viewer"
	RoleOperator   OperatorRole = "operator"
	RoleSuperAdmin OperatorRole = "super_admin"
)

// IsValid checks if the role is a valid operator role.
func (r OperatorRole) IsValid() bool {
	switch r {
	case RoleViewer, RoleOperator, RoleSuperAdmin:
		return true
	default:
		return false
	}
}

// IsAtLeast checks if this role is at least the given role.
func (r OperatorRole) IsAtLeast(other OperatorRole) bool {
	roleLevel := map[OperatorRole]int{
		RoleViewer:     1,
		RoleOperator:   2,
		RoleSuperAdmin: 3,
	}

	return roleLevel[r] >= roleLevel[other]
}
