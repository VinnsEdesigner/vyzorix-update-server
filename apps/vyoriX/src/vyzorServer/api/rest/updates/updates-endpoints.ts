/**
 * Updates REST Endpoints
 * 
 * REST API client for firmware updates operations.
 * Based on SERVER_BACKEND_UPDATES_API.md specification.
 * Uses session-based authentication (cookies).
 */

import { apiGet, apiPost } from "../_shared/rest-client";
import type { PaginatedResult } from "@/domain/_shared";
import { offsetPaginationFromRaw } from "@/domain/_shared";

// ============================================================================
// API Paths
// ============================================================================

/**
 * Updates API paths
 */
export const UPDATES_PATHS = {
  // GET /v1/updates/status - Current sync status
  status: "/v1/updates/status",
  // GET /v1/updates/versions - All available versions
  versions: "/v1/updates/versions",
  // GET /v1/updates/changelog - Release changelog
  changelog: "/v1/updates/changelog",
  // POST /v1/updates/push - Push update to devices
  push: "/v1/updates/push",
  // GET /v1/updates/history - Update push history
  history: "/v1/updates/history",
  // GET /v1/updates/history/:pushId - Single push detail
  historyDetail: (pushId: string) => `/v1/updates/history/${pushId}`,
  // POST /v1/updates/history/:pushId/cancel - Cancel push
  historyCancel: (pushId: string) => `/v1/updates/history/${pushId}/cancel`,
  // GET /v1/updates/export - Export version data
  export: "/v1/updates/export",
  // POST /v1/updates/sync - Sync from GitHub
  sync: "/v1/updates/sync",
} as const;

// ============================================================================
// Raw Types
// ============================================================================

/**
 * Raw sync status from API
 */
export interface RawSyncStatus {
  status: string;
  last_sync_at?: number;
  next_sync_at?: number;
  error?: string;
}

/**
 * Raw latest version from API
 */
export interface RawLatestVersion {
  version: string;
  apk_filename: string;
  apk_size: number;
  sha256: string;
  released_at: number;
}

/**
 * Raw version from API
 */
export interface RawVersion {
  version: string;
  apk_filename: string;
  apk_size: number;
  sha256: string;
  released_at: number;
  release_notes: string;
  status: string;
}

/**
 * Raw changelog entry from API
 */
export interface RawChangelogEntry {
  version: string;
  date: string;
  type: string;
  notes: string;
}

/**
 * Raw push devices summary from API
 */
export interface RawPushDevices {
  total: number;
  pending: number;
  sent: number;
  acknowledged: number;
  failed: number;
}

/**
 * Raw update push from API
 */
export interface RawUpdatePush {
  id: string;
  version: string;
  install_type: string;
  status: string;
  initiated_by: string;
  initiated_at: number;
  completed_at?: number;
  cancelled_at?: number;
  device_count: number;
  devices?: RawPushDevices;
}

// ============================================================================
// Transformed Types
// ============================================================================

/**
 * Sync status
 */
export interface SyncStatus {
  status: "idle" | "syncing" | "synced" | "error";
  lastSyncAt: Date | null;
  nextSyncAt: Date | null;
  error: string | null;
}

/**
 * Latest version info
 */
export interface LatestVersion {
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releasedAt: Date;
}

/**
 * Update version
 */
export interface UpdateVersion {
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releasedAt: Date;
  releaseNotes: string;
  status: "latest" | "previous" | "archived";
}

/**
 * Release type
 */
export type ReleaseType = "major" | "minor" | "patch";

/**
 * Changelog entry
 */
export interface ChangelogEntry {
  version: string;
  date: string;
  type: ReleaseType;
  notes: string;
}

/**
 * Push device status summary
 */
export interface PushDevices {
  total: number;
  pending: number;
  sent: number;
  acknowledged: number;
  failed: number;
}

/**
 * Update push
 */
export interface UpdatePush {
  id: string;
  version: string;
  installType: "immediate" | "scheduled";
  status: "pending" | "in_progress" | "completed" | "failed" | "cancelled";
  initiatedBy: string;
  initiatedAt: Date;
  completedAt: Date | null;
  cancelledAt: Date | null;
  deviceCount: number;
  devices: PushDevices;
}

/**
 * Update system status
 */
export interface UpdateSystemStatus {
  sync: SyncStatus;
  latest: LatestVersion | null;
  device: {
    currentVersion: string | null;
    needsUpdate: boolean;
  };
}

/**
 * Install type
 */
export type InstallType = "immediate" | "scheduled";

// ============================================================================
// Transform Functions
// ============================================================================

function syncStatusFromRaw(raw: RawSyncStatus): SyncStatus {
  return {
    status: (raw.status as SyncStatus["status"]) ?? "idle",
    lastSyncAt: raw.last_sync_at ? new Date(raw.last_sync_at) : null,
    nextSyncAt: raw.next_sync_at ? new Date(raw.next_sync_at) : null,
    error: raw.error ?? null,
  };
}

function latestVersionFromRaw(raw: RawLatestVersion): LatestVersion {
  return {
    version: raw.version,
    apkFilename: raw.apk_filename,
    apkSize: raw.apk_size,
    sha256: raw.sha256,
    releasedAt: new Date(raw.released_at),
  };
}

function versionFromRaw(raw: RawVersion): UpdateVersion {
  return {
    version: raw.version,
    apkFilename: raw.apk_filename,
    apkSize: raw.apk_size,
    sha256: raw.sha256,
    releasedAt: new Date(raw.released_at),
    releaseNotes: raw.release_notes,
    status: (raw.status as UpdateVersion["status"]) ?? "previous",
  };
}

function changelogEntryFromRaw(raw: RawChangelogEntry): ChangelogEntry {
  return {
    version: raw.version,
    date: raw.date,
    type: (raw.type as ReleaseType) ?? "patch",
    notes: raw.notes,
  };
}

function pushDevicesFromRaw(raw?: RawPushDevices): PushDevices {
  if (!raw) {
    return { total: 0, pending: 0, sent: 0, acknowledged: 0, failed: 0 };
  }
  return {
    total: raw.total,
    pending: raw.pending,
    sent: raw.sent,
    acknowledged: raw.acknowledged,
    failed: raw.failed,
  };
}

function updatePushFromRaw(raw: RawUpdatePush): UpdatePush {
  return {
    id: raw.id,
    version: raw.version,
    installType: (raw.install_type as UpdatePush["installType"]) ?? "immediate",
    status: (raw.status as UpdatePush["status"]) ?? "pending",
    initiatedBy: raw.initiated_by,
    initiatedAt: new Date(raw.initiated_at),
    completedAt: raw.completed_at ? new Date(raw.completed_at) : null,
    cancelledAt: raw.cancelled_at ? new Date(raw.cancelled_at) : null,
    deviceCount: raw.device_count,
    devices: pushDevicesFromRaw(raw.devices),
  };
}

// ============================================================================
// Fetch Operations
// ============================================================================

/**
 * Fetch update system status
 * GET /v1/updates/status
 */
export async function fetchUpdateStatus(): Promise<UpdateSystemStatus> {
  const data = await apiGet<{
    sync: RawSyncStatus;
    latest: RawLatestVersion;
    device: { current_version: string; needs_update: boolean };
  }>(UPDATES_PATHS.status);

  return {
    sync: syncStatusFromRaw(data.sync),
    latest: data.latest ? latestVersionFromRaw(data.latest) : null,
    device: {
      currentVersion: data.device.current_version ?? null,
      needsUpdate: data.device.needs_update ?? false,
    },
  };
}

/**
 * Fetch available versions
 * GET /v1/updates/versions
 */
export async function fetchVersions(params?: {
  status?: "all" | "latest" | "previous";
  page?: number;
  limit?: number;
}): Promise<PaginatedResult<UpdateVersion[]>> {
  const data = await apiGet<{
    versions: RawVersion[];
    pagination: {
      page: number;
      limit: number;
      total: number;
      total_pages: number;
    };
  }>(UPDATES_PATHS.versions, {
    status: params?.status,
    page: params?.page,
    limit: params?.limit,
  });

  return {
    items: data.versions.map(versionFromRaw),
    pagination: offsetPaginationFromRaw(data.pagination),
  };
}

/**
 * Fetch changelog
 * GET /v1/updates/changelog
 */
export async function fetchChangelog(params?: {
  version?: string;
}): Promise<ChangelogEntry[]> {
  const data = await apiGet<{ changelog: RawChangelogEntry[] }>(UPDATES_PATHS.changelog, {
    version: params?.version,
  });

  return data.changelog.map(changelogEntryFromRaw);
}

/**
 * Fetch update history
 * GET /v1/updates/history
 */
export async function fetchUpdateHistory(params?: {
  status?: "all" | "pending" | "in_progress" | "completed" | "failed" | "cancelled";
  page?: number;
  limit?: number;
}): Promise<PaginatedResult<UpdatePush[]>> {
  const data = await apiGet<{
    pushes: RawUpdatePush[];
    pagination: {
      page: number;
      limit: number;
      total: number;
      total_pages: number;
    };
  }>(UPDATES_PATHS.history, {
    status: params?.status,
    page: params?.page,
    limit: params?.limit,
  });

  return {
    items: data.pushes.map(updatePushFromRaw),
    pagination: offsetPaginationFromRaw(data.pagination),
  };
}

/**
 * Fetch single push detail
 * GET /v1/updates/history/:pushId
 */
export async function fetchUpdatePushDetail(pushId: string): Promise<UpdatePush | null> {
  const data = await apiGet<RawUpdatePush>(UPDATES_PATHS.historyDetail(pushId));
  if (!data) return null;
  return updatePushFromRaw(data);
}

// ============================================================================
// Write Operations
// ============================================================================

/**
 * Push update to devices
 * POST /v1/updates/push
 */
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
  status: "in_progress";
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

  const data = await apiPost<{
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
    installType: (data.install_type as InstallType) ?? "immediate",
    status: "in_progress",
    initiatedBy: data.initiated_by,
    initiatedAt: new Date(data.initiated_at),
    devices: pushDevicesFromRaw(data.devices),
  };
}

/**
 * Cancel update push
 * POST /v1/updates/history/:pushId/cancel
 */
export async function cancelUpdate(pushId: string): Promise<UpdatePush> {
  const data = await apiPost<RawUpdatePush>(UPDATES_PATHS.historyCancel(pushId));
  return updatePushFromRaw(data);
}

/**
 * Sync versions from GitHub
 * POST /v1/updates/sync
 */
export interface SyncResult {
  status: string;
  startedAt: Date;
  versionsFound?: number;
  message?: string;
}

export async function syncFromGitHub(): Promise<SyncResult> {
  const data = await apiPost<{
    status: string;
    started_at: number;
    versions_found?: number;
    message?: string;
  }>(UPDATES_PATHS.sync);

  return {
    status: data.status,
    startedAt: new Date(data.started_at),
    versionsFound: data.versions_found,
    message: data.message,
  };
}

/**
 * Export versions data
 * GET /v1/updates/export
 */
export interface ExportVersionsRequest {
  format?: "json" | "csv";
}

export async function exportVersions(params?: ExportVersionsRequest): Promise<Blob> {
  const response = await fetch(`/api${UPDATES_PATHS.export}?format=${params?.format ?? "json"}`, {
    credentials: "include",
  });

  return response.blob();
}
