import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  versionFromRaw,
  syncStateFromRaw,
  updatePushFromRaw,
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
  UpdateVersion,
  SyncState,
  UpdatePush,
  VersionListResult,
  UpdateHistoryResult,
  ChangelogEntry,
  PushUpdateRequest,
  UpdateStatus,
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
  latest?: UpdateVersion;
}

export interface VersionParams {
  status?: "all" | "latest" | "previous";
  page?: number;
  limit?: number;
  organizationId?: string;
}

export interface HistoryParams {
  status?: UpdateStatus;
  page?: number;
  limit?: number;
  organizationId?: string;
}

export const updates = {
  async getStatus(organizationId?: string): Promise<UpdateStatusResponse> {
    const response = await restClient.get<{
      sync: RawSyncState;
      latest?: RawVersion;
    }>(PATHS.status, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return {
      sync: syncStateFromRaw(response.sync),
      latest: response.latest ? versionFromRaw(response.latest) : undefined,
    };
  },

  async getVersions(params?: VersionParams): Promise<VersionListResult> {
    const response = await restClient.get<RawVersionListResult>(PATHS.versions, {
      params: {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return versionListResultFromRaw(response);
  },

  async getChangelog(version?: string, organizationId?: string): Promise<ChangelogEntry[]> {
    const response = await restClient.get<{ changelog: { version: string; date: string; type: string; notes: string }[] }>(
      PATHS.changelog,
      { params: { version, organization_id: organizationId || getOrganizationContext() } }
    );
    return response.changelog.map(changelogEntryFromRaw);
  },

  async getHistory(params?: HistoryParams): Promise<UpdateHistoryResult> {
    const response = await restClient.get<RawUpdateHistoryResult>(PATHS.history, {
      params: {
        status: params?.status,
        page: params?.page,
        limit: params?.limit,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return updateHistoryResultFromRaw(response);
  },

  async getPushDetail(pushId: string, organizationId?: string): Promise<UpdatePush | null> {
    const response = await restClient.get<RawUpdatePush | null>(PATHS.historyDetail(pushId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    if (!response) return null;
    return updatePushFromRaw(response);
  },

  async pushUpdate(request: PushUpdateRequest, organizationId?: string): Promise<UpdatePush> {
    const response = await restClient.post<RawUpdatePush>(PATHS.push, {
      version_id: request.versionId,
      device_ids: request.deviceIds,
      install_type: request.installType,
      scheduled_at: request.scheduledAt?.getTime(),
    }, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return updatePushFromRaw(response);
  },

  async cancelPush(pushId: string, organizationId?: string): Promise<UpdatePush> {
    const response = await restClient.post<RawUpdatePush>(PATHS.historyCancel(pushId), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return updatePushFromRaw(response);
  },

  async sync(organizationId?: string): Promise<{ status: string; startedAt: Date; versionsFound?: number }> {
    const response = await restClient.post<{
      status: string;
      started_at: number;
      versions_found?: number;
    }>(PATHS.sync, {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return {
      status: response.status,
      startedAt: new Date(response.started_at),
      versionsFound: response.versions_found,
    };
  },

  async exportVersions(format: "json" | "csv" = "json", organizationId?: string): Promise<Blob> {
    const orgId = organizationId || getOrganizationContext();
    const params = orgId ? `?format=${format}&organization_id=${orgId}` : `?format=${format}`;
    return restClient.get(`${PATHS.export}${params}`, { responseType: "blob" });
  },
};
