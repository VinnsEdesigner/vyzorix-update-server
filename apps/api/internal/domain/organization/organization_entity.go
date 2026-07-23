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

// Level constants for role hierarchy comparisons.
const (
	LevelViewer    = 1
	LevelOperator  = 2
	LevelAdmin     = 3
	LevelSuperAdmin = 4
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
	return true // All roles can view devices.
}

// CanManageAPIKeys returns true if the role can manage API keys.
func (r OrganizationRole) CanManageAPIKeys() bool {
	return r.Level() >= RoleAdmin.Level()
}

// Organization represents an organization (tenant) in the system with explicit lifecycle management.
type Organization struct {
	// Lifecycle tracks the organization lifecycle state.
	Lifecycle OrganizationLifecycle

	// Infrastructure fields.
	ID          string
	Name        string
	Description string // Optional description of the organization.
	CreatedBy   string
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time // Soft delete - set when archived.
	MaxMembers  int
	MemberCount int
}

// NewOrganization creates a new Organization with active lifecycle.
func NewOrganization(id, name, createdBy string) *Organization {
	return &Organization{
		ID:         id,
		Name:       name,
		CreatedBy:  createdBy,
		Lifecycle:  OrganizationLifecycleActive,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// Activate transitions the organization to active state.
func (o *Organization) Activate() error {
	return o.Lifecycle.TransitionTo(OrganizationLifecycleActive)
}

// Deactivate transitions the organization to inactive state (suspended).
func (o *Organization) Deactivate() error {
	return o.Lifecycle.TransitionTo(OrganizationLifecycleInactive)
}

// Archive transitions the organization to archived state (soft delete).
func (o *Organization) Archive() error {
	if err := o.Lifecycle.TransitionTo(OrganizationLifecycleArchived); err != nil {
		return err
	}
	now := time.Now()
	o.DeletedAt = &now
	return nil
}

// IsActive returns true if the organization is active.
func (o *Organization) IsActive() bool {
	return o.Lifecycle.IsActive()
}

// IsInactive returns true if the organization is inactive.
func (o *Organization) IsInactive() bool {
	return o.Lifecycle.IsInactive()
}

// IsArchived returns true if the organization has been archived.
func (o *Organization) IsArchived() bool {
	return o.Lifecycle.IsArchived()
}

// IsValid returns true if the organization has all required fields.
func (o *Organization) IsValid() bool {
	return o.ID != "" && o.Name != "" && o.CreatedBy != ""
}

// IsDeleted returns true if the organization has been soft-deleted.
func (o *Organization) IsDeleted() bool {
	return o.Lifecycle.IsArchived()
}

// CanAddMember returns true if the organization can accept more members.
func (o *Organization) CanAddMember() bool {
	return o.Lifecycle.CanAcceptMembers() && (o.MaxMembers <= 0 || o.MemberCount < o.MaxMembers)
}

// OrganizationMember represents a membership linking an operator to an organization.
type OrganizationMember struct {
	// Lifecycle tracks the membership lifecycle state.
	Lifecycle MemberLifecycle

	// Core fields.
	ID             string
	OrganizationID string
	OperatorID     string
	Role           OrganizationRole
	InvitedBy      *string
	JoinedAt       time.Time
	RemovedAt      *time.Time // Soft delete - set when removed.
	SuspendedAt    *time.Time

	// Populated fields (joined from other tables).
	OperatorName  string
	OperatorEmail string
}

// NewMember creates a new OrganizationMember with invited lifecycle.
func NewMember(id, organizationID, operatorID string, role OrganizationRole) *OrganizationMember {
	return &OrganizationMember{
		ID:             id,
		OrganizationID: organizationID,
		OperatorID:     operatorID,
		Role:           role,
		Lifecycle:      MemberLifecycleInvited,
		JoinedAt:       time.Now(),
	}
}

// Invite transitions to invited state (initial state).
func (m *OrganizationMember) Invite() {
	m.Lifecycle = MemberLifecycleInvited
	m.RemovedAt = nil
	m.SuspendedAt = nil
}

// Join transitions from invited to active.
func (m *OrganizationMember) Join() error {
	return m.Lifecycle.TransitionTo(MemberLifecycleActive)
}

// Suspend transitions from active to suspended.
func (m *OrganizationMember) Suspend() error {
	if err := m.Lifecycle.TransitionTo(MemberLifecycleSuspended); err != nil {
		return err
	}
	now := time.Now()
	m.SuspendedAt = &now
	return nil
}

// Reinstate transitions from suspended back to active.
func (m *OrganizationMember) Reinstate() error {
	return m.Lifecycle.TransitionTo(MemberLifecycleActive)
}

// Remove transitions to removed state (soft delete).
func (m *OrganizationMember) Remove() error {
	if err := m.Lifecycle.TransitionTo(MemberLifecycleRemoved); err != nil {
		return err
	}
	now := time.Now()
	m.RemovedAt = &now
	return nil
}

// IsInvited returns true if the member is invited.
func (m *OrganizationMember) IsInvited() bool {
	return m.Lifecycle.IsInvited()
}

// IsActive returns true if the member is active.
func (m *OrganizationMember) IsActive() bool {
	return m.Lifecycle.IsActive()
}

// IsSuspended returns true if the member is suspended.
func (m *OrganizationMember) IsSuspended() bool {
	return m.Lifecycle.IsSuspended()
}

// IsRemoved returns true if the member is removed.
func (m *OrganizationMember) IsRemoved() bool {
	return m.Lifecycle.IsRemoved()
}

// CanAccessResources returns true if the member can access organization resources.
func (m *OrganizationMember) CanAccessResources() bool {
	return m.Lifecycle.CanAccessResources()
}

// UpdateRole updates the member's role.
func (m *OrganizationMember) UpdateRole(role OrganizationRole) {
	m.Role = role
}

// CreateOrganizationRequest represents a request to create an organization.
type CreateOrganizationRequest struct {
	Name        string // Optional - defaults to "personal" if empty.
	Description string // Optional - organization description.
	MaxMembers int    // Optional - max members limit (0 = default).
	Role        string // Required - creator's role: "super_admin" or "admin".
}

// Validate validates the create organization request.
func (r *CreateOrganizationRequest) Validate() error {
	// Name is optional - defaults to "personal".
	if r.Name == "" {
		r.Name = "personal"
	}
	if len(r.Name) < 2 {
		return errors.New("organization name must be at least 2 characters")
	}
	if len(r.Name) > 100 {
		return errors.New("organization name must be at most 100 characters")
	}
	// Role is required - only super_admin or admin allowed.
	if r.Role != "super_admin" && r.Role != "admin" {
		return errors.New("role must be 'super_admin' or 'admin'")
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
