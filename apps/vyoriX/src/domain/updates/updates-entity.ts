export type ReleaseType = "major" | "minor" | "patch";

export type UpdateStatus = "pending" | "in_progress" | "completed" | "failed" | "cancelled";

export type InstallType = "immediate" | "scheduled";

export type DevicePushStatus = "pending" | "sent" | "acknowledged" | "failed";

export type SyncStatusValue = "idle" | "syncing" | "synced" | "error";

export type VersionStatus = "latest" | "previous";

export interface Version {
  id: string;
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releaseDate: Date;
  releaseNotes?: string;
  releaseType: ReleaseType;
  isLatest: boolean;
}

export interface ChangelogEntry {
  version: string;
  date: Date;
  type: ReleaseType;
  notes: string;
}

export interface PushDevice {
  id: string;
  deviceId: string;
  deviceName?: string;
  status: DevicePushStatus;
  sentAt?: Date;
  acknowledgedAt?: Date;
  error?: string;
}

export interface UpdatePush {
  id: string;
  version: string;
  installType: InstallType;
  scheduledAt?: Date;
  status: UpdateStatus;
  initiatedBy: string;
  initiatedAt: Date;
  completedAt?: Date;
  cancelledAt?: Date;
  deviceCount: number;
  devices: PushDevice[];
}

export interface SyncStatus {
  status: SyncStatusValue;
  lastSyncAt?: Date;
  nextSyncAt?: Date;
  versionsFound?: number;
  error?: string;
}

export interface DeviceUpdateStatus {
  currentVersion?: string;
  needsUpdate: boolean;
}

export interface UpdateStatusResult {
  sync: SyncStatus;
  latest?: Version;
  device: DeviceUpdateStatus;
}

export interface Pagination {
  page: number;
  limit: number;
  total: number;
  totalPages: number;
}

export interface VersionListResult {
  versions: Version[];
  pagination: Pagination;
}

export interface UpdateHistoryResult {
  pushes: UpdatePush[];
  pagination: Pagination;
}

export interface ChangelogResult {
  changelog: ChangelogEntry[];
}

export interface SyncResult {
  status: SyncStatusValue;
  startedAt: Date;
  versionsFound?: number;
  message?: string;
}

export interface PushUpdateRequest {
  version: string;
  deviceIds: string[];
  installType: InstallType;
  scheduledAt?: Date;
}

export interface ExportFormat {
  format: "json" | "csv" | "pdf";
  fields?: string[];
}

export function getReleaseTypeLabel(type: ReleaseType): string {
  const labels: Record<ReleaseType, string> = {
    major: "Major",
    minor: "Minor",
    patch: "Patch",
  };
  return labels[type];
}

export function getUpdateStatusLabel(status: UpdateStatus): string {
  const labels: Record<UpdateStatus, string> = {
    pending: "Pending",
    in_progress: "In Progress",
    completed: "Completed",
    failed: "Failed",
    cancelled: "Cancelled",
  };
  return labels[status];
}

export function getSyncStatusLabel(status: SyncStatusValue): string {
  const labels: Record<SyncStatusValue, string> = {
    idle: "Idle",
    syncing: "Syncing",
    synced: "Synced",
    error: "Error",
  };
  return labels[status];
}

export function isUpdateCancellable(push: UpdatePush): boolean {
  return push.status === "pending" || push.status === "in_progress";
}

export function formatApkSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}
