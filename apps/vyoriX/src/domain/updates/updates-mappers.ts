import type {
  Version,
  ChangelogEntry,
  UpdatePush,
  PushDevice,
  SyncStatus,
  DeviceUpdateStatus,
  UpdateStatusResult,
  Pagination,
  VersionListResult,
  UpdateHistoryResult,
  ChangelogResult,
  SyncResult,
  ReleaseType,
  UpdateStatus,
  DevicePushStatus,
  SyncStatusValue,
} from "./updates-entity";

export type RawVersion = {
  id: string;
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releaseDate: string;
  releaseNotes?: string;
  releaseType: string;
  isLatest: boolean;
};

export type RawChangelogEntry = {
  version: string;
  date: string;
  type: string;
  notes: string;
};

export type RawPushDevice = {
  id: string;
  deviceId: string;
  deviceName?: string;
  status: string;
  sentAt?: string | null;
  acknowledgedAt?: string | null;
  error?: string;
};

export type RawUpdatePush = {
  id: string;
  version: string;
  installType: string;
  status: string;
  initiatedBy: string;
  initiatedAt: string;
  completedAt?: string | null;
  cancelledAt?: string | null;
  deviceCount: number;
  devices: RawPushDevice[];
};

export type RawSyncStatus = {
  status: string;
  lastSyncAt?: string | null;
  nextSyncAt?: string | null;
  versionsFound?: number;
  error?: string;
};

export type RawDeviceUpdateStatus = {
  currentVersion?: string;
  needsUpdate: boolean;
};

export type RawUpdateStatusResult = {
  sync: RawSyncStatus;
  latest?: RawVersion;
  device: RawDeviceUpdateStatus;
};

export type RawPagination = {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
};

export type RawVersionListResult = {
  versions: RawVersion[];
  pagination: RawPagination;
};

export type RawUpdateHistoryResult = {
  pushes: RawUpdatePush[];
  pagination: RawPagination;
};

export type RawChangelogResult = {
  changelog: RawChangelogEntry[];
};

export type RawSyncResult = {
  status: string;
  startedAt: string;
  versionsFound?: number;
  message?: string;
};

export function versionFromRaw(raw: RawVersion): Version {
  return {
    id: raw.id,
    version: raw.version,
    apkFilename: raw.apkFilename,
    apkSize: raw.apkSize,
    sha256: raw.sha256,
    releaseDate: new Date(raw.releaseDate),
    releaseNotes: raw.releaseNotes,
    releaseType: raw.releaseType as ReleaseType,
    isLatest: raw.isLatest,
  };
}

export function changelogEntryFromRaw(raw: RawChangelogEntry): ChangelogEntry {
  return {
    version: raw.version,
    date: new Date(raw.date),
    type: raw.type as ReleaseType,
    notes: raw.notes,
  };
}

export function pushDeviceFromRaw(raw: RawPushDevice): PushDevice {
  return {
    id: raw.id,
    deviceId: raw.deviceId,
    deviceName: raw.deviceName,
    status: raw.status as DevicePushStatus,
    sentAt: raw.sentAt ? new Date(raw.sentAt) : undefined,
    acknowledgedAt: raw.acknowledgedAt ? new Date(raw.acknowledgedAt) : undefined,
    error: raw.error,
  };
}

export function updatePushFromRaw(raw: RawUpdatePush): UpdatePush {
  return {
    id: raw.id,
    version: raw.version,
    installType: raw.installType as "immediate" | "scheduled",
    status: raw.status as UpdateStatus,
    initiatedBy: raw.initiatedBy,
    initiatedAt: new Date(raw.initiatedAt),
    completedAt: raw.completedAt ? new Date(raw.completedAt) : undefined,
    cancelledAt: raw.cancelledAt ? new Date(raw.cancelledAt) : undefined,
    deviceCount: raw.deviceCount,
    devices: raw.devices.map(pushDeviceFromRaw),
  };
}

export function syncStatusFromRaw(raw: RawSyncStatus): SyncStatus {
  return {
    status: raw.status as SyncStatusValue,
    lastSyncAt: raw.lastSyncAt ? new Date(raw.lastSyncAt) : undefined,
    nextSyncAt: raw.nextSyncAt ? new Date(raw.nextSyncAt) : undefined,
    versionsFound: raw.versionsFound,
    error: raw.error,
  };
}

export function deviceUpdateStatusFromRaw(raw: RawDeviceUpdateStatus): DeviceUpdateStatus {
  return {
    currentVersion: raw.currentVersion,
    needsUpdate: raw.needsUpdate,
  };
}

export function updateStatusResultFromRaw(raw: RawUpdateStatusResult): UpdateStatusResult {
  return {
    sync: syncStatusFromRaw(raw.sync),
    latest: raw.latest ? versionFromRaw(raw.latest) : undefined,
    device: deviceUpdateStatusFromRaw(raw.device),
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

export function changelogResultFromRaw(raw: RawChangelogResult): ChangelogResult {
  return {
    changelog: raw.changelog.map(changelogEntryFromRaw),
  };
}

export function syncResultFromRaw(raw: RawSyncResult): SyncResult {
  return {
    status: raw.status as SyncStatusValue,
    startedAt: new Date(raw.startedAt),
    versionsFound: raw.versionsFound,
    message: raw.message,
  };
}
