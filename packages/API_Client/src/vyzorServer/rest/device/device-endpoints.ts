import { apiGet, apiPost, apiDelete } from "../_shared/rest-client";
import { deviceFromRaw, deviceListItemFromRaw, deviceStatsFromRaw, deviceListItemsFromRaw, type RawDevice, type RawDeviceListItem } from "@/domain/device/device-mappers";
import type { Device, DeviceListItem, DeviceStats } from "@/domain/device/device-entity";
import type { PaginatedResult } from "@/domain/_shared";
import { offsetPaginationFromRaw } from "@/domain/_shared";

const BASE = "/v1/devices";

export const DEVICE_PATHS = {
  devices: `${BASE}`,
  device: (imei: string) => `${BASE}/${imei}`,
  deviceStats: `${BASE}/stats`,
} as const;

export interface RegisterDeviceRequest {
  imei: string;
  deviceName?: string;
  model?: string;
  fcmToken: string;
}

export interface RegisterDeviceResponse {
  success: boolean;
  device?: RawDevice;
  commandSecret?: string;
}

export async function fetchDevices(
  params?: { page?: number; limit?: number; status?: "online" | "offline" | "all" }
): Promise<PaginatedResult<DeviceListItem[]>> {
  const response = await apiGet<{
    devices: RawDeviceListItem[];
    pagination: { page: number; limit: number; total: number; total_pages: number };
  }>(DEVICE_PATHS.devices, {
    page: params?.page,
    limit: params?.limit,
    status: params?.status,
  });
  
  return {
    items: deviceListItemsFromRaw(response.devices),
    pagination: offsetPaginationFromRaw(response.pagination),
  };
}

export async function fetchDevice(imei: string): Promise<Device | null> {
  const data = await apiGet<RawDevice | null>(DEVICE_PATHS.device(imei));
  if (!data || !data.imei) return null;
  return deviceFromRaw(data);
}

export async function fetchDeviceStats(): Promise<DeviceStats> {
  const data = await apiGet<{ total: number; online: number; offline: number }>(DEVICE_PATHS.deviceStats);
  return deviceStatsFromRaw(data);
}

export async function registerDevice(request: RegisterDeviceRequest): Promise<{ success: boolean; device?: Device; commandSecret?: string }> {
  const data = await apiPost<RegisterDeviceResponse>(DEVICE_PATHS.device, request);
  return {
    success: data.success,
    device: data.device ? deviceFromRaw(data.device) : undefined,
    commandSecret: data.commandSecret,
  };
}

export async function deregisterDevice(imei: string): Promise<{ success: boolean }> {
  return apiDelete<{ success: boolean }>(DEVICE_PATHS.device(imei));
}
