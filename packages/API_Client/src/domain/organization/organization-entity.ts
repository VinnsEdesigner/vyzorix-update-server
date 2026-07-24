/**
 * Organization domain types aligned with server models.
 * Supports multi-tenant organization model.
 */

// Organization lifecycle states
export type OrganizationLifecycle = "active" | "inactive" | "archived";

// Organization roles with privilege hierarchy
export type OrganizationRole = "super_admin" | "admin" | "operator" | "viewer";

// Role privilege levels (higher = more privileges)
export const ROLE_LEVELS: Record<OrganizationRole, number> = {
  viewer: 1,
  operator: 2,
  admin: 3,
  super_admin: 4,
};

/**
 * Get the privilege level for a role.
 */
export function getRoleLevel(role: OrganizationRole): number {
  return ROLE_LEVELS[role] ?? 0;
}

/**
 * Check if a role has admin privileges or higher.
 */
export function isAdminRole(role: OrganizationRole): boolean {
  return getRoleLevel(role) >= ROLE_LEVELS.admin;
}

/**
 * Check if a role is super_admin.
 */
export function isSuperAdminRole(role: OrganizationRole): boolean {
  return role === "super_admin";
}

// Organization entity
export interface Organization {
  id: string;
  name: string;
  description?: string;
  createdBy: string;
  maxMembers: number;
  memberCount: number;
  lifecycle: OrganizationLifecycle;
  createdAt: Date;
  updatedAt: Date;
  deletedAt?: Date;
}

// Organization list item (lightweight)
export interface OrganizationListItem {
  id: string;
  name: string;
  description?: string;
  memberCount: number;
  lifecycle: OrganizationLifecycle;
  createdAt: Date;
}

// Member lifecycle states
export type MemberLifecycle = "invited" | "active" | "suspended" | "removed";

// Organization member entity
export interface OrganizationMember {
  id: string;
  organizationId: string;
  operatorId: string;
  role: OrganizationRole;
  lifecycle: MemberLifecycle;
  invitedBy?: string;
  joinedAt: Date;
  removedAt?: Date;
  suspendedAt?: Date;
  // Populated fields
  operatorName?: string;
  operatorEmail?: string;
}

// Member list item
export interface MemberListItem {
  id: string;
  operatorId: string;
  operatorName: string;
  operatorEmail: string;
  role: OrganizationRole;
  lifecycle: MemberLifecycle;
  joinedAt: Date;
}

// Create organization request
export interface CreateOrganizationRequest {
  name: string;
  description?: string;
  maxMembers?: number;
}

// Update organization request
export interface UpdateOrganizationRequest {
  name?: string;
  description?: string;
  maxMembers?: number;
}

// Create member request
export interface CreateMemberRequest {
  email: string;
  role: OrganizationRole;
  inviterNotes?: string;
}

// Update member role request
export interface UpdateMemberRoleRequest {
  role: OrganizationRole;
}

// API response types (raw from server)
export interface OrganizationApiResponse {
  id: string;
  name: string;
  description?: string;
  created_by: string;
  max_members: number;
  member_count: number;
  lifecycle: string;
  created_at: string;
  updated_at: string;
  deleted_at?: string;
}

export interface MemberApiResponse {
  id: string;
  organization_id: string;
  operator_id: string;
  role: string;
  lifecycle: string;
  invited_by?: string;
  joined_at: string;
  removed_at?: string;
  suspended_at?: string;
  operator_name?: string;
  operator_email?: string;
}

/**
 * Check if organization is active and can accept members.
 */
export function canAcceptMembers(org: Organization): boolean {
  if (org.lifecycle !== "active") return false;
  if (org.maxMembers > 0 && org.memberCount >= org.maxMembers) return false;
  return true;
}

/**
 * Check if organization is deleted (archived).
 */
export function isOrganizationDeleted(org: Organization): boolean {
  return org.lifecycle === "archived" || org.deletedAt !== undefined;
}

/**
 * Check if member is active and can access resources.
 */
export function canMemberAccessResources(member: OrganizationMember): boolean {
  return member.lifecycle === "active";
}

/**
 * Check if member can be managed by the given role.
 */
export function canManageMember(
  managerRole: OrganizationRole,
  targetRole: OrganizationRole
): boolean {
  const managerLevel = getRoleLevel(managerRole);
  const targetLevel = getRoleLevel(targetRole);
  // Can only manage roles at or below your level
  return managerLevel > targetLevel;
}
