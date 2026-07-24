import { restClient } from "../_shared/rest-client";

const PATHS = {
  settings: (orgId: string) => `/v1/organizations/${orgId}/settings`,
  thresholds: (orgId: string) => `/v1/organizations/${orgId}/settings/thresholds`,
};

export interface OrganizationSettings {
  default_thresholds?: {
    temp_min?: number;
    temp_max?: number;
    battery_min?: number;
    battery_max?: number;
    speed_max?: number;
    distance_max?: number;
  };
  updated_at?: string;
}

export interface ThresholdUpdateRequest {
  temp_min?: number;
  temp_max?: number;
  battery_min?: number;
  battery_max?: number;
  speed_max?: number;
  distance_max?: number;
}

export interface SettingsUpdateRequest {
  default_thresholds?: ThresholdUpdateRequest;
}

export const settings = {
  async get(orgId: string): Promise<OrganizationSettings> {
    return restClient.get<OrganizationSettings>(PATHS.settings(orgId));
  },

  async update(orgId: string, request: SettingsUpdateRequest): Promise<OrganizationSettings> {
    return restClient.patch<OrganizationSettings>(PATHS.settings(orgId), request);
  },

  async getThresholds(orgId: string): Promise<{ thresholds: ThresholdUpdateRequest }> {
    return restClient.get<{ thresholds: ThresholdUpdateRequest }>(PATHS.thresholds(orgId));
  },

  async updateThresholds(orgId: string, request: ThresholdUpdateRequest): Promise<{ thresholds: ThresholdUpdateRequest }> {
    return restClient.patch<{ thresholds: ThresholdUpdateRequest }>(PATHS.thresholds(orgId), request);
  },
};
