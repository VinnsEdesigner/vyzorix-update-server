export type OrganizationRole = "super_admin" | "admin" | "operator" | "viewer";

export type MemberLifecycle = "invited" | "active" | "suspended" | "removed";

export interface Organization {
  id: string;
  name: string;
  description: string;
  createdBy: string;
  maxMembers: number;
  memberCount: number;
  isActive: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface OrganizationMember {
  id: string;
  organizationId: string;
  operatorId: string;
  role: OrganizationRole;
  status: MemberLifecycle;
  invitedBy?: string;
  joinedAt: Date;
  removedAt?: Date;
  operatorName?: string;
  operatorEmail?: string;
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
