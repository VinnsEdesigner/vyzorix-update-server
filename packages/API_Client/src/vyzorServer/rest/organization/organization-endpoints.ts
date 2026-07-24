/**
 * Organization REST API endpoints.
 */

import { restClient } from "../_shared/rest-client";
import {
  mapApiToOrganization,
  mapApiToOrganizationListItem,
  type OrganizationApiResponse,
  type Organization,
  type OrganizationListItem,
  type CreateOrganizationRequest,
  type UpdateOrganizationRequest,
} from "@/domain/organization";

const PATHS = {
  organizations: "/v1/organizations",
  organization: (id: string) => `/v1/organizations/${id}`,
} as const;

export interface OrganizationListParams {
  page?: number;
  limit?: number;
  includeArchived?: boolean;
}

export const organizations = {
  /**
   * Create a new organization.
   */
  async create(request: CreateOrganizationRequest): Promise<Organization> {
    const response = await restClient.post<OrganizationApiResponse>(
      PATHS.organizations,
      request
    );
    return mapApiToOrganization(response);
  },

  /**
   * List organizations the current user is a member of.
   */
  async list(params?: OrganizationListParams): Promise<OrganizationListItem[]> {
    const response = await restClient.get<OrganizationApiResponse[]>(
      PATHS.organizations,
      {
        params: {
          page: params?.page,
          limit: params?.limit,
          include_archived: params?.includeArchived,
        },
      }
    );
    return response.map(mapApiToOrganizationListItem);
  },

  /**
   * Get organization details by ID.
   */
  async get(id: string): Promise<Organization | null> {
    const response = await restClient.get<OrganizationApiResponse | null>(
      PATHS.organization(id)
    );
    if (!response) return null;
    return mapApiToOrganization(response);
  },

  /**
   * Update organization details.
   */
  async update(
    id: string,
    request: UpdateOrganizationRequest
  ): Promise<Organization> {
    const response = await restClient.patch<OrganizationApiResponse>(
      PATHS.organization(id),
      request
    );
    return mapApiToOrganization(response);
  },

  /**
   * Delete (archive) an organization.
   */
  async delete(id: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(PATHS.organization(id));
  },
};
