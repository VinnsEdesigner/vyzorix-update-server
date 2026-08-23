// Updates domain — generated types + constants.
export type {
  UpdaterVersionManifestResult,
  UpdateChangelogEntryResult,
  UpdaterCheckResult,
  DownloadProgressRequest,
  DownloadProgressResult,
  UpdateVersionResponse,
  UpdateVersionListResult,
  UpdateChangelogResult,
  UpdateStatusResult,
  UpdateSyncResponse,
  UpdateExportResult,
  UpdatePushRequest,
  UpdatePushResult,
  UpdatePushHistoryEntry,
  UpdatePushHistoryListResult,
  UpdatePushDetailResult,
  UpdateCancelPushResult,
  UpdateSyncStatusInfo,
  UpdateSyncStatusResult,
  DeviceUpdateStatusRequest,
  DeviceUpdateStatusResponse,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

// ---- Constants (hand-rolled) ----

export type SyncStatus = 'idle' | 'syncing' | 'synced' | 'error';
// Server-authoritative release types (internal/domain/updates).
export type ReleaseType = 'major' | 'minor' | 'patch';
export type InstallType = 'immediate' | 'scheduled';

// ---- Hook-facing domain types (camelCase, Date-based; not in OpenAPI) ----
// Used by the GraphQL fallback path and the updates UI. The REST wire DTOs
// above are snake_case; these are the normalized shapes hooks work with.

import type { Pagination } from '../_shared';
export type { Pagination };

export type UpdateStatus = 'pending' | 'in_progress' | 'completed' | 'failed' | 'cancelled';
export type DevicePushStatus = 'pending' | 'sent' | 'in_progress' | 'acknowledged' | 'completed' | 'failed';

export interface UpdateVersion {
  id: string;
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releaseType: ReleaseType;
  releaseNotes?: string;
  releaseDate: Date;
  isLatest: boolean;
  createdAt: Date;
  updatedAt: Date;
}

export interface SyncState {
  status: SyncStatus;
  lastSyncAt?: Date;
  nextSyncAt?: Date;
  versionsFound?: number;
  error?: string;
}

export interface PushDevices {
  total: number;
  pending: number;
  sent: number;
  acknowledged: number;
  failed: number;
}

export interface UpdatePush {
  id: string;
  version: string;
  installType: InstallType;
  status: UpdateStatus;
  initiatedBy: string;
  initiatedAt: Date;
  scheduledAt?: Date;
  completedAt?: Date;
  cancelledAt?: Date;
  cancelledBy?: string;
  devices: PushDevices;
}

export interface VersionListResult {
  versions: UpdateVersion[];
  pagination: Pagination;
}

export interface UpdateHistoryResult {
  pushes: UpdatePush[];
  pagination: Pagination;
}

export interface ChangelogEntry {
  version: string;
  date: Date;
  type: ReleaseType;
  notes: string;
}

/** Subset of the status query the GraphQL fallback can serve. */
export interface UpdateStatusResponse {
  sync: SyncState;
  latest?: UpdateVersion;
}

export function formatApkSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function getReleaseTypeLabel(type: ReleaseType): string {
  const labels: Record<ReleaseType, string> = { major: 'Major', minor: 'Minor', patch: 'Patch' };
  return labels[type];
}

export function getUpdateStatusLabel(status: UpdateStatus): string {
  const labels: Record<UpdateStatus, string> = {
    pending: 'Pending',
    in_progress: 'In Progress',
    completed: 'Completed',
    failed: 'Failed',
    cancelled: 'Cancelled',
  };
  return labels[status];
}

