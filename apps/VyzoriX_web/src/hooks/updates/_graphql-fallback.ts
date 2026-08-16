import {
  queryUpdates,
  queryUpdatesStatus,
  queryUpdatesChangelog,
  queryUpdatesHistory,
  mutatePushUpdate,
  mutateCancelUpdate,
  mutateSyncUpdates,
  type UpdateVersion,
  type UpdatePush,
  type VersionListResult,
  type UpdateHistoryResult,
  type ChangelogEntry,
  type SyncState,
  type PushDevices,
  type InstallType,
  type UpdateStatus,
  type UpdateStatusResponse,
} from '@vyzorix/api-client';

interface VersionFallbackParams {
  status?: string;
  limit?: number;
  offset?: number;
}

interface HistoryFallbackParams {
  status?: string;
  page?: number;
  limit?: number;
}

interface PushFallbackParams {
  version: string;
  deviceIds: string[];
  installType: InstallType;
  scheduledAt?: Date;
}

const EMPTY_DEVICES: PushDevices = {
  total: 0,
  pending: 0,
  sent: 0,
  acknowledged: 0,
  failed: 0,
};

export async function fetchVersionsViaGraphQL(
  organizationId: string,
  params?: VersionFallbackParams,
): Promise<VersionListResult> {
  return queryUpdates({
    organizationId,
    status: params?.status,
    limit: params?.limit,
    offset: params?.offset,
  });
}

export async function fetchUpdateStatusViaGraphQL(
  organizationId: string,
): Promise<UpdateStatusResponse> {
  const result = await queryUpdatesStatus({ organizationId });
  return {
    sync: result.sync,
    latest: result.latest,
  };
}

export async function fetchChangelogViaGraphQL(
  organizationId: string,
  version?: string,
): Promise<ChangelogEntry[]> {
  return queryUpdatesChangelog({ organizationId, version });
}

export async function fetchUpdateHistoryViaGraphQL(
  organizationId: string,
  params?: HistoryFallbackParams,
): Promise<UpdateHistoryResult> {
  return queryUpdatesHistory({
    organizationId,
    status: params?.status,
    page: params?.page,
    limit: params?.limit,
  });
}

export async function pushUpdateViaGraphQL(
  organizationId: string,
  params: PushFallbackParams,
): Promise<UpdatePush> {
  const result = await mutatePushUpdate({
    organizationId,
    version: params.version,
    deviceIds: params.deviceIds,
    installType: params.installType,
    scheduledAt: params.scheduledAt?.getTime(),
  });
  return {
    id: result.pushId,
    version: result.version,
    installType: result.installType,
    status: result.status,
    initiatedBy: result.initiatedBy,
    initiatedAt: result.initiatedAt,
    scheduledAt: result.scheduledAt,
    devices: { ...EMPTY_DEVICES, total: result.deviceCount },
  };
}

export async function cancelUpdateViaGraphQL(
  organizationId: string,
  pushId: string,
): Promise<UpdatePush> {
  const result = await mutateCancelUpdate({ organizationId, id: pushId });
  return {
    id: result.id,
    version: '',
    installType: 'immediate',
    status: result.status,
    initiatedBy: '',
    initiatedAt: result.cancelledAt,
    cancelledAt: result.cancelledAt,
    cancelledBy: result.cancelledBy,
    devices: { ...EMPTY_DEVICES },
  };
}

export async function syncUpdatesViaGraphQL(
  organizationId: string,
): Promise<{ status: string; startedAt: Date; versionsFound?: number }> {
  const result = await mutateSyncUpdates({ organizationId });
  return {
    status: result.status,
    startedAt: result.startedAt,
    versionsFound: result.versionsFound,
  };
}

export type {
  UpdateVersion,
  UpdatePush,
  VersionListResult,
  UpdateHistoryResult,
  ChangelogEntry,
  SyncState,
  PushDevices,
  InstallType,
  UpdateStatus,
  UpdateStatusResponse,
};
