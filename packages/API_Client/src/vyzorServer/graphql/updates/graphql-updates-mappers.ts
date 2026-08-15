import type {
  UpdateVersion,
  UpdatePush,
  SyncStatus,
  ChangelogEntry,
  Pagination,
} from "../../../domain/updates";
import type {
  RawUpdateVersion,
  RawUpdatePush,
  RawSyncStatus,
  RawChangelogEntry,
  RawPushDevice,
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
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
    updatedAt: parseTimestamp(raw.updatedAt) ?? new Date(),
  };
}

export interface PushDeviceResult {
  id: string;
  deviceId: string;
  deviceName: string;
  status: "pending" | "sent" | "acknowledged" | "failed";
  sentAt?: Date;
  acknowledgedAt?: Date;
  error?: string;
}

export function pushDeviceFromRaw(raw: RawPushDevice): PushDeviceResult {
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
  const deviceArray = raw.devices.map(pushDeviceFromRaw);
  return {
    id: raw.id,
    versionId: raw.version,
    installType: raw.installType as UpdatePush["installType"],
    scheduledAt: undefined,
    status: raw.status as UpdatePush["status"],
    initiatedBy: raw.initiatedBy,
    initiatedAt: parseTimestamp(raw.initiatedAt) ?? new Date(),
    completedAt: parseTimestamp(raw.completedAt),
    cancelledAt: parseTimestamp(raw.cancelledAt),
    devices: {
      total: deviceArray.length,
      pending: deviceArray.filter(d => d.status === "pending").length,
      sent: deviceArray.filter(d => d.status === "sent").length,
      acknowledged: deviceArray.filter(d => d.status === "acknowledged").length,
      failed: deviceArray.filter(d => d.status === "failed").length,
    },
  };
}

export function syncStatusFromRaw(raw: RawSyncStatus): SyncStatus {
  return raw.status as SyncStatus;
}

export function changelogEntryFromRaw(raw: RawChangelogEntry): ChangelogEntry {
  return {
    version: raw.version,
    date: parseTimestamp(raw.date) ?? new Date(),
    type: raw.type as ChangelogEntry["type"],
    notes: raw.notes,
  };
}

export function paginationFromRaw(raw: { page?: number; limit: number; total: number; totalPages?: number; hasMore: boolean; offset?: number }): Pagination {
  return {
    page: raw.page ?? Math.floor((raw.offset ?? 0) / raw.limit) + 1,
    limit: raw.limit,
    total: raw.total,
    totalPages: raw.totalPages ?? Math.ceil(raw.total / raw.limit),
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
