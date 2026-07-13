package operator

// OperatorRole represents the role of an operator in the system.
// Role hierarchy (higher level = more permissions):
//   - Viewer (10): Read-only access
//   - Operator (50): Standard operations, manage own resources
//   - Admin (80): Tenant management, manage other operators
//   - SuperAdmin (100): System-wide ownership, can do anything
type OperatorRole string

const (
	RoleViewer     OperatorRole = "viewer"
	RoleOperator   OperatorRole = "operator"
	RoleAdmin      OperatorRole = "admin"
	RoleSuperAdmin OperatorRole = "super_admin"
)

// Role levels for hierarchy comparisons
const (
	LevelViewer     = 10
	LevelOperator   = 50
	LevelAdmin      = 80
	LevelSuperAdmin = 100
)

// roleLevels maps roles to their permission levels
var roleLevels = map[OperatorRole]int{
	RoleViewer:     LevelViewer,
	RoleOperator:   LevelOperator,
	RoleAdmin:      LevelAdmin,
	RoleSuperAdmin: LevelSuperAdmin,
}

// IsValid checks if the role is a valid operator role.
func (r OperatorRole) IsValid() bool {
	switch r {
	case RoleViewer, RoleOperator, RoleAdmin, RoleSuperAdmin:
		return true
	default:
		return false
	}
}

// Level returns the permission level for this role.
func (r OperatorRole) Level() int {
	if level, ok := roleLevels[r]; ok {
		return level
	}
	return 0
}

// IsAtLeast checks if this role is at least the given role.
func (r OperatorRole) IsAtLeast(other OperatorRole) bool {
	return r.Level() >= other.Level()
}

// CanPromoteTo checks if this role can promote to the target role.
func (r OperatorRole) CanPromoteTo(target OperatorRole) bool {
	// Can only promote to roles at or below own level, but not equal or higher
	return r.Level() > target.Level() && r.Level() >= LevelAdmin
}

// CanDemote checks if this role can demote the target role.
func (r OperatorRole) CanDemote(target OperatorRole) bool {
	// Can only demote roles strictly below own level
	return r.Level() > target.Level() && r.Level() >= LevelAdmin
}

// IsAdmin returns true if this role has admin privileges.
func (r OperatorRole) IsAdmin() bool {
	return r.Level() >= LevelAdmin
}

// IsSuperAdmin returns true if this role has super admin privileges.
func (r OperatorRole) IsSuperAdmin() bool {
	return r == RoleSuperAdmin
}
