export type InboxStatus = "pending" | "acknowledged" | "approving" | "approved" | "rejected" | "expired";
export type AcknowledgeAction = "acknowledge" | "approve" | "reject";
export type DeviceStatus = "online" | "offline" | "deregistered";


export interface CreateInboxRequest {
  imei: string;
  deviceName?: string;
  deviceClass?: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  fcmToken: string;
  firebaseInstallId: string;
  idempotencyKey?: string;
}


export interface CreateInboxResult {
  id: string;
  imei: string;
  status: InboxStatus;
  createdAt: Date;
}


export interface ConfirmDeviceResult {
  deviceId: string;
  imei: string;
  confirmed: boolean;
  online: boolean;
  registeredAt: Date;
  serverTime: Date;
}

export interface InboxEntry {
  id: string;
  imei: string;
  deviceName: string;
  deviceClass: string;
  model: string;
  manufacturer: string;
  osVersion: string;
  appVersion: string;
  fcmToken: string;
  firebaseInstallId: string;
  status: InboxStatus;
  acknowledgedAt: Date | null;
  approvingAt: Date | null;
  approvedAt: Date | null;
  rejectedAt: Date | null;
  notes: string | null;
  operatorId: string | null;
  createdAt: Date;
}

export interface Device {
  id: string;
  imei: string;
  deviceName: string;
  model: string;
  manufacturer: string;
  osVersion: string;
  appVersion: string;
  status: DeviceStatus;
  registeredAt: Date | null;
  lastSeen: Date | null;
  online: boolean;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface InboxListResult {
  requests: InboxEntry[];
  pagination: Pagination;
}

export interface DeviceListResult {
  devices: Device[];
  pagination: Pagination;
}

export interface AckResult {
  id: string;
  imei: string;
  status: InboxStatus;
  acknowledgedAt: Date | null;
  approvingAt: Date | null;
  approvedAt: Date | null;
  rejectedAt: Date | null;
  commandSecret: string | null;
  fcmPushSent: boolean;
  notes: string | null;
}

export interface DeregisterResult {
  imei: string;
  status: string;
  deregisteredAt: Date;
  retentionUntil: Date;
}