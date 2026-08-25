/**
 * Integration tests for updates hooks.
 *
 * These tests render the REAL hooks via React Testing Library. The hooks call
 * the REAL API client functions (updates.getVersions, updates.getStatus, etc.)
 * which use the REAL restClient (axios) + domain mappers. MSW intercepts the
 * HTTP requests and returns mock server responses.
 *
 * GraphQL fallback paths are tested by making MSW return 500 for REST endpoints
 * and registering GraphQL response handlers for the fallback queries/mutations.
 *
 * No vi.mock — the real code path runs end-to-end.
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
import { registerGraphQLResponse } from '../msw/vyzor-msw-handlers-graphql';
import { useAuthStore } from '@/stores/auth-store';
import { useUpdatesStore } from '@/stores';
import { graphqlClient } from '@vyzorix/api-client';
import {
  useVersions,
  useUpdateStatus,
  useChangelog,
  useUpdateHistory,
  usePushDetail,
  usePushUpdate,
  useCancelUpdate,
  useSyncUpdates,
} from '@/hooks/updates';

const { server } = setupIntegrationTest();

function setOrg(orgId: string | null) {
  useAuthStore.getState().setOrganization(orgId);
  if (orgId) {
    graphqlClient.setOrganization(orgId);
  }
}

// --- GraphQL response fixtures ---

function makeGraphQLVersion(version: string = 'v1.2.0', isLatest = true) {
  const now = new Date().toISOString();
  return {
    __typename: 'UpdateVersion',
    id: `ver-${version}`,
    version,
    apkFilename: 'app.apk',
    apkSize: 1000,
    sha256: 'abc',
    releaseType: 'minor',
    releaseNotes: 'Bug fixes',
    releasedAt: now,
    createdAt: now,
    isLatest,
  };
}

function makeGraphQLPushHistory(id: string = 'push-1', status: string = 'pending') {
  const now = Date.now();
  return {
    __typename: 'PushHistoryEntry',
    id,
    version: 'v1.2.0',
    installType: 'immediate',
    status,
    initiatedBy: 'op-1',
    initiatedAt: now,
    completedAt: null,
    deviceCount: 2,
    pending: 2,
    acknowledged: 0,
    failed: 0,
  };
}

/** Override a REST handler to return 500, triggering GraphQL fallback. */
function makeRestFail(method: 'get' | 'post', path: string) {
  server.use(http[method](path, () => HttpResponse.json({ error: 'REST down' }, { status: 500 })));
}

beforeEach(() => {
  setOrg(null);
});

describe('useVersions', () => {
  beforeEach(() => setOrg('org-1'));

  it('fetches versions via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useVersions({ status: 'latest' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.versions).toHaveLength(3);
    expect(result.current.data?.versions[0]?.version).toBe('v1.2.0');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/updates/versions');
    registerGraphQLResponse('GetUpdates', () => ({
      updatesVersions: {
        versions: [makeGraphQLVersion('v3.0.0')],
        pagination: { total: 1, limit: 20, offset: 0, hasMore: false },
      },
    }));
    const { result } = renderHookWithQueryClient(() => useVersions());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.versions).toHaveLength(1);
    expect(result.current.data?.versions[0]?.version).toBe('v3.0.0');
  });

  it('is disabled when organizationId is null', () => {
    setOrg(null);
    const { result } = renderHookWithQueryClient(() => useVersions());
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useUpdateStatus', () => {
  beforeEach(() => setOrg('org-1'));

  it('fetches status via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useUpdateStatus());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.sync).toBeDefined();
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/updates/status');
    registerGraphQLResponse('GetUpdatesStatus', () => ({
      updatesStatus: {
        version: 'v1.0.0',
        sync: {
          __typename: 'SyncStatus',
          status: 'synced',
          lastSyncAt: new Date().toISOString(),
          nextSyncAt: null,
          versionsFound: 3,
          error: null,
        },
        latest: null,
        device: {
          __typename: 'DeviceUpdateStatus',
          currentVersion: 'v1.0.0',
          needsUpdate: false,
        },
        apkFilename: null,
        sha256: null,
      },
    }));
    const { result } = renderHookWithQueryClient(() => useUpdateStatus());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.sync.status).toBe('synced');
    expect(result.current.data?.sync.versionsFound).toBe(3);
  });
});

describe('useChangelog', () => {
  beforeEach(() => setOrg('org-1'));

  it('fetches changelog via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useChangelog('v1.2.0'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(3);
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/updates/changelog');
    registerGraphQLResponse('GetUpdatesChangelog', () => ({
      updatesChangelog: [
        { __typename: 'ChangelogEntry', version: 'v1.2.0', date: new Date().toISOString(), type: 'minor', notes: 'GraphQL fixes' },
      ],
    }));
    const { result } = renderHookWithQueryClient(() => useChangelog());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]?.notes).toBe('GraphQL fixes');
  });
});

describe('useUpdateHistory', () => {
  beforeEach(() => setOrg('org-1'));

  it('fetches history via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useUpdateHistory({ page: 1 }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.pushes).toHaveLength(2);
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/updates/history');
    registerGraphQLResponse('GetUpdatesHistory', () => ({
      updatesHistory: {
        pushes: [makeGraphQLPushHistory('push-gql-1', 'pending')],
        pagination: { total: 1, limit: 20, offset: 0, hasMore: false },
      },
    }));
    const { result } = renderHookWithQueryClient(() => useUpdateHistory());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.pushes).toHaveLength(1);
    expect(result.current.data?.pushes[0]?.id).toBe('push-gql-1');
  });
});

describe('usePushDetail', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when pushId is undefined', () => {
    const { result } = renderHookWithQueryClient(() => usePushDetail(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches push detail by id', async () => {
    const { result } = renderHookWithQueryClient(() => usePushDetail('push-test-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.id).toBe('push-test-1');
  });
});

describe('usePushUpdate', () => {
  beforeEach(() => {
    setOrg('org-1');
    useUpdatesStore.setState({
      draft: { version: 'v1.2.0', deviceIds: ['dev-1'], installType: 'immediate', scheduledAt: null },
      lastSyncStatus: 'idle',
      lastSyncAt: null,
      isSyncing: false,
    });
  });

  it('pushes via REST and resets draft on success', async () => {
    const { result } = renderHookWithQueryClient(() => usePushUpdate());
    result.current.mutate({ version: 'v1.2.0', deviceIds: ['dev-1'], installType: 'immediate' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.version).toBe('v1.2.0');
    expect(useUpdatesStore.getState().draft.version).toBe('');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('post', '/v1/updates/push');
    registerGraphQLResponse('PushUpdate', () => ({
      pushUpdate: {
        pushId: 'push-gql-1',
        version: 'v1.2.0',
        installType: 'immediate',
        scheduledAt: null,
        status: 'pending',
        initiatedBy: 'op-1',
        initiatedAt: Date.now(),
        deviceCount: 1,
      },
    }));
    const { result } = renderHookWithQueryClient(() => usePushUpdate());
    result.current.mutate({ version: 'v1.2.0', deviceIds: ['dev-1'], installType: 'immediate' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.id).toBe('push-gql-1');
    expect(result.current.data?.devices.total).toBe(1);
  });
});

describe('useCancelUpdate', () => {
  beforeEach(() => setOrg('org-1'));

  it('cancels via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useCancelUpdate());
    result.current.mutate({ pushId: 'push-test-1' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.status).toBe('cancelled');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('post', '/v1/updates/history/:pushId/cancel');
    registerGraphQLResponse('CancelUpdate', () => ({
      cancelUpdate: {
        id: 'push-gql-1',
        status: 'cancelled',
        cancelledAt: Date.now(),
        cancelledBy: 'op-1',
      },
    }));
    const { result } = renderHookWithQueryClient(() => useCancelUpdate());
    result.current.mutate({ pushId: 'push-gql-1' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.id).toBe('push-gql-1');
    expect(result.current.data?.status).toBe('cancelled');
  });
});

describe('useSyncUpdates', () => {
  beforeEach(() => setOrg('org-1'));

  it('syncs via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useSyncUpdates());
    result.current.mutate();
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.status).toBeDefined();
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('post', '/v1/updates/sync');
    registerGraphQLResponse('SyncUpdates', () => ({
      syncUpdates: {
        status: 'started',
        startedAt: Date.now(),
        versionsFound: 2,
      },
    }));
    const { result } = renderHookWithQueryClient(() => useSyncUpdates());
    result.current.mutate();
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.versionsFound).toBe(2);
  });
});
