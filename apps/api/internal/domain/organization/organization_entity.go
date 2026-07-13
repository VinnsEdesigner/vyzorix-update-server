package organization

import (
	"errors"
	"time"
)

// OrganizationRole represents the role of a member within an organization.
type OrganizationRole string

const (
	RoleSuperAdmin OrganizationRole = "super_admin"
	RoleAdmin     OrganizationRole = "admin"
	RoleOperator  OrganizationRole = "operator"
	RoleViewer    OrganizationRole = "viewer"
)

// RoleLevel returns the privilege level for a role (higher = more privileges).
func (r OrganizationRole) Level() int {
	switch r {
	case RoleSuperAdmin:
		return 4
	case RoleAdmin:
		return 3
	case RoleOperator:
		return 2
	case RoleViewer:
		return 1
	default:
		return 0
	}
}

// IsSuperAdmin returns true if the role is super_admin.
func (r OrganizationRole) IsSuperAdmin() bool {
	return r == RoleSuperAdmin
}

// IsAdmin returns true if the role is admin or super_admin.
func (r OrganizationRole) IsAdmin() bool {
	return r == RoleAdmin || r == RoleSuperAdmin
}

// CanManageMembers returns true if the role can manage organization members.
func (r OrganizationRole) CanManageMembers() bool {
	return r.Level() >= RoleAdmin.Level()
}

// CanManageOrganization returns true if the role can manage organization settings.
func (r OrganizationRole) CanManageOrganization() bool {
	return r.Level() >= RoleAdmin.Level()
}

// CanDeleteOrganization returns true if the role can delete the organization.
func (r OrganizationRole) CanDeleteOrganization() bool {
	return r == RoleSuperAdmin
}

// CanManageDevices returns true if the role can register/manage devices.
func (r OrganizationRole) CanManageDevices() bool {
	return r.Level() >= RoleOperator.Level()
}

// CanViewDevices returns true if the role can view devices.
func (r OrganizationRole) CanViewDevices() bool {
	return true // All roles can view devices
}

// CanManageAPIKeys returns true if the role can manage API keys.
func (r OrganizationRole) CanManageAPIKeys() bool {
	return r.Level() >= RoleAdmin.Level()
}

// Organization represents an organization (tenant) in the system.
type Organization struct {
	ID         string
	Name       string
	CreatedBy  string
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  *time.Time // Soft delete
	IsActive   bool
	MaxMembers int
	MemberCount int
}

// IsValid returns true if the organization has all required fields.
func (o *Organization) IsValid() bool {
	return o.ID != "" && o.Name != "" && o.CreatedBy != ""
}

// IsDeleted returns true if the organization has been soft-deleted.
func (o *Organization) IsDeleted() bool {
	return o.DeletedAt != nil
}

// CanAddMember returns true if the organization can accept more members.
func (o *Organization) CanAddMember() bool {
	return o.IsActive && (o.MaxMembers <= 0 || o.MemberCount < o.MaxMembers)
}

// OrganizationMember represents a membership linking an operator to an organization.
type OrganizationMember struct {
	ID             string
	OrganizationID string
	OperatorID     string
	Role           OrganizationRole
	InvitedBy      *string
	JoinedAt       time.Time
	RemovedAt      *time.Time // Soft delete
	Status         MemberStatus

	// Populated fields (joined from other tables)
	OperatorName  string
	OperatorEmail string
}

// MemberStatus represents the status of a membership.
type MemberStatus string

const (
	MemberStatusActive  MemberStatus = "active"
	MemberStatusRemoved MemberStatus = "removed"
)

// IsActive returns true if the membership is active.
func (m *OrganizationMember) IsActive() bool {
	return m.Status == MemberStatusActive
}

// CreateOrganizationRequest represents a request to create an organization.
type CreateOrganizationRequest struct {
	Name       string
	MaxMembers int
}

// Validate validates the create organization request.
func (r *CreateOrganizationRequest) Validate() error {
	if r.Name == "" {
		return errors.New("organization name is required")
	}
	if len(r.Name) < 2 {
		return errors.New("organization name must be at least 2 characters")
	}
	if len(r.Name) > 100 {
		return errors.New("organization name must be at most 100 characters")
	}
	return nil
}

// UpdateOrganizationRequest represents a request to update an organization.
type UpdateOrganizationRequest struct {
	Name       *string
	MaxMembers *int
}

// UpdateMemberRoleRequest represents a request to update a member's role.
type UpdateMemberRoleRequest struct {
	Role OrganizationRole
}

// Validate validates the update member role request.
func (r *UpdateMemberRoleRequest) Validate() error {
	switch r.Role {
	case RoleSuperAdmin, RoleAdmin, RoleOperator, RoleViewer:
		return nil
	default:
		return errors.New("invalid role")
	}
}
