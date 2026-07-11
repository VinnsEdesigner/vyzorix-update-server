import type { InboxEntry, Device, Pagination, InboxStatus, DeviceStatus } from "./registration-entity";

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

function parseTimestamp(value: number | null | undefined): Date | null {
  if (!value) return null;
  return new Date(value > 1e12 ? value : value * 1000);
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