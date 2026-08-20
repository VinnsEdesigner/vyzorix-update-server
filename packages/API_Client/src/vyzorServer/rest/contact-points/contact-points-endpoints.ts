import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  contactPointFromRaw,
  contactPointsFromRaw,
  type ContactPoint,
  type ContactPointRequest,
} from "../../../domain/contact-points";

const PATHS = {
  contactPoints: "/v1/notifications/contact-points",
  contactPoint: (id: string) => `/v1/notifications/contact-points/${id}`,
  testContactPoint: (id: string) => `/v1/notifications/contact-points/${id}/test`,
} as const;

export const contactPoints = {
  async listContactPoints(organizationId?: string): Promise<ContactPoint[]> {
    const response = await restClient.get<{ contact_points: Parameters<typeof contactPointsFromRaw>[0] }>(
      PATHS.contactPoints,
      { params: { organization_id: organizationId || getOrganizationContext() } },
    );
    return contactPointsFromRaw(response.contact_points);
  },

  async getContactPoint(id: string, organizationId?: string): Promise<ContactPoint> {
    const response = await restClient.get<Parameters<typeof contactPointFromRaw>[0]>(PATHS.contactPoint(id), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return contactPointFromRaw(response);
  },

  async createContactPoint(req: ContactPointRequest, organizationId?: string): Promise<ContactPoint> {
    const response = await restClient.post<Parameters<typeof contactPointFromRaw>[0]>(PATHS.contactPoints, req, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return contactPointFromRaw(response);
  },

  async updateContactPoint(id: string, req: ContactPointRequest, organizationId?: string): Promise<ContactPoint> {
    const response = await restClient.patch<Parameters<typeof contactPointFromRaw>[0]>(PATHS.contactPoint(id), req, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return contactPointFromRaw(response);
  },

  async deleteContactPoint(id: string, organizationId?: string): Promise<void> {
    await restClient.delete(PATHS.contactPoint(id), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async testContactPoint(id: string, organizationId?: string): Promise<boolean> {
    const response = await restClient.post<{ sent: boolean }>(PATHS.testContactPoint(id), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return response.sent;
  },
};
