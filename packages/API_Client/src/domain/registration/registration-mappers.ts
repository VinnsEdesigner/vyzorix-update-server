import type { InboxEntry, InboxListItem, InboxStatus } from "./registration-entity";

export interface RawInboxEntry {
  id?: string;
  imei?: string;
  model?: string;
  manufacturer?: string;
  os_version?: string;
  app_version?: string;
  fcm_token?: string;
  firebase_install_id?: string;
  status?: string;
  created_at?: number;
  updated_at?: number;
  acknowledged_at?: number;
  approved_at?: number;
  rejected_at?: number;
  notes?: string;
  operator_id?: string;
}

export interface RawPagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

function parseTimestamp(value?: number | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value > 1e12 ? value : value * 1000);
}

export function inboxEntryFromRaw(raw: RawInboxEntry): InboxEntry {
  return {
    id: raw.id ?? "",
    imei: raw.imei ?? "",
    model: raw.model,
    manufacturer: raw.manufacturer,
    osVersion: raw.os_version,
    appVersion: raw.app_version,
    fcmToken: raw.fcm_token ?? "",
    firebaseInstallId: raw.firebase_install_id ?? "",
    status: (raw.status as InboxStatus) ?? "pending",
    createdAt: parseTimestamp(raw.created_at) ?? new Date(),
    approvedAt: parseTimestamp(raw.approved_at),
    rejectedAt: parseTimestamp(raw.rejected_at),
    notes: raw.notes,
    operatorId: raw.operator_id,
  };
}

export function inboxListItemFromRaw(raw: RawInboxEntry): InboxListItem {
  return {
    id: raw.id ?? "",
    imei: raw.imei ?? "",
    model: raw.model,
    manufacturer: raw.manufacturer,
    status: (raw.status as InboxStatus) ?? "pending",
    createdAt: parseTimestamp(raw.created_at) ?? new Date(),
  };
}