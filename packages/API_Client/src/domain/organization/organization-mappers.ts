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
    created_by: raw.created_by,
    max_members: raw.max_members,
    member_count: raw.member_count,
    is_active: raw.is_active,
    created_at: raw.created_at,
    updated_at: raw.updated_at,
  };
}

export function mapMember(raw: MemberApiResponse): OrganizationMember {
  return {
    id: raw.id,
    organization_id: raw.organization_id,
    operator_id: raw.operator_id,
    role: raw.role as OrganizationRole,
    status: raw.status as MemberLifecycle,
    invited_by: raw.invited_by,
    joined_at: raw.joined_at,
    removed_at: raw.removed_at,
    operator_name: raw.operator_name,
    operator_email: raw.operator_email,
  };
}
