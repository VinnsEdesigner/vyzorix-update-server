import type {
  UpdateVersion,
  UpdatePush,
  SyncStatus,
  ChangelogEntry,
  Pagination,
} from "@/domain/updates";
import type {
  RawUpdateVersion,
  RawUpdatePush,
  RawSyncStatus,
  RawChangelogEntry,
  RawVersionConnection,
  RawUpdateHistoryConnection,
  RawChangelogConnection,
} from "./graphql-updates-types";

function parseTimestamp(value?: string | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value);
}

export function updateVersionFromRaw(raw: RawUpdateVersion): UpdateVersion {
  return {
    id: raw.id,
    version: raw.version,
    apkFilename: raw.apkFilename,
    apkSize: raw.apkSize,
    sha256: raw.sha256,
    releaseDate: parseTimestamp(raw.releaseDate) ?? new Date(),
    releaseNotes: raw.releaseNotes,
    releaseType: raw.releaseType as UpdateVersion["releaseType"],
    isLatest: raw.isLatest,
  };
}

export function pushDeviceFromRaw(raw: RawPushDevice): UpdatePush["devices"][number] {
  return {
    id: raw.id,
    deviceId: raw.deviceId,
    deviceName: raw.deviceName,
    status: raw.status as "pending" | "sent" | "acknowledged" | "failed",
    sentAt: parseTimestamp(raw.sentAt),
    acknowledgedAt: parseTimestamp(raw.acknowledgedAt),
    error: raw.error,
  };
}

export function updatePushFromRaw(raw: RawUpdatePush): UpdatePush {
  return {
    id: raw.id,
    version: raw.version,
    installType: raw.installType as UpdatePush["installType"],
    scheduledAt: undefined,
    status: raw.status as UpdatePush["status"],
    initiatedBy: raw.initiatedBy,
    initiatedAt: parseTimestamp(raw.initiatedAt) ?? new Date(),
    completedAt: parseTimestamp(raw.completedAt),
    cancelledAt: parseTimestamp(raw.cancelledAt),
    deviceCount: raw.deviceCount,
    devices: raw.devices.map(pushDeviceFromRaw),
  };
}

export function syncStatusFromRaw(raw: RawSyncStatus): SyncStatus {
  return {
    status: raw.status as SyncStatus["status"],
    lastSyncAt: parseTimestamp(raw.lastSyncAt),
    nextSyncAt: parseTimestamp(raw.nextSyncAt),
    versionsFound: raw.versionsFound ?? 0,
    error: raw.error,
  };
}

export function changelogEntryFromRaw(raw: RawChangelogEntry): ChangelogEntry {
  return {
    version: raw.version,
    date: parseTimestamp(raw.date) ?? new Date(),
    type: raw.type as ChangelogEntry["type"],
    notes: raw.notes,
  };
}

export function paginationFromRaw(raw: { page: number; limit: number; total: number; totalPages: number; hasMore: boolean }): Pagination {
  return {
    page: raw.page,
    limit: raw.limit,
    total: raw.total,
    totalPages: raw.totalPages,
  };
}

export function versionListResultFromRaw(raw: RawVersionConnection): { versions: UpdateVersion[]; pagination: Pagination } {
  return {
    versions: raw.versions.map(updateVersionFromRaw),
    pagination: paginationFromRaw(raw.pagination),
  };
}

export function updateHistoryResultFromRaw(raw: RawUpdateHistoryConnection): { pushes: UpdatePush[]; pagination: Pagination } {
  return {
    pushes: raw.pushes.map(updatePushFromRaw),
    pagination: paginationFromRaw(raw.pagination),
  };
}

export function changelogResultFromRaw(raw: RawChangelogConnection): { changelog: ChangelogEntry[] } {
  return {
    changelog: raw.changelog.map(changelogEntryFromRaw),
  };
}
