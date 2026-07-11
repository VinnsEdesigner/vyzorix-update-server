import { restClient } from "../_shared/rest-client";
import {
  deviceFromRaw,
  deviceListItemFromRaw,
  deviceStatsFromRaw,
  paginationFromRaw,
  type RawDevice,
  type RawDeviceListItem,
  type RawDeviceStats,
  type RawPagination,
  type RawDeviceListResult,
} from "@/domain/device";
import type {
  Device,
  DeviceListItem,
  DeviceStats,
  DeviceListResult,
  RegisterDeviceRequest,
} from "@/domain/device";

const PATHS = {
  devices: "/v1/devices",
  stats: "/v1/devices/stats",
} as const;

export interface DeviceParams {
  page?: number;
  limit?: number;
  status?: "online" | "offline" | "all";
}

export const devices = {
  async list(params?: DeviceParams): Promise<DeviceListResult> {
    const response = await restClient.get<RawDeviceListResult>(PATHS.devices, {
      params: {
        page: params?.page,
        limit: params?.limit,
        status: params?.status,
      },
    });
    return {
      devices: response.devices.map(deviceListItemFromRaw),
      pagination: paginationFromRaw(response.pagination),
    };
  },

  async get(imei: string): Promise<Device | null> {
    const response = await restClient.get<RawDevice | null>(`${PATHS.devices}/${imei}`);
    if (!response?.imei) return null;
    return deviceFromRaw(response);
  },

  async stats(): Promise<DeviceStats> {
    const response = await restClient.get<RawDeviceStats>(PATHS.stats);
    return deviceStatsFromRaw(response);
  },

  async register(request: RegisterDeviceRequest): Promise<{ device?: Device; commandSecret?: string }> {
    const response = await restClient.post<{
      device?: RawDevice;
      commandSecret?: string;
    }>(PATHS.devices, request);
    return {
      device: response.device ? deviceFromRaw(response.device) : undefined,
      commandSecret: response.commandSecret,
    };
  },

  async deregister(imei: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(`${PATHS.devices}/${imei}`);
  },
};
