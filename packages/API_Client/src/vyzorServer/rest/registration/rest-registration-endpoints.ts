import { restClient } from "../_shared/rest-client";
import type { InboxEntry, InboxStatus, AcknowledgeAction, AckResult, DeregisterResult, InboxListResult } from "@/domain/registration";

export const REGISTRATION_PATHS = {
  inbox: "/v1/device/inbox",
  inboxEntry: (imei: string) => `/v1/device/inbox/${imei}`,
  inboxAck: (imei: string) => `/v1/device/inbox/${imei}/ack`,
  devices: "/v1/devices",
  device: (imei: string) => `/v1/devices/${imei}`,
  deregister: (imei: string) => `/v1/devices/${imei}`,
  register: "/v1/device/register",
  confirm: "/v1/device/confirm",
} as const;

interface RawInboxEntry {
  id: string;
  imei: string;
  model?: string;
  manufacturer?: string;
  os_version?: string;
  app_version?: string;
  fcm_token: string;
  firebase_install_id: string;
  status: string;
  created_at: number;
  approved_at?: number;
  rejected_at?: number;
  command_secret?: string;
  notes?: string;
  operator_id?: string;
}

interface RawPagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

function parseTimestamp(value?: number | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value > 1e12 ? value : value * 1000);
}

function inboxEntryFromRaw(raw: RawInboxEntry): InboxEntry {
  return {
    id: raw.id,
    imei: raw.imei,
    model: raw.model,
    manufacturer: raw.manufacturer,
    osVersion: raw.os_version,
    appVersion: raw.app_version,
    fcmToken: raw.fcm_token,
    firebaseInstallId: raw.firebase_install_id,
    status: raw.status as InboxStatus,
    createdAt: parseTimestamp(raw.created_at) ?? new Date(),
    approvedAt: parseTimestamp(raw.approved_at),
    rejectedAt: parseTimestamp(raw.rejected_at),
    commandSecret: raw.command_secret,
    notes: raw.notes,
    operatorId: raw.operator_id,
  };
}

export async function fetchInbox(params?: {
  status?: InboxStatus;
  page?: number;
  limit?: number;
}): Promise<InboxListResult> {
  const data = await restClient.get<{
    requests: RawInboxEntry[];
    pagination: RawPagination;
  }>(REGISTRATION_PATHS.inbox, {
    status: params?.status,
    page: params?.page,
    limit: params?.limit,
  });

  return {
    requests: data.requests.map(inboxEntryFromRaw),
    pagination: {
      page: data.pagination.page,
      limit: data.pagination.limit,
      total: data.pagination.total,
      totalPages: data.pagination.total_pages,
    },
  };
}

export async function fetchInboxEntry(imei: string): Promise<InboxEntry | null> {
  const data = await restClient.get<RawInboxEntry | null>(REGISTRATION_PATHS.inboxEntry(imei));
  if (!data || !data.imei) return null;
  return inboxEntryFromRaw(data);
}

export async function acknowledgeInbox(
  imei: string,
  action: AcknowledgeAction,
  notes?: string
): Promise<AckResult> {
  const data = await restClient.post<{
    id: string;
    imei: string;
    status: string;
    approved_at?: number;
    rejected_at?: number;
    command_secret?: string;
    fcm_push_sent?: boolean;
    notes?: string;
  }>(REGISTRATION_PATHS.inboxAck(imei), { action, notes });

  return {
    id: data.id,
    imei: data.imei,
    status: data.status as InboxStatus,
    approvedAt: parseTimestamp(data.approved_at),
    rejectedAt: parseTimestamp(data.rejected_at),
    commandSecret: data.command_secret,
    fcmPushSent: data.fcm_push_sent,
    notes: data.notes,
  };
}

export async function deregisterDevice(imei: string): Promise<DeregisterResult> {
  const data = await restClient.delete<{
    imei: string;
    status: string;
    deregistered_at: number;
    retention_until: number;
  }>(REGISTRATION_PATHS.deregister(imei));

  return {
    imei: data.imei,
    status: data.status,
    deregisteredAt: parseTimestamp(data.deregistered_at) ?? new Date(),
    retentionUntil: parseTimestamp(data.retention_until) ?? new Date(),
  };
}
