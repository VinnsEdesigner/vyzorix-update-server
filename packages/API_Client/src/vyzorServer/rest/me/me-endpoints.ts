import { restClient } from "../_shared/rest-client";
import type { MeResponse, OrganizationInfo } from "@/domain/auth/auth-entity";

const PATHS = {
  me: "/v1/auth/me",
  organizations: "/v1/auth/organizations",
  selectOrganization: "/v1/auth/organizations/select",
};

export interface SelectOrganizationRequest {
  organization_id: string;
}

export const me = {
  async getMe(): Promise<MeResponse> {
    return restClient.get<MeResponse>(PATHS.me);
  },

  async getOrganizations(): Promise<{ organizations: OrganizationInfo[] }> {
    return restClient.get<{ organizations: OrganizationInfo[] }>(PATHS.organizations);
  },

  async selectOrganization(request: SelectOrganizationRequest): Promise<OrganizationInfo> {
    return restClient.post<OrganizationInfo>(PATHS.selectOrganization, request);
  },
};
