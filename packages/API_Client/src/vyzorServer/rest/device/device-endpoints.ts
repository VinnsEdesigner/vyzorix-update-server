import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  deviceFromRaw,
  deviceListItemFromRaw,
  deviceStatsFromRaw,
  paginationFromRaw,
  type RawDevice,
  type RawDeviceStats,
  type RawDeviceListResult,
} from "@/domain/device";
import type {
  Device,
  DeviceStats,
  DeviceListResult,
} from "@/domain/device";

const PATHS = {
  devices: "/v1/devices",
  device: (imei: string) => `/v1/devices/${imei}`,
  stats: "/v1/devices/stats",
  count: "/v1/devices/count",
  stream: (imei: string) => `/v1/devices/${imei}/stream`,
  connectionStatus: (imei: string) => `/v1/devices/${imei}/connection-status`,
  disconnect: (imei: string) => `/v1/devices/${imei}/disconnect`,
  fcmToken: (imei: string) => `/v1/devices/${imei}/fcm-token`,
  settings: (imei: string) => `/v1/devices/${imei}/settings`,
  settingsThresholds: (imei: string) => `/v1/devices/${imei}/settings/thresholds`,
} as const;

export interface DeviceParams {
  page?: number;
  limit?: number;
  status?: "online" | "offline" | "all";
  search?: string;
  organizationId?: string;
}

export interface DeviceSettings {
  id: string;
  imei: string;
  orgId: string;
  fcmEnabled: boolean;
  fcmToken?: string;
  thresholds?: {
    tempMin?: number;
    tempMax?: number;
    batteryMin?: number;
    batteryMax?: number;
    speedMax?: number;
    distanceMax?: number;
  };
  createdAt: string;
  updatedAt: string;
}

export interface ConnectionStatus {
  imei: string;
  connected: boolean;
  connectionId?: string;
  connectedAt?: number;
  lastActivityAt?: number;
  ipAddress?: string;
  userAgent?: string;
}

export const devices = {
  async list(params?: DeviceParams): Promise<DeviceListResult> {
    const response = await restClient.get<RawDeviceListResult>(PATHS.devices, {
      params: {
        page: params?.page,
        limit: params?.limit,
        status: params?.status,
        search: params?.search,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return {
      devices: response.devices.map(deviceListItemFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async get(imei: string, organizationId?: string): Promise<Device | null> {
    const response = await restClient.get<RawDevice | null>(PATHS.device(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    if (!response?.imei) return null;
    return deviceFromRaw(response);
  },

  async count(organizationId?: string): Promise<{ count: number }> {
    return restClient.get<{ count: number }>(PATHS.count, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async stats(organizationId?: string): Promise<DeviceStats> {
    const response = await restClient.get<RawDeviceStats>(PATHS.stats, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return deviceStatsFromRaw(response);
  },

  async getStream(imei: string, organizationId?: string): Promise<string> {
    return restClient.get<string>(PATHS.stream(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async getConnectionStatus(imei: string, organizationId?: string): Promise<ConnectionStatus> {
    return restClient.get<ConnectionStatus>(PATHS.connectionStatus(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async disconnect(imei: string, organizationId?: string): Promise<{ success: boolean }> {
    return restClient.post<{ success: boolean }>(PATHS.disconnect(imei), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async updateFcmToken(imei: string, fcmToken: string, organizationId?: string): Promise<{ success: boolean }> {
    return restClient.patch<{ success: boolean }>(PATHS.fcmToken(imei), { fcmToken }, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async deregister(imei: string, organizationId?: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(PATHS.device(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async getSettings(imei: string, organizationId?: string): Promise<DeviceSettings> {
    return restClient.get<DeviceSettings>(PATHS.settings(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async updateSettings(imei: string, settings: Partial<DeviceSettings>, organizationId?: string): Promise<DeviceSettings> {
    return restClient.patch<DeviceSettings>(PATHS.settings(imei), settings, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async getSettingsThresholds(imei: string, organizationId?: string): Promise<{ thresholds: DeviceSettings['thresholds'] }> {
    return restClient.get<{ thresholds: DeviceSettings['thresholds'] }>(PATHS.settingsThresholds(imei), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async updateSettingsThresholds(imei: string, thresholds: DeviceSettings['thresholds'], organizationId?: string): Promise<{ thresholds: DeviceSettings['thresholds'] }> {
    return restClient.patch<{ thresholds: DeviceSettings['thresholds'] }>(PATHS.settingsThresholds(imei), thresholds, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },
};
