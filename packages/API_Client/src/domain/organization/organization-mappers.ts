import type {
  Organization,
  OrganizationMember,
  OrganizationRole,
  MemberLifecycle,
} from "./organization-entity";

export interface OrganizationApiResponse {
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

export interface MemberApiResponse {
  id: string;
  organization_id: string;
  operator_id: string;
  role: string;
  status: string;
  invited_by?: string;
  joined_at: string;
  removed_at?: string;
  operator_name?: string;
  operator_email?: string;
}

export function mapOrganization(raw: OrganizationApiResponse): Organization {
  return {
    id: raw.id,
    name: raw.name,
    description: raw.description,
    createdBy: raw.created_by,
    maxMembers: raw.max_members,
    memberCount: raw.member_count,
    isActive: raw.is_active,
    createdAt: new Date(raw.created_at),
    updatedAt: new Date(raw.updated_at),
  };
}

export function mapMember(raw: MemberApiResponse): OrganizationMember {
  return {
    id: raw.id,
    organizationId: raw.organization_id,
    operatorId: raw.operator_id,
    role: raw.role as OrganizationRole,
    status: raw.status as MemberLifecycle,
    invitedBy: raw.invited_by,
    joinedAt: new Date(raw.joined_at),
    removedAt: raw.removed_at ? new Date(raw.removed_at) : undefined,
    operatorName: raw.operator_name,
    operatorEmail: raw.operator_email,
  };
}

export function organizationToRaw(org: Partial<Organization>): Partial<OrganizationApiResponse> {
  return {
    name: org.name,
    description: org.description,
    max_members: org.maxMembers,
    is_active: org.isActive,
  };
}
