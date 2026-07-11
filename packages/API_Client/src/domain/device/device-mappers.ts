import type {
  Device,
  DeviceListItem,
  DeviceStats,
  DeviceStatus,
  DeviceConnection,
} from "./device-entity";

export interface RawDevice {
  id: string;
  imei: string;
  deviceName?: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  securityPatch?: string;
  online: boolean;
  fcmTokenValid: boolean;
  commandSecretSet: boolean;
  registeredAt?: number;
  lastSeen?: number;
  createdAt: number;
  updatedAt: number;
}

export interface RawDeviceListItem {
  id: string;
  imei: string;
  deviceName?: string;
  model?: string;
  manufacturer?: string;
  status: string;
  lastSeen?: number;
}

export interface RawDeviceStats {
  total: number;
  online: number;
  offline: number;
}

export interface RawPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface RawDeviceListResult {
  devices: RawDeviceListItem[];
  pagination: RawPagination;
}

function parseTimestamp(value?: number | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value > 1e12 ? value : value * 1000);
}

function statusFromOnline(online: boolean, deregisteredAt?: number): DeviceStatus {
  if (deregisteredAt) return "deregistered";
  return online ? "online" : "offline";
}

export function deviceFromRaw(raw: RawDevice): Device {
  return {
    id: raw.id,
    imei: raw.imei,
    deviceName: raw.deviceName,
    model: raw.model,
    manufacturer: raw.manufacturer,
    osVersion: raw.osVersion,
    appVersion: raw.appVersion,
    securityPatch: raw.securityPatch,
    status: statusFromOnline(raw.online),
    registeredAt: parseTimestamp(raw.registeredAt),
    lastSeen: parseTimestamp(raw.lastSeen),
    fcmTokenValid: raw.fcmTokenValid,
    commandSecretSet: raw.commandSecretSet,
    connection: {} as DeviceConnection,
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
    updatedAt: parseTimestamp(raw.updatedAt) ?? new Date(),
  };
}

export function deviceListItemFromRaw(raw: RawDeviceListItem): DeviceListItem {
  return {
    id: raw.id,
    imei: raw.imei,
    deviceName: raw.deviceName,
    model: raw.model,
    manufacturer: raw.manufacturer,
    status: (raw.status as DeviceStatus) ?? "offline",
    lastSeen: parseTimestamp(raw.lastSeen),
  };
}

export function deviceStatsFromRaw(raw: RawDeviceStats): DeviceStats {
  return {
    total: raw.total ?? 0,
    online: raw.online ?? 0,
    offline: raw.offline ?? 0,
  };
}

export function paginationFromRaw(raw: RawPagination): Pagination {
  return {
    page: raw.page,
    limit: raw.limit,
    total: raw.total,
    totalPages: raw.totalPages,
  };
}
