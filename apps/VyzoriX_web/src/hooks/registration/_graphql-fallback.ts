import {
  queryInboxEntries,
  queryInboxEntry,
  mutateAckInbox,
  mutateDeregisterDevice,
  type InboxEntry,
  type RegisteredDevice,
  type InboxListResult,
  type RegisteredDeviceListResult,
  type AckResult,
  type DeregisterResult,
  type AcknowledgeAction,
  type InboxStatus,
  type Pagination,
} from '@vyzorix/api-client';

interface RawInboxEntryFields {
  id: string;
  imei: string;
  deviceName?: string;
  deviceClass?: string;
  model?: string;
  manufacturer?: string;
  osVersion?: string;
  appVersion?: string;
  fcmToken?: string;
  firebaseInstallId?: string;
  status: InboxStatus;
  acknowledgedAt?: number | null;
  approvingAt?: number | null;
  approvedAt?: number | null;
  rejectedAt?: number | null;
  notes?: string | null;
  operatorId?: string | null;
  createdAt: number;
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
    id: raw.id,
    imei: raw.imei,
    deviceName: raw.deviceName ?? '',
    deviceClass: raw.deviceClass ?? '',
    model: raw.model ?? '',
    manufacturer: raw.manufacturer ?? '',
    osVersion: raw.osVersion ?? '',
    appVersion: raw.appVersion ?? '',
    fcmToken: raw.fcmToken ?? '',
    firebaseInstallId: raw.firebaseInstallId ?? '',
    status: raw.status,
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

function extractData<T>(response: unknown, primary: string, fallback?: string): T | null {
  const r = response as Record<string, unknown> | null;
  if (!r) return null;
  const primaryData = r[primary] as T | undefined;
  if (primaryData !== undefined) return primaryData;
  // The registration GraphQL wrappers return the raw Apollo QueryResult, which
  // nests the operation payload under `.data` (e.g. `{ data: { inbox: {...} } }`).
  const nested = r.data as Record<string, unknown> | undefined;
  if (nested) {
    const nestedPrimary = nested[primary] as T | undefined;
    if (nestedPrimary !== undefined) return nestedPrimary;
  }
  if (fallback) return (r[fallback] as T | undefined) ?? null;
  return null;
}

export async function fetchInboxViaGraphQL(
  organizationId: string,
  params?: { status?: InboxStatus | 'all'; page?: number; limit?: number },
): Promise<InboxListResult> {
  const response = await queryInboxEntries({
    organizationId,
    status: params?.status === 'all' ? undefined : params?.status,
    page: params?.page,
    limit: params?.limit,
  });
  const connection = extractData<Record<string, unknown>>(response, 'inbox', 'data');
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
  const response = await queryInboxEntry({ organizationId, imei });
  const raw = extractData<RawInboxEntryFields>(response, 'inboxEntry', 'data');
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

interface RawAckResultFields {
  id: string;
  imei: string;
  status: InboxStatus;
  acknowledgedAt?: number | null;
  approvingAt?: number | null;
  approvedAt?: number | null;
  rejectedAt?: number | null;
  commandSecret?: string | null;
  fcmPushSent?: boolean;
  notes?: string | null;
}

export async function acknowledgeViaGraphQL(
  _organizationId: string,
  imei: string,
  action: AcknowledgeAction,
  notes?: string,
): Promise<AckResult> {
  const response = await mutateAckInbox({ imei, action, notes });
  const result = extractData<RawAckResultFields>(response, 'ackInbox', 'data');
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
    status: result.status,
    acknowledgedAt: toDate(result.acknowledgedAt),
    approvingAt: toDate(result.approvingAt),
    approvedAt: toDate(result.approvedAt),
    rejectedAt: toDate(result.rejectedAt),
    commandSecret: result.commandSecret ?? null,
    fcmPushSent: result.fcmPushSent ?? false,
    notes: result.notes ?? null,
  };
}

interface RawDeregisterResultFields {
  imei: string;
  status: string;
  deregisteredAt?: number;
  retentionUntil?: number;
}

export async function deregisterViaGraphQL(
  _organizationId: string,
  imei: string,
  hard?: boolean,
): Promise<DeregisterResult> {
  const response = await mutateDeregisterDevice({ imei, hard });
  const result = extractData<RawDeregisterResultFields>(response, 'deregisterDevice', 'data');
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

export { normalizeInboxEntry, normalizePagination };
export type { RawInboxEntryFields, RawPaginationFields };
