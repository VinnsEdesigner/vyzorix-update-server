import {
  gqlQuery,
  gqlMutate,
  type InboxEntry,
  type RegisteredDevice,
  type InboxEntriesResult,
  type RegisteredDeviceListResult,
  type AckResult,
  type DeregisterResult,
  type AcknowledgeAction,
  type InboxStatus,
  type Pagination,
} from '@vyzorix/api-client';
import {
  GetInboxEntriesDocument,
  GetInboxEntryDocument,
  AckInboxDocument,
  DeregisterDeviceDocument,
  type GetInboxEntriesQuery,
  type GetInboxEntryQuery,
  type AckInboxMutation,
  type DeregisterDeviceMutation,
  type AckAction,
} from '@vyzorix/api-client/generated-graphql';

interface RawInboxEntryFields {
  id?: string;
  imei?: string;
  deviceName?: string;
  deviceClass?: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  fcmToken?: string;
  firebaseInstallId?: string;
  status?: InboxStatus | string;
  acknowledgedAt?: number | null;
  approvingAt?: number | null;
  approvedAt?: number | null;
  rejectedAt?: number | null;
  notes?: string | null;
  operatorId?: string | null;
  createdAt?: number;
}

interface RawPaginationFields {
  page?: number;
  limit?: number;
  total?: number;
  totalPages?: number;
  total_pages?: number;
}

function toDate(value: number | null | undefined): Date | null {
  if (!value) return null;
  return new Date(value > 1e12 ? value : value * 1000);
}

function normalizeInboxEntry(raw: RawInboxEntryFields): InboxEntry {
  return {
    id: raw.id ?? '',
    imei: raw.imei ?? '',
    deviceName: raw.deviceName ?? '',
    deviceClass: raw.deviceClass ?? '',
    model: raw.model ?? '',
    manufacturer: raw.manufacturer ?? '',
    osVersion: raw.osVersion ?? '',
    appVersion: raw.appVersion ?? '',
    fcmToken: raw.fcmToken ?? '',
    firebaseInstallId: raw.firebaseInstallId ?? '',
    status: (raw.status ?? 'pending') as InboxStatus,
    acknowledgedAt: toDate(raw.acknowledgedAt),
    approvingAt: toDate(raw.approvingAt),
    approvedAt: toDate(raw.approvedAt),
    rejectedAt: toDate(raw.rejectedAt),
    notes: raw.notes ?? null,
    operatorId: raw.operatorId ?? null,
    createdAt: toDate(raw.createdAt) ?? new Date(),
  };
}

function normalizePagination(raw: RawPaginationFields): Pagination {
  return {
    page: raw.page ?? 1,
    limit: raw.limit ?? 20,
    total: raw.total ?? 0,
    totalPages: raw.totalPages ?? raw.total_pages ?? 0,
  };
}


export async function fetchInboxViaGraphQL(
  organizationId: string,
  params?: { status?: InboxStatus | 'all'; page?: number; limit?: number },
): Promise<InboxEntriesResult> {
  const data = await gqlQuery<GetInboxEntriesQuery>(GetInboxEntriesDocument, {
    organizationId,
    status: params?.status === 'all' ? undefined : params?.status,
    page: params?.page,
    limit: params?.limit,
  });
  const connection = data.inbox;
  const requests = (connection?.requests as RawInboxEntryFields[] | undefined) ?? [];
  const paginationRaw = (connection?.pagination as RawPaginationFields | undefined) ?? {};
  return {
    requests: requests.map(normalizeInboxEntry),
    pagination: normalizePagination(paginationRaw),
  };
}

export async function fetchInboxEntryViaGraphQL(
  organizationId: string,
  imei: string,
): Promise<InboxEntry | null> {
  const data = await gqlQuery<GetInboxEntryQuery>(GetInboxEntryDocument, { organizationId, imei });
  const raw = data.inboxEntry as RawInboxEntryFields | null;
  if (!raw?.imei) return null;
  return normalizeInboxEntry(raw);
}

export async function fetchRegisteredDevicesViaGraphQL(
  _organizationId: string,
  _params?: { status?: string; page?: number; limit?: number },
): Promise<RegisteredDeviceListResult> {
  throw new Error('GraphQL fallback for registered device list is not implemented');
}

export async function fetchRegisteredDeviceViaGraphQL(
  _organizationId: string,
  _imei: string,
): Promise<RegisteredDevice | null> {
  throw new Error('GraphQL fallback for a single registered device is not implemented');
}

export async function acknowledgeViaGraphQL(
  _organizationId: string,
  imei: string,
  action: AcknowledgeAction,
  notes?: string,
): Promise<AckResult> {
  const data = await gqlMutate<AckInboxMutation>(AckInboxDocument, {
    imei,
    action: action.toUpperCase() as AckAction,
    notes,
  });
  const result = data.ackInbox;
  if (!result) {
    return {
      id: '',
      imei,
      status: action === 'reject' ? 'rejected' : 'approved',
      acknowledgedAt: null,
      approvingAt: null,
      approvedAt: null,
      rejectedAt: null,
      commandSecret: null,
      fcmPushSent: false,
      notes: notes ?? null,
    };
  }
  return {
    id: result.id,
    imei: result.imei,
    status: result.status as InboxStatus,
    acknowledgedAt: null,
    approvingAt: null,
    approvedAt: toDate(result.approvedAt),
    rejectedAt: toDate(result.rejectedAt),
    commandSecret: result.commandSecret ?? null,
    fcmPushSent: result.fcmPushSent ?? false,
    notes: result.notes ?? null,
  };
}

export async function deregisterViaGraphQL(
  _organizationId: string,
  imei: string,
  hard?: boolean,
): Promise<DeregisterResult> {
  const data = await gqlMutate<DeregisterDeviceMutation>(DeregisterDeviceDocument, { imei, hard });
  const result = data.deregisterDevice;
  if (!result) {
    return {
      imei,
      status: 'deregistered',
      deregisteredAt: new Date(),
      retentionUntil: new Date(Date.now() + 30 * 24 * 60 * 60 * 1000),
    };
  }
  return {
    imei: result.imei,
    status: result.status,
    deregisteredAt: toDate(result.deregisteredAt) ?? new Date(),
    retentionUntil: toDate(result.retentionUntil) ?? new Date(),
  };
}

export function normalizeRegisteredDevice(raw: {
  id?: string;
  imei?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  app_version?: string;
  status?: string;
  registered_at?: number;
  last_seen?: number;
  online?: boolean;
}): RegisteredDevice {
  return {
    id: raw.id ?? '',
    imei: raw.imei ?? '',
    deviceName: raw.device_name ?? '',
    model: raw.model ?? '',
    manufacturer: raw.manufacturer ?? '',
    osVersion: '',
    appVersion: raw.app_version ?? '',
    status: (raw.status ?? 'offline') as RegisteredDevice['status'],
    registeredAt: toDate(raw.registered_at),
    lastSeen: toDate(raw.last_seen),
    online: raw.online ?? false,
  };
}

export function normalizeAckResult(raw: {
  id?: string;
  imei?: string;
  status?: string;
  acknowledgedAt?: number | null;
  approvingAt?: number | null;
  approvedAt?: number | null;
  rejectedAt?: number | null;
  commandSecret?: string | null;
  fcmPushSent?: boolean;
  notes?: string | null;
}): AckResult {
  return {
    id: raw.id ?? '',
    imei: raw.imei ?? '',
    status: (raw.status ?? 'pending') as InboxStatus,
    acknowledgedAt: toDate(raw.acknowledgedAt),
    approvingAt: toDate(raw.approvingAt),
    approvedAt: toDate(raw.approvedAt),
    rejectedAt: toDate(raw.rejectedAt),
    commandSecret: raw.commandSecret ?? null,
    fcmPushSent: raw.fcmPushSent ?? false,
    notes: raw.notes ?? null,
  };
}

export { normalizeInboxEntry, normalizePagination };
export type { RawInboxEntryFields, RawPaginationFields };
