import type {
  Version,
  SyncState,
  PushDevices,
  UpdatePush,
  Pagination,
  VersionListResult,
  UpdateHistoryResult,
  ChangelogEntry,
  ReleaseType,
  UpdateStatus,
  InstallType,
  DevicePushStatus,
  SyncStatus,
} from "./updates-entity";

export interface RawVersion {
  id: string;
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releaseType: string;
  releaseNotes?: string;
  releaseDate: number;
  isLatest: boolean;
  createdAt: number;
  updatedAt: number;
}

export interface RawSyncState {
  status: string;
  lastSyncAt?: number | null;
  nextSyncAt?: number | null;
  versionsFound?: number;
  error?: string;
}

export interface RawPushDevices {
  total: number;
  pending: number;
  sent: number;
  acknowledged: number;
  failed: number;
}

export interface RawUpdatePush {
  id: string;
  versionId: string;
  installType: string;
  status: string;
  initiatedBy: string;
  initiatedAt: number;
  scheduledAt?: number | null;
  completedAt?: number | null;
  cancelledAt?: number | null;
  cancelledBy?: string;
  devices: RawPushDevices;
}

export interface RawPagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface RawVersionListResult {
  versions: RawVersion[];
  pagination: RawPagination;
}

export interface RawUpdateHistoryResult {
  pushes: RawUpdatePush[];
  pagination: RawPagination;
}

export interface RawChangelogEntry {
  version: string;
  date: string;
  type: string;
  notes: string;
}

function parseTimestamp(value?: number | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value > 1e12 ? value : value * 1000);
}

export function versionFromRaw(raw: RawVersion): Version {
  return {
    id: raw.id,
    version: raw.version,
    apkFilename: raw.apkFilename,
    apkSize: raw.apkSize,
    sha256: raw.sha256,
    releaseType: (raw.releaseType as ReleaseType) ?? "patch",
    releaseNotes: raw.releaseNotes,
    releaseDate: parseTimestamp(raw.releaseDate) ?? new Date(),
    isLatest: raw.isLatest,
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
    updatedAt: parseTimestamp(raw.updatedAt) ?? new Date(),
  };
}

export function syncStateFromRaw(raw: RawSyncState): SyncState {
  return {
    status: (raw.status as SyncStatus) ?? "idle",
    lastSyncAt: parseTimestamp(raw.lastSyncAt),
    nextSyncAt: parseTimestamp(raw.nextSyncAt),
    versionsFound: raw.versionsFound,
    error: raw.error,
  };
}

export function pushDevicesFromRaw(raw: RawPushDevices): PushDevices {
  return {
    total: raw.total ?? 0,
    pending: raw.pending ?? 0,
    sent: raw.sent ?? 0,
    acknowledged: raw.acknowledged ?? 0,
    failed: raw.failed ?? 0,
  };
}

export function updatePushFromRaw(raw: RawUpdatePush): UpdatePush {
  return {
    id: raw.id,
    versionId: raw.versionId,
    installType: (raw.installType as InstallType) ?? "immediate",
    status: (raw.status as UpdateStatus) ?? "pending",
    initiatedBy: raw.initiatedBy,
    initiatedAt: parseTimestamp(raw.initiatedAt) ?? new Date(),
    scheduledAt: parseTimestamp(raw.scheduledAt),
    completedAt: parseTimestamp(raw.completedAt),
    cancelledAt: parseTimestamp(raw.cancelledAt),
    cancelledBy: raw.cancelledBy,
    devices: pushDevicesFromRaw(raw.devices),
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

export function versionListResultFromRaw(raw: RawVersionListResult): VersionListResult {
  return {
    versions: raw.versions.map(versionFromRaw),
    pagination: paginationFromRaw(raw.pagination),
  };
}

export function updateHistoryResultFromRaw(raw: RawUpdateHistoryResult): UpdateHistoryResult {
  return {
    pushes: raw.pushes.map(updatePushFromRaw),
    pagination: paginationFromRaw(raw.pagination),
  };
}

export function changelogEntryFromRaw(raw: RawChangelogEntry): ChangelogEntry {
  return {
    version: raw.version,
    date: new Date(raw.date),
    type: (raw.type as ReleaseType) ?? "patch",
    notes: raw.notes,
  };
}
