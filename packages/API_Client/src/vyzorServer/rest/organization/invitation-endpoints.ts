import { restClient } from "../_shared/rest-client";
import type { Invitation, InvitationResponseRequest } from "@/domain/invitation";
import { mapInvitation, type InvitationApiResponse } from "@/domain/invitation";
import type { OrganizationRole } from "@/domain/organization";

interface CreateInvitationRequest {
  email: string;
  role: OrganizationRole;
}

const PATHS = {
  invitations: "/v1/invitations",
  orgInvitations: (orgId: string) => `/v1/organizations/${orgId}/invitations`,
  invitationByToken: (token: string) => `/v1/invite/${token}`,
  accept: (token: string) => `/v1/invite/${token}/accept`,
  reject: (token: string) => `/v1/invite/${token}/reject`,
};

export const invitations = {
  async create(request: CreateInvitationRequest): Promise<Invitation> {
    const response = await restClient.post<InvitationApiResponse>(PATHS.invitations, request);
    return mapInvitation(response);
  },

  async listByOrganization(orgId: string, status?: string): Promise<Invitation[]> {
    const url = status ? `${PATHS.orgInvitations(orgId)}?status=${status}` : PATHS.orgInvitations(orgId);
    const response = await restClient.get<{ invitations: InvitationApiResponse[] }>(url);
    return response.invitations.map(mapInvitation);
  },

  async getByToken(token: string): Promise<Invitation> {
    const response = await restClient.get<InvitationApiResponse>(PATHS.invitationByToken(token));
    return mapInvitation(response);
  },

  async accept(token: string, request?: InvitationResponseRequest): Promise<{ message: string }> {
    return restClient.post<{ message: string }>(PATHS.accept(token), request ?? {});
  },

  async reject(token: string, request?: InvitationResponseRequest): Promise<{ message: string }> {
    return restClient.post<{ message: string }>(PATHS.reject(token), request ?? {});
  },
};
