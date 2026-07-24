import type {
  InboxEntry,
  Device,
  Pagination,
  InboxStatus,
  DeviceStatus,
  CreateInboxRequest,
  CreateInboxResult,
  ConfirmDeviceResult,
} from "./registration-entity";

export interface RawInboxEntry {
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
  acknowledgedAt: number | null;
  approvingAt: number | null;
  approvedAt: number | null;
  rejectedAt: number | null;
  notes: string | null;
  operatorId: string | null;
  createdAt: number;
}

export interface RawDevice {
  id: string;
  imei: string;
  deviceName: string;
  model: string;
  manufacturer: string;
  osVersion: string;
  appVersion: string;
  status: DeviceStatus;
  registeredAt: number | null;
  lastSeen: number | null;
  online: boolean;
}

export interface RawPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}


export interface RawCreateInboxRequest {
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


export interface RawCreateInboxResponse {
  id: string;
  imei: string;
  status: InboxStatus;
  createdAt: number;
}


export interface RawConfirmDeviceResponse {
  device_id: string;
  imei: string;
  confirmed: boolean;
  online: boolean;
  registered_at: number;
  server_time: number;
}

function parseTimestamp(value: number | null | undefined): Date | null {
  if (!value) return null;
  return new Date(value > 1e12 ? value : value * 1000);
}


export function createInboxRequestToRaw(request: CreateInboxRequest): RawCreateInboxRequest {
  return {
    imei: request.imei,
    deviceName: request.deviceName,
    deviceClass: request.deviceClass,
    model: request.model,
    manufacturer: request.manufacturer,
    osVersion: request.osVersion,
    appVersion: request.appVersion,
    fcmToken: request.fcmToken,
    firebaseInstallId: request.firebaseInstallId,
    idempotencyKey: request.idempotencyKey,
  };
}


export function createInboxResultFromRaw(raw: RawCreateInboxResponse): CreateInboxResult {
  return {
    id: raw.id,
    imei: raw.imei,
    status: raw.status,
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
  };
}


export function confirmDeviceResultFromRaw(raw: RawConfirmDeviceResponse): ConfirmDeviceResult {
  return {
    deviceId: raw.device_id,
    imei: raw.imei,
    confirmed: raw.confirmed,
    online: raw.online,
    registeredAt: parseTimestamp(raw.registered_at) ?? new Date(),
    serverTime: parseTimestamp(raw.server_time) ?? new Date(),
  };
}

export function inboxEntryFromRaw(raw: RawInboxEntry): InboxEntry {
  return {
    id: raw.id,
    imei: raw.imei,
    deviceName: raw.deviceName,
    deviceClass: raw.deviceClass,
    model: raw.model,
    manufacturer: raw.manufacturer,
    osVersion: raw.osVersion,
    appVersion: raw.appVersion,
    fcmToken: raw.fcmToken,
    firebaseInstallId: raw.firebaseInstallId,
    status: raw.status,
    acknowledgedAt: parseTimestamp(raw.acknowledgedAt),
    approvingAt: parseTimestamp(raw.approvingAt),
    approvedAt: parseTimestamp(raw.approvedAt),
    rejectedAt: parseTimestamp(raw.rejectedAt),
    notes: raw.notes,
    operatorId: raw.operatorId,
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
  };
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
    status: raw.status,
    registeredAt: parseTimestamp(raw.registeredAt),
    lastSeen: parseTimestamp(raw.lastSeen),
    online: raw.online,
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