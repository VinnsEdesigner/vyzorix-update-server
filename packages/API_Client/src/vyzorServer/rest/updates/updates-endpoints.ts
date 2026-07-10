// Updates REST Endpoints
// Based on SERVER_BACKEND_UPDATES_API.md

import { restClient } from '../_shared/rest-client';
import type { OffsetPagination } from '@/domain/_shared';

// ============================================================================
// API Paths
// ============================================================================

export const UPDATES_PATHS = {
  status: '/v1/updates/status',
  versions: '/v1/updates/versions',
  changelog: '/v1/updates/changelog',
  push: '/v1/updates/push',
  history: '/v1/updates/history',
  historyDetail: (pushId: string) => `/v1/updates/history/${pushId}`,
  historyCancel: (pushId: string) => `/v1/updates/history/${pushId}/cancel`,
  export: '/v1/updates/export',
  sync: '/v1/updates/sync',
} as const;

// ============================================================================
// Types (snake_case from server)
// ============================================================================

export interface RawSyncStatus {
  status: string;
  last_sync_at?: number;
  next_sync_at?: number;
  error?: string;
}

export interface RawLatestVersion {
  version: string;
  apk_filename: string;
  apk_size: number;
  sha256: string;
  released_at: number;
}

export interface RawVersion {
  version: string;
  apk_filename: string;
  apk_size: number;
  sha256: string;
  released_at: number;
  release_notes: string;
  status: string;
}

export interface RawChangelogEntry {
  version: string;
  date: string;
  type: string;
  notes: string;
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
  version: string;
  install_type: string;
  scheduled_at?: number;
  status: string;
  initiated_by: string;
  initiated_at: number;
  completed_at?: number;
  cancelled_at?: number;
  device_count: number;
  devices?: RawPushDevices;
}

// ============================================================================
// Domain Types
// ============================================================================

export type SyncStatusValue = 'idle' | 'syncing' | 'synced' | 'error';
export type VersionStatus = 'latest' | 'previous' | 'archived';
export type ReleaseType = 'major' | 'minor' | 'patch';
export type UpdateStatus = 'pending' | 'in_progress' | 'completed' | 'failed' | 'cancelled';
export type InstallType = 'immediate' | 'scheduled';

export interface SyncStatus {
  status: SyncStatusValue;
  lastSyncAt: Date | null;
  nextSyncAt: Date | null;
  error: string | null;
}

export interface LatestVersion {
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releasedAt: Date;
}

export interface Version {
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releasedAt: Date;
  releaseNotes: string;
  status: VersionStatus;
}

export interface ChangelogEntry {
  version: string;
  date: string;
  type: ReleaseType;
  notes: string;
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
  scheduledAt?: Date;
  status: UpdateStatus;
  initiatedBy: string;
  initiatedAt: Date;
  completedAt: Date | null;
  cancelledAt: Date | null;
  deviceCount: number;
  devices: PushDevices;
}

export interface UpdateSystemStatus {
  sync: SyncStatus;
  latest: LatestVersion | null;
  device: {
    currentVersion: string | null;
    needsUpdate: boolean;
  };
}

export interface PaginatedVersions {
  versions: Version[];
  pagination: OffsetPagination;
}

export interface PaginatedPushes {
  pushes: UpdatePush[];
  pagination: OffsetPagination;
}

// ============================================================================
// Transform Functions
// ============================================================================

function transformSyncStatus(raw: RawSyncStatus): SyncStatus {
  return {
    status: (raw.status ?? 'idle') as SyncStatusValue,
    lastSyncAt: raw.last_sync_at ? new Date(raw.last_sync_at) : null,
    nextSyncAt: raw.next_sync_at ? new Date(raw.next_sync_at) : null,
    error: raw.error ?? null,
  };
}

function transformLatestVersion(raw: RawLatestVersion): LatestVersion {
  return {
    version: raw.version,
    apkFilename: raw.apk_filename,
    apkSize: raw.apk_size,
    sha256: raw.sha256,
    releasedAt: new Date(raw.released_at),
  };
}

function transformVersion(raw: RawVersion): Version {
  return {
    version: raw.version,
    apkFilename: raw.apk_filename,
    apkSize: raw.apk_size,
    sha256: raw.sha256,
    releasedAt: new Date(raw.released_at),
    releaseNotes: raw.release_notes,
    status: (raw.status ?? 'previous') as VersionStatus,
  };
}

function transformChangelogEntry(raw: RawChangelogEntry): ChangelogEntry {
  return {
    version: raw.version,
    date: raw.date,
    type: (raw.type ?? 'patch') as ReleaseType,
    notes: raw.notes,
  };
}

function transformPushDevices(raw?: RawPushDevices): PushDevices {
  if (!raw) return { total: 0, pending: 0, sent: 0, acknowledged: 0, failed: 0 };
  return {
    total: raw.total,
    pending: raw.pending,
    sent: raw.sent,
    acknowledged: raw.acknowledged,
    failed: raw.failed,
  };
}

function transformUpdatePush(raw: RawUpdatePush): UpdatePush {
  return {
    id: raw.id,
    version: raw.version,
    installType: (raw.install_type ?? 'immediate') as InstallType,
    scheduledAt: raw.scheduled_at ? new Date(raw.scheduled_at) : undefined,
    status: (raw.status ?? 'pending') as UpdateStatus,
    initiatedBy: raw.initiated_by,
    initiatedAt: new Date(raw.initiated_at),
    completedAt: raw.completed_at ? new Date(raw.completed_at) : null,
    cancelledAt: raw.cancelled_at ? new Date(raw.cancelled_at) : null,
    deviceCount: raw.device_count,
    devices: transformPushDevices(raw.devices),
  };
}

function transformPagination(raw: { page: number; limit: number; total: number; total_pages: number }): OffsetPagination {
  return {
    page: raw.page,
    limit: raw.limit,
    total: raw.total,
    totalPages: raw.total_pages,
  };
}

// ============================================================================
// Read Operations
// ============================================================================

export async function fetchUpdateStatus(): Promise<UpdateSystemStatus> {
  const data = await restClient.get<{
    sync: RawSyncStatus;
    latest: RawLatestVersion;
    device: { current_version: string; needs_update: boolean };
  }>(UPDATES_PATHS.status);

  return {
    sync: transformSyncStatus(data.sync),
    latest: data.latest ? transformLatestVersion(data.latest) : null,
    device: {
      currentVersion: data.device.current_version ?? null,
      needsUpdate: data.device.needs_update ?? false,
    },
  };
}

export async function fetchVersions(params?: {
  status?: 'all' | 'latest' | 'previous';
  page?: number;
  limit?: number;
}): Promise<PaginatedVersions> {
  const data = await restClient.get<{
    versions: RawVersion[];
    pagination: { page: number; limit: number; total: number; total_pages: number };
  }>(UPDATES_PATHS.versions, { params });

  return {
    versions: data.versions.map(transformVersion),
    pagination: transformPagination(data.pagination),
  };
}

export async function fetchChangelog(params?: { version?: string }): Promise<ChangelogEntry[]> {
  const data = await restClient.get<{ changelog: RawChangelogEntry[] }>(UPDATES_PATHS.changelog, { params });
  return data.changelog.map(transformChangelogEntry);
}

export async function fetchUpdateHistory(params?: {
  status?: 'all' | 'pending' | 'in_progress' | 'completed' | 'failed' | 'cancelled';
  page?: number;
  limit?: number;
}): Promise<PaginatedPushes> {
  const data = await restClient.get<{
    pushes: RawUpdatePush[];
    pagination: { page: number; limit: number; total: number; total_pages: number };
  }>(UPDATES_PATHS.history, { params });

  return {
    pushes: data.pushes.map(transformUpdatePush),
    pagination: transformPagination(data.pagination),
  };
}

export async function fetchUpdatePushDetail(pushId: string): Promise<UpdatePush | null> {
  const data = await restClient.get<RawUpdatePush>(UPDATES_PATHS.historyDetail(pushId));
  return data ? transformUpdatePush(data) : null;
}

// ============================================================================
// Write Operations
// ============================================================================

export interface PushUpdateRequest {
  version: string;
  deviceIds: string[];
  installType: InstallType;
  scheduledAt?: Date;
}

export interface PushUpdateResponse {
  pushId: string;
  version: string;
  deviceIds: string[];
  installType: InstallType;
  status: UpdateStatus;
  initiatedBy: string;
  initiatedAt: Date;
  devices: PushDevices;
}

export async function pushUpdate(request: PushUpdateRequest): Promise<PushUpdateResponse> {
  const payload = {
    version: request.version,
    device_ids: request.deviceIds,
    install_type: request.installType,
    scheduled_at: request.scheduledAt?.getTime(),
  };

  const data = await restClient.post<{
    push_id: string;
    version: string;
    device_ids: string[];
    install_type: string;
    status: string;
    initiated_by: string;
    initiated_at: number;
    devices: RawPushDevices;
  }>(UPDATES_PATHS.push, payload);

  return {
    pushId: data.push_id,
    version: data.version,
    deviceIds: data.device_ids,
    installType: (data.install_type ?? 'immediate') as InstallType,
    status: 'in_progress',
    initiatedBy: data.initiated_by,
    initiatedAt: new Date(data.initiated_at),
    devices: transformPushDevices(data.devices),
  };
}

export async function cancelUpdate(pushId: string): Promise<UpdatePush> {
  const data = await restClient.post<RawUpdatePush>(UPDATES_PATHS.historyCancel(pushId));
  return transformUpdatePush(data);
}

export interface SyncResult {
  status: SyncStatusValue;
  startedAt: Date;
  versionsFound?: number;
  message?: string;
}

export async function syncFromGitHub(): Promise<SyncResult> {
  const data = await restClient.post<{
    status: string;
    started_at: number;
    versions_found?: number;
    message?: string;
  }>(UPDATES_PATHS.sync);

  return {
    status: (data.status ?? 'synced') as SyncStatusValue,
    startedAt: new Date(data.started_at),
    versionsFound: data.versions_found,
    message: data.message,
  };
}

export async function exportVersions(format: 'json' | 'csv' = 'json'): Promise<Blob> {
  return restClient.get(UPDATES_PATHS.export + `?format=${format}`, { responseType: 'blob' });
}
