import type {
  Device,
  DeviceListItem,
  DeviceStats,
  DeviceStatus,
  DeviceConnection,
} from "./device-entity";
import type { RawPagination } from "../_shared";

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
  organization_id?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  status: string;
  last_seen?: number;
  registered_at?: number;
}

export interface RawDeviceStats {
  total: number;
  online: number;
  offline: number;
}

export interface RawDeviceListResult {
  devices: RawDeviceListItem[];
  pagination: RawPagination;
}

function statusFromOnline(online: boolean, deregisteredAt?: number): DeviceStatus {
  if (deregisteredAt) return "deregistered";
  return online ? "online" : "offline";
}

export function deviceFromRaw(raw: RawDevice): Device {
  return {
    id: raw.id,
    imei: raw.imei,
    device_name: raw.deviceName,
    model: raw.model,
    manufacturer: raw.manufacturer,
    os_version: raw.osVersion,
    app_version: raw.appVersion,
    security_patch: raw.securityPatch,
    status: statusFromOnline(raw.online),
    registered_at: raw.registeredAt ? new Date(raw.registeredAt).toISOString() : undefined,
    last_seen: raw.lastSeen ? new Date(raw.lastSeen).toISOString() : undefined,
    fcm_token_valid: raw.fcmTokenValid,
    command_secret_set: raw.commandSecretSet,
    connection: {} as DeviceConnection,
    created_at: new Date(raw.createdAt).toISOString(),
    updated_at: new Date(raw.updatedAt).toISOString(),
  };
}

export function deviceListItemFromRaw(raw: RawDeviceListItem): DeviceListItem {
  return {
    id: raw.id,
    imei: raw.imei,
    organization_id: raw.organization_id,
    device_name: raw.device_name,
    model: raw.model,
    manufacturer: raw.manufacturer,
    status: (raw.status as DeviceStatus) ?? "offline",
    last_seen: raw.last_seen ? new Date(raw.last_seen).toISOString() : undefined,
  };
}

export function deviceStatsFromRaw(raw: RawDeviceStats): DeviceStats {
  return {
    total: raw.total ?? 0,
    online: raw.online ?? 0,
    offline: raw.offline ?? 0,
  };
}
