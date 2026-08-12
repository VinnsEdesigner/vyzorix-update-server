import type { Pagination } from "../_shared";
export type ReleaseType = "major" | "minor" | "patch";
export type UpdateStatus = "pending" | "in_progress" | "completed" | "failed" | "cancelled";
export type InstallType = "immediate" | "scheduled";
export type DevicePushStatus = "pending" | "sent" | "in_progress" | "acknowledged" | "completed" | "failed";
export type SyncStatus = "idle" | "syncing" | "synced" | "error";

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
  versionId: string;
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

export type { Pagination } from "../_shared";

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

export interface PushUpdateRequest {
  versionId: string;
  deviceIds: string[];
  installType: InstallType;
  scheduledAt?: Date;
}

export function formatApkSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

export function getReleaseTypeLabel(type: ReleaseType): string {
  const labels: Record<ReleaseType, string> = { major: "Major", minor: "Minor", patch: "Patch" };
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
