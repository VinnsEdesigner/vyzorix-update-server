export interface RawUpdateVersion {
  __typename?: "UpdateVersion";
  id: string;
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releasedAt: string;
  releaseNotes?: string;
  releaseType: string;
  isLatest: boolean;
  createdAt?: string;
  updatedAt?: string;
}

export interface RawPushDevice {
  __typename?: "PushDevice";
  id: string;
  deviceId: string;
  deviceName: string;
  status: string;
  sentAt?: string | null;
  acknowledgedAt?: string | null;
  error?: string;
}

export interface RawUpdatePush {
  __typename?: "UpdatePush";
  id: string;
  version: string;
  installType: string;
  status: string;
  initiatedBy: string;
  initiatedAt: string;
  scheduledAt?: string | null;
  completedAt?: string | null;
  cancelledAt?: string | null;
  deviceCount: number;
  devices?: RawPushDevice[] | null;
}

export interface RawPushHistoryEntry {
  __typename?: "PushHistoryEntry";
  id: string;
  version: string;
  installType: string;
  status: string;
  initiatedBy: string;
  initiatedAt: number;
  completedAt?: number | null;
  deviceCount: number;
  pending: number;
  acknowledged: number;
  failed: number;
}

export interface RawSyncStatus {
  __typename?: "SyncStatus";
  status: string;
  lastSyncAt?: string | null;
  nextSyncAt?: string | null;
  versionsFound?: number;
  error?: string;
}

export interface RawChangelogEntry {
  __typename?: "ChangelogEntry";
  version: string;
  date: string;
  type: string;
  notes: string;
}

export interface RawDeviceUpdateStatus {
  __typename?: "DeviceUpdateStatus";
  currentVersion: string;
  needsUpdate: boolean;
}

export interface RawVersionConnection {
  versions: RawUpdateVersion[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    hasMore: boolean;
  };
}

export interface RawUpdateHistoryConnection {
  pushes: RawPushHistoryEntry[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    hasMore: boolean;
  };
}

export interface RawUpdateStatusResponse {
  version: string;
  sync: RawSyncStatus;
  latest?: RawUpdateVersion | null;
  device: RawDeviceUpdateStatus;
  apkFilename?: string | null;
  sha256?: string | null;
}

export interface RawChangelogConnection {
  changelog: RawChangelogEntry[];
}

export interface RawPushUpdateResponse {
  pushId: string;
  version: string;
  installType: string;
  scheduledAt?: number;
  status: string;
  initiatedBy: string;
  initiatedAt: number;
  deviceCount: number;
}

export interface RawCancelUpdateResponse {
  id: string;
  status: string;
  cancelledAt: number;
  cancelledBy: string;
}

export interface RawSyncResponse {
  status: string;
  startedAt: number;
  message?: string;
  versionsFound?: number;
}
