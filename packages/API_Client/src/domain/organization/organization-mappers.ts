/**
 * Mappers for converting API responses to domain entities.
 */

import type {
  Organization,
  OrganizationListItem,
  OrganizationMember,
  MemberListItem,
  OrganizationApiResponse,
  MemberApiResponse,
  OrganizationRole,
  MemberLifecycle,
  OrganizationLifecycle,
} from "./organization-entity";

/**
 * Map API organization response to domain entity.
 */
export function mapApiToOrganization(
  api: OrganizationApiResponse
): Organization {
  return {
    id: api.id,
    name: api.name,
    description: api.description,
    createdBy: api.created_by,
    maxMembers: api.max_members,
    memberCount: api.member_count,
    lifecycle: api.lifecycle as OrganizationLifecycle,
    createdAt: new Date(api.created_at),
    updatedAt: new Date(api.updated_at),
    deletedAt: api.deleted_at ? new Date(api.deleted_at) : undefined,
  };
}

/**
 * Map API response to organization list item (lightweight).
 */
export function mapApiToOrganizationListItem(
  api: OrganizationApiResponse
): OrganizationListItem {
  return {
    id: api.id,
    name: api.name,
    description: api.description,
    memberCount: api.member_count,
    lifecycle: api.lifecycle as OrganizationLifecycle,
    createdAt: new Date(api.created_at),
  };
}

/**
 * Map API member response to domain entity.
 */
export function mapApiToMember(api: MemberApiResponse): OrganizationMember {
  return {
    id: api.id,
    organizationId: api.organization_id,
    operatorId: api.operator_id,
    role: api.role as OrganizationRole,
    lifecycle: api.lifecycle as MemberLifecycle,
    invitedBy: api.invited_by,
    joinedAt: new Date(api.joined_at),
    removedAt: api.removed_at ? new Date(api.removed_at) : undefined,
    suspendedAt: api.suspended_at ? new Date(api.suspended_at) : undefined,
    operatorName: api.operator_name,
    operatorEmail: api.operator_email,
  };
}

/**
 * Map API response to member list item (lightweight).
 */
export function mapApiToMemberListItem(
  api: MemberApiResponse
): MemberListItem {
  return {
    id: api.id,
    operatorId: api.operator_id,
    operatorName: api.operator_name ?? "Unknown",
    operatorEmail: api.operator_email ?? "",
    role: api.role as OrganizationRole,
    lifecycle: api.lifecycle as MemberLifecycle,
    joinedAt: new Date(api.joined_at),
  };
}

/**
 * Map array of API responses to organization list.
 */
export function mapApiToOrganizationList(
  apiList: OrganizationApiResponse[]
): OrganizationListItem[] {
  return apiList.map(mapApiToOrganizationListItem);
}

/**
 * Map array of API responses to member list.
 */
export function mapApiToMemberList(apiList: MemberApiResponse[]): MemberListItem[] {
  return apiList.map(mapApiToMemberListItem);
}
