export type OrganizationRole = "super_admin" | "admin" | "operator" | "viewer";

export type MemberLifecycle = "invited" | "active" | "suspended" | "removed";

export interface Organization {
  id: string;
  name: string;
  description: string;
  created_by: string;
  max_members: number;
  member_count: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

export interface OrganizationMember {
  id: string;
  organization_id: string;
  operator_id: string;
  role: OrganizationRole;
  status: MemberLifecycle;
  invited_by?: string;
  joined_at: string;
  removed_at?: string;
  operator_name?: string;
  operator_email?: string;
}

export interface CreateOrganizationRequest {
  name: string;
  description: string;
  role: OrganizationRole;
  maxMembers?: number;
}

export interface UpdateOrganizationRequest {
  name?: string;
  maxMembers?: number;
  isActive?: boolean;
}

export interface UpdateMemberRoleRequest {
  role: OrganizationRole;
}

export interface CreateInvitationRequest {
  organizationId: string;
  email: string;
  role: OrganizationRole;
  notes?: string;
}
