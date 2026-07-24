import { restClient } from "../_shared/rest-client";
import type { OrganizationMember, UpdateMemberRoleRequest } from "@/domain/organization";
import { mapMember, type MemberApiResponse } from "@/domain/organization";

const PATHS = {
  members: (orgId: string) => `/v1/organizations/${orgId}/members`,
  member: (orgId: string, memberId: string) => `/v1/organizations/${orgId}/members/${memberId}`,
  suspend: (orgId: string, memberId: string) => `/v1/organizations/${orgId}/members/${memberId}/suspend`,
  reinstate: (orgId: string, memberId: string) => `/v1/organizations/${orgId}/members/${memberId}/reinstate`,
  transfer: (orgId: string, memberId: string) => `/v1/organizations/${orgId}/members/${memberId}/transfer`,
};

export const members = {
  async list(orgId: string): Promise<OrganizationMember[]> {
    const response = await restClient.get<{ members: MemberApiResponse[] }>(PATHS.members(orgId));
    return response.members.map(mapMember);
  },

  async get(orgId: string, memberId: string): Promise<OrganizationMember> {
    const response = await restClient.get<MemberApiResponse>(PATHS.member(orgId, memberId));
    return mapMember(response);
  },

  async updateRole(orgId: string, memberId: string, request: UpdateMemberRoleRequest): Promise<OrganizationMember> {
    const response = await restClient.patch<MemberApiResponse>(PATHS.member(orgId, memberId), request);
    return mapMember(response);
  },

  async remove(orgId: string, memberId: string): Promise<{ message: string }> {
    return restClient.delete<{ message: string }>(PATHS.member(orgId, memberId));
  },

  async suspend(orgId: string, memberId: string): Promise<{ message: string }> {
    return restClient.post<{ message: string }>(PATHS.suspend(orgId, memberId), {});
  },

  async reinstate(orgId: string, memberId: string): Promise<{ message: string }> {
    return restClient.post<{ message: string }>(PATHS.reinstate(orgId, memberId), {});
  },

  async transferOwnership(orgId: string, memberId: string): Promise<{ message: string }> {
    return restClient.post<{ message: string }>(PATHS.transfer(orgId, memberId), {});
  },
};
