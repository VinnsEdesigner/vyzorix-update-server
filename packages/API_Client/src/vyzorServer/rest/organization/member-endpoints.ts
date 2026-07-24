/**
 * Organization member REST API endpoints.
 */

import { restClient } from "../_shared/rest-client";
import {
  mapApiToMember,
  mapApiToMemberListItem,
  type MemberApiResponse,
  type OrganizationMember,
  type MemberListItem,
  type CreateMemberRequest,
  type UpdateMemberRoleRequest,
} from "@/domain/organization";

const PATHS = {
  members: (orgId: string) => `/v1/organizations/${orgId}/members`,
  member: (orgId: string, memberId: string) =>
    `/v1/organizations/${orgId}/members/${memberId}`,
} as const;

export interface MemberListParams {
  page?: number;
  limit?: number;
  includeRemoved?: boolean;
}

export const members = {
  /**
   * List members of an organization.
   */
  async list(
    organizationId: string,
    params?: MemberListParams
  ): Promise<MemberListItem[]> {
    const response = await restClient.get<MemberApiResponse[]>(
      PATHS.members(organizationId),
      {
        params: {
          page: params?.page,
          limit: params?.limit,
          include_removed: params?.includeRemoved,
        },
      }
    );
    return response.map(mapApiToMemberListItem);
  },

  /**
   * Get member details.
   */
  async get(
    organizationId: string,
    memberId: string
  ): Promise<OrganizationMember | null> {
    const response = await restClient.get<MemberApiResponse | null>(
      PATHS.member(organizationId, memberId)
    );
    if (!response) return null;
    return mapApiToMember(response);
  },

  /**
   * Add a member to an organization (via invitation).
   */
  async create(
    organizationId: string,
    request: CreateMemberRequest
  ): Promise<{ invitationId: string }> {
    return restClient.post<{ invitation_id: string }>(
      PATHS.members(organizationId),
      request
    ).then((r) => ({ invitationId: r.invitation_id }));
  },

  /**
   * Update member role.
   */
  async updateRole(
    organizationId: string,
    memberId: string,
    request: UpdateMemberRoleRequest
  ): Promise<OrganizationMember> {
    const response = await restClient.patch<MemberApiResponse>(
      PATHS.member(organizationId, memberId),
      request
    );
    return mapApiToMember(response);
  },

  /**
   * Remove a member from an organization.
   */
  async remove(
    organizationId: string,
    memberId: string
  ): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(
      PATHS.member(organizationId, memberId)
    );
  },

  /**
   * Suspend a member.
   */
  async suspend(
    organizationId: string,
    memberId: string
  ): Promise<OrganizationMember> {
    const response = await restClient.post<MemberApiResponse>(
      `${PATHS.member(organizationId, memberId)}/suspend`,
      {}
    );
    return mapApiToMember(response);
  },

  /**
   * Reinstate a suspended member.
   */
  async reinstate(
    organizationId: string,
    memberId: string
  ): Promise<OrganizationMember> {
    const response = await restClient.post<MemberApiResponse>(
      `${PATHS.member(organizationId, memberId)}/reinstate`,
      {}
    );
    return mapApiToMember(response);
  },
};
