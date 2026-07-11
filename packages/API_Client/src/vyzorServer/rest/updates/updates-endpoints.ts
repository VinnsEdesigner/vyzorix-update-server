import { restClient } from "../_shared/rest-client";
import {
  versionFromRaw,
  syncStateFromRaw,
  updatePushFromRaw,
  paginationFromRaw,
  versionListResultFromRaw,
  updateHistoryResultFromRaw,
  changelogEntryFromRaw,
  type RawVersion,
  type RawSyncState,
  type RawUpdatePush,
  type RawVersionListResult,
  type RawUpdateHistoryResult,
} from "@/domain/updates";
import type {
  Version,
  SyncState,
  UpdatePush,
  VersionListResult,
  UpdateHistoryResult,
  ChangelogEntry,
  PushUpdateRequest,
  UpdateStatus,
  InstallType,
} from "@/domain/updates";

const PATHS = {
  status: "/v1/updates/status",
  versions: "/v1/updates/versions",
  changelog: "/v1/updates/changelog",
  push: "/v1/updates/push",
  history: "/v1/updates/history",
  historyDetail: (pushId: string) => `/v1/updates/history/${pushId}`,
  historyCancel: (pushId: string) => `/v1/updates/history/${pushId}/cancel`,
  sync: "/v1/updates/sync",
  export: "/v1/updates/export",
} as const;

export interface UpdateStatusResponse {
  sync: SyncState;
  latest?: Version;
}

export interface VersionParams {
  status?: "all" | "latest" | "previous";
  page?: number;
  limit?: number;
}

export interface HistoryParams {
  status?: UpdateStatus;
  page?: number;
  limit?: number;
}

export const updates = {
  async getStatus(): Promise<UpdateStatusResponse> {
    const response = await restClient.get<{
      sync: RawSyncState;
      latest?: RawVersion;
    }>(PATHS.status);
    return {
      sync: syncStateFromRaw(response.sync),
      latest: response.latest ? versionFromRaw(response.latest) : undefined,
    };
  },

  async getVersions(params?: VersionParams): Promise<VersionListResult> {
    const response = await restClient.get<RawVersionListResult>(PATHS.versions, { params });
    return versionListResultFromRaw(response);
  },

  async getChangelog(version?: string): Promise<ChangelogEntry[]> {
    const response = await restClient.get<{ changelog: { version: string; date: string; type: string; notes: string }[] }>(
      PATHS.changelog,
      { params: { version } }
    );
    return response.changelog.map(changelogEntryFromRaw);
  },

  async getHistory(params?: HistoryParams): Promise<UpdateHistoryResult> {
    const response = await restClient.get<RawUpdateHistoryResult>(PATHS.history, { params });
    return updateHistoryResultFromRaw(response);
  },

  async getPushDetail(pushId: string): Promise<UpdatePush | null> {
    const response = await restClient.get<RawUpdatePush | null>(PATHS.historyDetail(pushId));
    if (!response) return null;
    return updatePushFromRaw(response);
  },

  async pushUpdate(request: PushUpdateRequest): Promise<UpdatePush> {
    const response = await restClient.post<RawUpdatePush>(PATHS.push, {
      version_id: request.versionId,
      device_ids: request.deviceIds,
      install_type: request.installType,
      scheduled_at: request.scheduledAt?.getTime(),
    });
    return updatePushFromRaw(response);
  },

  async cancelPush(pushId: string): Promise<UpdatePush> {
    const response = await restClient.post<RawUpdatePush>(PATHS.historyCancel(pushId));
    return updatePushFromRaw(response);
  },

  async sync(): Promise<{ status: string; startedAt: Date; versionsFound?: number }> {
    const response = await restClient.post<{
      status: string;
      started_at: number;
      versions_found?: number;
    }>(PATHS.sync);
    return {
      status: response.status,
      startedAt: new Date(response.started_at),
      versionsFound: response.versions_found,
    };
  },

  async exportVersions(format: "json" | "csv" = "json"): Promise<Blob> {
    return restClient.get(`${PATHS.export}?format=${format}`, { responseType: "blob" });
  },
};
