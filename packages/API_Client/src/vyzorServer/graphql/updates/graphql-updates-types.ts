export type RawUpdateVersion = {
  __typename?: "UpdateVersion";
  id: string;
  version: string;
  apkFilename: string;
  apkSize: number;
  sha256: string;
  releaseDate: string;
  releaseNotes: string;
  releaseType: string;
  isLatest: boolean;
};

export type RawPushDevice = {
  __typename?: "PushDevice";
  id: string;
  deviceId: string;
  deviceName: string;
  status: string;
  sentAt?: string | null;
  acknowledgedAt?: string | null;
  error?: string;
};

export type RawUpdatePush = {
  __typename?: "UpdatePush";
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
  __typename?: "SyncStatus";
  status: string;
  lastSyncAt?: string | null;
  nextSyncAt?: string | null;
  versionsFound?: number;
  error?: string;
};

export type RawChangelogEntry = {
  __typename?: "ChangelogEntry";
  version: string;
  date: string;
  type: string;
  notes: string;
};

export type RawDeviceUpdateStatus = {
  __typename?: "DeviceUpdateStatus";
  currentVersion: string;
  needsUpdate: boolean;
};

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
  pushes: RawUpdatePush[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
    hasMore: boolean;
  };
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
