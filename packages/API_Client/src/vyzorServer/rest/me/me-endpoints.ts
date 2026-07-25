
import { restClient } from "../_shared/rest-client";
import { meResponseFromRaw, type MeResponse, type OrganizationInfo, type OrganizationMembership } from "@/domain/auth";
import type { Invitation, InvitationApiResponse } from "@/domain/invitation";
import { mapInvitation } from "@/domain/invitation";

const PATHS = {
  me: "/v1/auth/me",
  organizations: "/v1/auth/organizations",
  invitations: "/v1/invitations",
  selectOrganization: "/v1/auth/organizations/select",
};

export interface SelectOrganizationRequest {
  organization_id: string;
}

export const me = {
  async getMe(): Promise<MeResponse> {
    const raw = await restClient.get<{
      id: string;
      email: string;
      name: string;
      mfa_enabled: boolean;
      email_verified: boolean;
      thresholds?: unknown;
      client?: unknown;
      needs_organization: boolean;
      organizations: OrganizationInfo[];
      memberships?: OrganizationMembership[];
      last_organization_id?: string;
      selected_organization?: OrganizationInfo;
    }>(PATHS.me);
    return meResponseFromRaw(raw as any);
  },

  async getOrganizations(): Promise<{ organizations: OrganizationInfo[] }> {
    return restClient.get<{ organizations: OrganizationInfo[] }>(PATHS.organizations);
  },

  async getInvitations(): Promise<Invitation[]> {
    const response = await restClient.get<InvitationApiResponse[]>(PATHS.invitations);
    return response.map(mapInvitation);
  },

  async selectOrganization(request: SelectOrganizationRequest): Promise<OrganizationInfo> {
    return restClient.post<OrganizationInfo>(PATHS.selectOrganization, request);
  },
};
