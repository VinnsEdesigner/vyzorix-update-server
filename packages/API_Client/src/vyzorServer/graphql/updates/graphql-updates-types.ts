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

export interface RawVersionConnection {
  versions: RawUpdateVersion[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasMore: boolean;
  };
}

export interface RawUpdateHistoryConnection {
  pushes: RawUpdatePush[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    totalPages: number;
    hasMore: boolean;
  };
}

export interface RawChangelogConnection {
  changelog: RawChangelogEntry[];
}

export interface RawPushUpdateResponse {
  success: boolean;
  pushId?: string;
  status?: string;
  error?: string;
}

export interface RawCancelUpdateResponse {
  success: boolean;
  error?: string;
}

export interface RawSyncResponse {
  success: boolean;
  status?: string;
  versionsFound?: number;
  error?: string;
}
