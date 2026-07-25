import { restClient } from "../_shared/rest-client";
import type { Organization, CreateOrganizationRequest, UpdateOrganizationRequest } from "@/domain/organization";
import { mapOrganization, type OrganizationApiResponse } from "@/domain/organization";

const PATHS = {
  organizations: "/v1/organizations",
  organization: (id: string) => `/v1/organizations/${id}`,
  transferDevice: (orgId: string, imei: string) => `/v1/organizations/${orgId}/devices/${imei}/transfer`,
};

export interface TransferDeviceRequest {
  targetOrganizationId: string;
  reason?: string;
}

export const organizations = {
  async list(): Promise<Organization[]> {
    const response = await restClient.get<{ organizations: OrganizationApiResponse[] }>(PATHS.organizations);
    return response.organizations.map(mapOrganization);
  },

  async create(request: CreateOrganizationRequest): Promise<Organization> {
    const response = await restClient.post<OrganizationApiResponse>(PATHS.organizations, request);
    return mapOrganization(response);
  },

  async get(id: string): Promise<Organization> {
    const response = await restClient.get<OrganizationApiResponse>(PATHS.organization(id));
    return mapOrganization(response);
  },

  async update(id: string, request: UpdateOrganizationRequest): Promise<Organization> {
    const response = await restClient.patch<OrganizationApiResponse>(PATHS.organization(id), request);
    return mapOrganization(response);
  },

  async delete(id: string): Promise<{ message: string }> {
    return restClient.delete<{ message: string }>(PATHS.organization(id));
  },

  async transferDevice(orgId: string, imei: string, request: TransferDeviceRequest): Promise<{ success: boolean; message: string }> {
    return restClient.post<{ success: boolean; message: string }>(PATHS.transferDevice(orgId, imei), request);
  },
};
