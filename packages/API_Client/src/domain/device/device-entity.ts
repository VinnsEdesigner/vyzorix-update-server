export type DeviceStatus = "online" | "offline" | "deregistered";

export interface DeviceConnection {
  webSocketStatus?: string;
  connectedAt?: Date;
  protocol?: string;
  clientIp?: string;
}

export interface Device {
  id: string;
  imei: string;
  organizationId?: string; // Multi-tenant: organization this device belongs to
  deviceName?: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  securityPatch?: string;
  status: DeviceStatus;
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
  organizationId?: string; // Multi-tenant: organization this device belongs to
  deviceName?: string;
  model?: string;
  manufacturer?: string;
  status: DeviceStatus;
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

export interface RegisterDeviceRequest {
  imei: string;
  deviceName?: string;
  model?: string;
  fcmToken: string;
}

export function formatDeviceName(device: Device | DeviceListItem): string {
  return device.deviceName || device.model || `Device ${device.imei.slice(-4)}`;
}

export function isDeviceOnline(device: Device | DeviceListItem): boolean {
  return device.status === "online";
}
