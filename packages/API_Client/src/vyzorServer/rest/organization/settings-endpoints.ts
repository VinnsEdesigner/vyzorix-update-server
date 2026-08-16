import { restClient } from "../_shared/rest-client";
import type { Thresholds } from "../../../domain/settings";

const PATHS = {
  settings: (orgId: string) => `/v1/organizations/${orgId}/settings`,
  thresholds: (orgId: string) => `/v1/organizations/${orgId}/settings/thresholds`,
} as const;

export interface OrganizationSettings {
  id: string;
  organizationId: string;
  timezone: string;
  dateFormat: string;
  alertCooldownMinutes: number;
  defaultThresholds?: Thresholds;
  createdAt: string;
  updatedAt: string;
}

export type ThresholdUpdateRequest = Partial<Thresholds>;

export interface SettingsUpdateRequest {
  timezone?: string;
  dateFormat?: string;
  alertCooldownMinutes?: number;
  defaultThresholds?: Thresholds;
}

export const settings = {
  async get(orgId: string): Promise<OrganizationSettings> {
    return restClient.get<OrganizationSettings>(PATHS.settings(orgId));
  },

  async update(orgId: string, request: SettingsUpdateRequest): Promise<OrganizationSettings> {
    return restClient.patch<OrganizationSettings>(PATHS.settings(orgId), request);
  },

  async getThresholds(orgId: string): Promise<{ thresholds: Thresholds }> {
    return restClient.get<{ thresholds: Thresholds }>(PATHS.thresholds(orgId));
  },

  async updateThresholds(orgId: string, request: ThresholdUpdateRequest): Promise<{ thresholds: Thresholds }> {
    return restClient.patch<{ thresholds: Thresholds }>(PATHS.thresholds(orgId), request);
  },
};
