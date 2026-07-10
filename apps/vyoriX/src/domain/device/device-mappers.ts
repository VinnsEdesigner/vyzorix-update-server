import type { Device, DeviceListItem, DeviceStats, DeviceOnlineStatus, DeviceConnection } from "./device-entity";

export interface RawDevice {
  id?: string;
  imei?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  app_version?: string;
  os_version?: string;
  security_patch?: string;
  build_id?: string;
  online?: boolean;
  fcm_token_valid?: boolean;
  command_secret_set?: boolean;
  registered_at?: number;
  last_seen?: number;
  created_at?: number;
  updated_at?: number;
}

export interface RawDeviceListItem {
  id?: string;
  imei?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  status?: string;
  last_seen?: number;
}

export interface RawDeviceStats {
  total?: number;
  online?: number;
  offline?: number;
}

export interface RawDeviceConnection {
  web_socket_status?: string;
  connected_at?: number;
  protocol?: string;
  client_ip?: string;
}

function parseTimestamp(value?: number | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value > 1e12 ? value : value * 1000);
}

export function deviceFromRaw(raw: RawDevice): Device {
  return {
    id: raw.id ?? "",
    imei: raw.imei ?? "",
    deviceName: raw.device_name,
    model: raw.model,
    manufacturer: raw.manufacturer,
    osVersion: raw.os_version,
    appVersion: raw.app_version,
    securityPatch: raw.security_patch,
    status: (raw.online ? "online" : "offline") as DeviceOnlineStatus,
    registeredAt: parseTimestamp(raw.registered_at),
    lastSeen: parseTimestamp(raw.last_seen),
    fcmTokenValid: raw.fcm_token_valid ?? false,
    commandSecretSet: raw.command_secret_set ?? false,
    connection: {} as DeviceConnection,
    createdAt: parseTimestamp(raw.created_at) ?? new Date(),
    updatedAt: parseTimestamp(raw.updated_at) ?? new Date(),
  };
}

export function deviceListItemFromRaw(raw: RawDeviceListItem): DeviceListItem {
  return {
    id: raw.id ?? "",
    imei: raw.imei ?? "",
    deviceName: raw.device_name,
    model: raw.model,
    manufacturer: raw.manufacturer,
    status: (raw.status as DeviceOnlineStatus) ?? "offline",
    lastSeen: parseTimestamp(raw.last_seen),
  };
}

export function deviceStatsFromRaw(raw: RawDeviceStats): DeviceStats {
  return {
    total: raw.total ?? 0,
    online: raw.online ?? 0,
    offline: raw.offline ?? 0,
  };
}

export function deviceListItemsFromRaw(raw: RawDeviceListItem[]): DeviceListItem[] {
  return raw.map(deviceListItemFromRaw);
}
