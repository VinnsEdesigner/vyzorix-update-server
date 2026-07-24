export type DeviceStatus = "online" | "offline" | "deregistered";

export interface DeviceConnection {
  web_socket_status?: string;
  connected_at?: string;
  protocol?: string;
  client_ip?: string;
}

export interface Device {
  id: string;
  imei: string;
  organization_id?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  os_version?: string;
  app_version?: string;
  security_patch?: string;
  status: DeviceStatus;
  registered_at?: string;
  last_seen?: string;
  fcm_token_valid: boolean;
  command_secret_set: boolean;
  connection: DeviceConnection;
  created_at: string;
  updated_at: string;
}

export interface DeviceListItem {
  id: string;
  imei: string;
  organization_id?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  status: DeviceStatus;
  last_seen?: string;
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
  total_pages: number;
}

export interface DeviceListResult {
  devices: DeviceListItem[];
  pagination: Pagination;
}

export interface RegisterDeviceRequest {
  imei: string;
  device_name?: string;
  model?: string;
  fcm_token: string;
}

export function formatDeviceName(device: Device | DeviceListItem): string {
  return device.device_name || device.model || `Device ${device.imei.slice(-4)}`;
}

export function isDeviceOnline(device: Device | DeviceListItem): boolean {
  return device.status === "online";
}
