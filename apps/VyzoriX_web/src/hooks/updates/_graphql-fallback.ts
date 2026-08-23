import {
  queryUpdates,
  queryUpdatesStatus,
  queryUpdatesChangelog,
  queryUpdatesHistory,
  mutatePushUpdate,
  mutateCancelUpdate,
  mutateSyncUpdates,
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
  return queryUpdates({
    organizationId,
    status: params?.status,
    limit: params?.limit,
    offset: params?.offset,
  });
}

export async function fetchUpdateStatusViaGraphQL(
  organizationId: string,
): Promise<UpdateStatusResponse> {
  const result = await queryUpdatesStatus({ organizationId });
  return {
    sync: result.sync,
    latest: result.latest,
  };
}

export async function fetchChangelogViaGraphQL(
  organizationId: string,
  version?: string,
): Promise<ChangelogEntry[]> {
  return queryUpdatesChangelog({ organizationId, version });
}

export async function fetchUpdateHistoryViaGraphQL(
  organizationId: string,
  params?: HistoryFallbackParams,
): Promise<UpdateHistoryResult> {
  return queryUpdatesHistory({
    organizationId,
    status: params?.status,
    page: params?.page,
    limit: params?.limit,
  });
}

export async function pushUpdateViaGraphQL(
  organizationId: string,
  params: PushFallbackParams,
): Promise<UpdatePush> {
  const result = await mutatePushUpdate({
    organizationId,
    version: params.version,
    deviceIds: params.deviceIds,
    installType: params.installType,
    scheduledAt: params.scheduledAt?.getTime(),
  });
  return {
    id: result.pushId,
    version: result.version,
    installType: result.installType,
    status: result.status,
    initiatedBy: result.initiatedBy,
    initiatedAt: result.initiatedAt,
    scheduledAt: result.scheduledAt,
    devices: { ...EMPTY_DEVICES, total: result.deviceCount },
  };
}

export async function cancelUpdateViaGraphQL(
  organizationId: string,
  pushId: string,
): Promise<UpdatePush> {
  const result = await mutateCancelUpdate({ organizationId, id: pushId });
  return {
    id: result.id,
    version: '',
    installType: 'immediate',
    status: result.status,
    initiatedBy: '',
    initiatedAt: result.cancelledAt,
    cancelledAt: result.cancelledAt,
    cancelledBy: result.cancelledBy,
    devices: { ...EMPTY_DEVICES },
  };
}

export async function syncUpdatesViaGraphQL(
  organizationId: string,
): Promise<{ status: string; startedAt: Date; versionsFound?: number }> {
  const result = await mutateSyncUpdates({ organizationId });
  return {
    status: result.status,
    startedAt: result.startedAt,
    versionsFound: result.versionsFound,
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
  version?: string;
  apkFilename?: string;
  apkSize?: number;
  sha256?: string;
  releaseType?: string;
  releaseNotes?: string;
  releasedAt?: number;
  isLatest?: boolean;
}): UpdateVersion {
  const releasedAt = toDate(v.releasedAt) ?? new Date();
  return {
    id: v.version ?? '',
    version: v.version ?? '',
    apkFilename: v.apkFilename ?? '',
    apkSize: v.apkSize ?? 0,
    sha256: v.sha256 ?? '',
    releaseType: (v.releaseType ?? 'patch') as ReleaseType,
    releaseNotes: v.releaseNotes,
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
