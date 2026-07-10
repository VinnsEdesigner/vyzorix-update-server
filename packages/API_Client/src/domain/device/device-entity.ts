export type DeviceOnlineStatus = "online" | "offline";

export type DeviceRegistrationStatus = "pending" | "registered" | "deregistered" | "suspended";

export interface DeviceStatus {
  online: DeviceOnlineStatus;
  registration: DeviceRegistrationStatus;
}

export interface DeviceConnection {
  webSocketStatus?: string;
  connectedAt?: Date;
  protocol?: string;
  clientIp?: string;
}

export interface Device {
  id: string;
  imei: string;
  deviceName?: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  securityPatch?: string;
  status: DeviceOnlineStatus;
  registeredAt?: Date;
  lastSeen?: Date;
  fcmTokenValid: boolean;
  commandSecretSet: boolean;
  connection: DeviceConnection;
  createdAt: Date;
  updatedAt: Date;
}

export interface DeviceListItem {
  id: string;
  imei: string;
  deviceName?: string;
  model?: string;
  manufacturer?: string;
  status: DeviceOnlineStatus;
  lastSeen?: Date;
}

export interface DeviceStats {
  total: number;
  online: number;
  offline: number;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface DeviceListResult {
  devices: DeviceListItem[];
  pagination: Pagination;
}

export function isDeviceOnline(device: Device | DeviceListItem): boolean {
  return device.status === "online";
}

export function isDeviceRegistered(device: Device): boolean {
  return device.registeredAt !== undefined;
}

export function formatDeviceName(device: Device | DeviceListItem): string {
  return device.deviceName || device.model || `Device ${device.imei.slice(-4)}`;
}
