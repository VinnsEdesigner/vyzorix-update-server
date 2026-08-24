import {
  gqlQuery,
  gqlMutate,
  type UpdateVersion,
  type UpdatePush,
  type VersionListResult,
  type UpdateHistoryResult,
  type ChangelogEntry,
  type SyncState,
  type PushDevices,
  type InstallType,
  type UpdateStatus,
  type UpdateStatusResponse,
  type ReleaseType,
  type Pagination,
  type UpdateVersionListResult,
  type UpdatePushHistoryListResult,
  type UpdatePushResult as WireUpdatePushResult,
  type UpdateCancelPushResult,
  type UpdateChangelogResult,
  type UpdateStatusResult,
} from '@vyzorix/api-client';
import {
  GetUpdatesDocument,
  GetUpdatesStatusDocument,
  GetUpdatesChangelogDocument,
  GetUpdatesHistoryDocument,
  PushUpdateDocument,
  CancelUpdateDocument,
  SyncUpdatesDocument,
  type GetUpdatesQuery,
  type GetUpdatesStatusQuery,
  type GetUpdatesChangelogQuery,
  type GetUpdatesHistoryQuery,
  type PushUpdateMutation,
  type CancelUpdateMutation,
  type SyncUpdatesMutation,
} from '@vyzorix/api-client/generated-graphql';

interface VersionFallbackParams {
  status?: string;
  limit?: number;
  offset?: number;
}

interface HistoryFallbackParams {
  status?: string;
  page?: number;
  limit?: number;
}

interface PushFallbackParams {
  version: string;
  deviceIds: string[];
  installType: InstallType;
  scheduledAt?: Date;
}

const EMPTY_DEVICES: PushDevices = {
  total: 0,
  pending: 0,
  sent: 0,
  acknowledged: 0,
  failed: 0,
};

export async function fetchVersionsViaGraphQL(
  organizationId: string,
  params?: VersionFallbackParams,
): Promise<VersionListResult> {
  const data = await gqlQuery<GetUpdatesQuery>(GetUpdatesDocument, {
    organizationId,
    status: params?.status,
    limit: params?.limit,
    offset: params?.offset,
  });
  const connection = data.updatesVersions;
  const wirePagination = connection?.pagination;
  return {
    versions: (connection?.versions ?? []).map((v) => normalizeWireVersion(v)),
    pagination: wirePagination
      ? {
          page: Math.floor((wirePagination.offset ?? 0) / (wirePagination.limit || 20)) + 1,
          limit: wirePagination.limit,
          total: wirePagination.total,
          totalPages: Math.ceil((wirePagination.total ?? 0) / (wirePagination.limit || 20)),
        }
      : { page: 1, limit: 20, total: 0, totalPages: 0 },
  };
}

export async function fetchUpdateStatusViaGraphQL(
  organizationId: string,
): Promise<UpdateStatusResponse> {
  const data = await gqlQuery<GetUpdatesStatusQuery>(GetUpdatesStatusDocument, { organizationId });
  const status = data.updatesStatus;
  if (!status) throw new Error('GraphQL updatesStatus returned no data');
  return normalizeWireUpdateStatus(status as unknown as UpdateStatusResult);
}

export async function fetchChangelogViaGraphQL(
  organizationId: string,
  version?: string,
): Promise<ChangelogEntry[]> {
  const data = await gqlQuery<GetUpdatesChangelogQuery>(GetUpdatesChangelogDocument, { organizationId, version });
  return (data.updatesChangelog ?? []).map((e) => ({
    version: e.version,
    date: toDate(e.date) ?? new Date(),
    type: (e.type ?? 'patch') as ReleaseType,
    notes: e.notes,
  }));
}

export async function fetchUpdateHistoryViaGraphQL(
  organizationId: string,
  params?: HistoryFallbackParams,
): Promise<UpdateHistoryResult> {
  const data = await gqlQuery<GetUpdatesHistoryQuery>(GetUpdatesHistoryDocument, {
    organizationId,
    status: params?.status,
    page: params?.page,
    limit: params?.limit,
  });
  const connection = data.updatesHistory;
  const wirePagination = connection?.pagination;
  return {
    pushes: (connection?.pushes ?? []).filter((p): p is NonNullable<typeof p> => p != null).map((p) => normalizeWireHistoryEntry({
      id: p.id,
      version: p.version,
      installType: p.installType,
      status: p.status,
      initiatedBy: p.initiatedBy,
      initiatedAt: p.initiatedAt,
      completedAt: p.completedAt ?? undefined,
      deviceCount: p.deviceCount,
      devices: { pending: p.pending, acknowledged: p.acknowledged, failed: p.failed },
    })),
    pagination: wirePagination
      ? {
          page: Math.floor((wirePagination.offset ?? 0) / (wirePagination.limit || 20)) + 1,
          limit: wirePagination.limit,
          total: wirePagination.total,
          totalPages: Math.ceil((wirePagination.total ?? 0) / (wirePagination.limit || 20)),
        }
      : { page: 1, limit: 20, total: 0, totalPages: 0 },
  };
}

export async function pushUpdateViaGraphQL(
  organizationId: string,
  params: PushFallbackParams,
): Promise<UpdatePush> {
  const data = await gqlMutate<PushUpdateMutation>(PushUpdateDocument, {
    organizationId,
    version: params.version,
    deviceIds: params.deviceIds,
    installType: params.installType,
    scheduledAt: params.scheduledAt?.getTime(),
  });
  const result = data.pushUpdate;
  if (!result) throw new Error('GraphQL pushUpdate returned no data');
  return {
    id: result.pushId,
    version: result.version,
    installType: result.installType as InstallType,
    status: result.status as UpdateStatus,
    initiatedBy: result.initiatedBy,
    initiatedAt: toDate(result.initiatedAt) ?? new Date(),
    scheduledAt: toDate(result.scheduledAt),
    devices: { ...EMPTY_DEVICES, total: result.deviceCount },
  };
}

export async function cancelUpdateViaGraphQL(
  organizationId: string,
  pushId: string,
): Promise<UpdatePush> {
  const data = await gqlMutate<CancelUpdateMutation>(CancelUpdateDocument, { organizationId, id: pushId });
  const result = data.cancelUpdate;
  if (!result) throw new Error('GraphQL cancelUpdate returned no data');
  return {
    id: result.id,
    version: '',
    installType: 'immediate',
    status: result.status as UpdateStatus,
    initiatedBy: '',
    initiatedAt: toDate(result.cancelledAt) ?? new Date(),
    cancelledAt: toDate(result.cancelledAt),
    cancelledBy: result.cancelledBy,
    devices: { ...EMPTY_DEVICES },
  };
}

export async function syncUpdatesViaGraphQL(
  _organizationId: string,
): Promise<{ status: string; startedAt: Date; versionsFound?: number }> {
  const data = await gqlMutate<SyncUpdatesMutation>(SyncUpdatesDocument, {});
  const result = data.syncUpdates;
  if (!result) throw new Error('GraphQL syncUpdates returned no data');
  return {
    status: result.status,
    startedAt: toDate(result.startedAt) ?? new Date(),
    versionsFound: result.versionsFound ?? undefined,
  };
}

export type {
  UpdateVersion,
  UpdatePush,
  VersionListResult,
  UpdateHistoryResult,
  ChangelogEntry,
  SyncState,
  PushDevices,
  InstallType,
  UpdateStatus,
  UpdateStatusResponse,
};

// ---- Wire (REST DTO) -> domain normalizers ----

function toDate(value?: number | string | null): Date | undefined {
  if (value === undefined || value === null) return undefined;
  return new Date(value);
}

function normalizePagination(raw?: { page?: number; limit?: number; total?: number; total_pages?: number }): Pagination {
  return {
    page: raw?.page ?? 1,
    limit: raw?.limit ?? 20,
    total: raw?.total ?? 0,
    totalPages: raw?.total_pages ?? 0,
  };
}

export function normalizeWireVersion(v: {
  version?: string | null;
  apkFilename?: string | null;
  apkSize?: number | null;
  sha256?: string | null;
  releaseType?: string | null;
  releaseNotes?: string | null;
  releasedAt?: number | string | null;
  isLatest?: boolean | null;
}): UpdateVersion {
  const releasedAt = toDate(v.releasedAt) ?? new Date();
  return {
    id: v.version ?? '',
    version: v.version ?? '',
    apkFilename: v.apkFilename ?? '',
    apkSize: v.apkSize ?? 0,
    sha256: v.sha256 ?? '',
    releaseType: (v.releaseType ?? 'patch') as ReleaseType,
    releaseNotes: v.releaseNotes ?? undefined,
    releaseDate: releasedAt,
    isLatest: v.isLatest ?? false,
    createdAt: releasedAt,
    updatedAt: releasedAt,
  };
}

export function normalizeWireVersionList(raw: UpdateVersionListResult): VersionListResult {
  return {
    versions: (raw.versions ?? []).map(normalizeWireVersion),
    pagination: normalizePagination(raw.pagination),
  };
}

export function normalizeWireHistoryEntry(e: {
  id?: string;
  version?: string;
  installType?: string;
  status?: string;
  initiatedBy?: string;
  initiatedAt?: number;
  scheduledAt?: number;
  completedAt?: number;
  cancelledAt?: number;
  deviceCount?: number;
  devices?: { pending?: number; sent?: number; acknowledged?: number; failed?: number };
}): UpdatePush {
  return {
    id: e.id ?? '',
    version: e.version ?? '',
    installType: (e.installType ?? 'immediate') as InstallType,
    status: (e.status ?? 'pending') as UpdateStatus,
    initiatedBy: e.initiatedBy ?? '',
    initiatedAt: toDate(e.initiatedAt) ?? new Date(),
    scheduledAt: toDate(e.scheduledAt),
    completedAt: toDate(e.completedAt),
    cancelledAt: toDate(e.cancelledAt),
    devices: {
      total: e.deviceCount ?? 0,
      pending: e.devices?.pending ?? 0,
      sent: e.devices?.sent ?? 0,
      acknowledged: e.devices?.acknowledged ?? 0,
      failed: e.devices?.failed ?? 0,
    },
  };
}

export function normalizeWireHistoryList(raw: UpdatePushHistoryListResult): UpdateHistoryResult {
  return {
    pushes: (raw.pushes ?? []).map(normalizeWireHistoryEntry),
    pagination: normalizePagination(raw.pagination),
  };
}

export function normalizeWirePushResult(r: WireUpdatePushResult): UpdatePush {
  return {
    id: r.pushId ?? '',
    version: r.version ?? '',
    installType: (r.installType ?? 'immediate') as InstallType,
    status: (r.status ?? 'pending') as UpdateStatus,
    initiatedBy: r.initiatedBy ?? '',
    initiatedAt: toDate(r.initiatedAt) ?? new Date(),
    scheduledAt: toDate(r.scheduledAt),
    devices: {
      total: r.devices?.total ?? r.deviceIds?.length ?? 0,
      pending: r.devices?.pending ?? 0,
      sent: r.devices?.sent ?? 0,
      acknowledged: r.devices?.acknowledged ?? 0,
      failed: r.devices?.failed ?? 0,
    },
  };
}

export function normalizeWireCancelResult(r: UpdateCancelPushResult): UpdatePush {
  const cancelledAt = toDate(r.cancelledAt);
  return {
    id: r.id ?? '',
    version: '',
    installType: 'immediate',
    status: (r.status ?? 'cancelled') as UpdateStatus,
    initiatedBy: '',
    initiatedAt: cancelledAt ?? new Date(),
    cancelledAt,
    cancelledBy: r.cancelledBy,
    devices: { ...EMPTY_DEVICES },
  };
}

export function normalizeWireChangelog(raw: UpdateChangelogResult): ChangelogEntry[] {
  return (raw.changelog ?? []).map((e) => ({
    version: e.version ?? '',
    date: toDate(e.date) ?? new Date(),
    type: (e.type ?? 'patch') as ReleaseType,
    notes: e.notes ?? '',
  }));
}

export function normalizeWireUpdateStatus(raw: UpdateStatusResult): UpdateStatusResponse {
  return {
    sync: {
      status: (raw.sync?.status ?? 'idle') as SyncState['status'],
      lastSyncAt: toDate(raw.sync?.lastSyncAt),
      nextSyncAt: toDate(raw.sync?.nextSyncAt),
      versionsFound: raw.sync?.versionsFound,
      error: raw.sync?.error,
    },
    latest: raw.latest ? normalizeWireVersion(raw.latest) : undefined,
  };
}
