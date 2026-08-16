/**
 * Integration tests for logs hooks.
 *
 * These tests render the REAL hooks via React Testing Library. The hooks call
 * the REAL API client functions (logs.list, logs.get, logs.stats) which use the
 * REAL restClient (axios) + domain mappers. MSW intercepts the HTTP requests
 * and returns mock server responses in the raw (snake_case) server shape.
 *
 * The GraphQL fallback path (useDeviceLogs -> fetchDeviceLogsViaGraphQL) is
 * exercised by making MSW return 500 for the REST list endpoint and registering
 * a GraphQL response handler for the GetLogs operation. The real queryLogs
 * function runs against the real Apollo client, which MSW intercepts.
 *
 * No vi.mock / vi.hoisted — the real code path runs end-to-end.
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { waitFor } from '@testing-library/react';
import { http, HttpResponse } from 'msw';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
import { registerGraphQLResponse } from '../msw/vyzor-msw-handlers-graphql';
import { useAuthStore } from '@/stores/auth-store';
import { graphqlClient } from '@vyzorix/api-client';
import { useDeviceLogs, useLog, useLogStats } from '@/hooks/logs/use-logs';

const { server, resetApiState } = setupIntegrationTest();

function setOrg(orgId: string | null) {
  useAuthStore.getState().setOrganization(orgId);
  if (orgId) {
    graphqlClient.setOrganization(orgId);
  }
}

// Make a REST endpoint fail so the hook exercises its GraphQL fallback.
function makeRestFail(method: 'get' | 'post', path: string) {
  server.use(
    http[method](path, () => HttpResponse.json({ error: 'REST down' }, { status: 500 })),
  );
}

describe('useDeviceLogs', () => {
  beforeEach(() => {
    resetApiState();
    setOrg('org-1');
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useDeviceLogs(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled when organizationId is null', () => {
    setOrg(null);
    const { result } = renderHookWithQueryClient(() => useDeviceLogs('123'));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches logs via REST with organizationId', async () => {
    const { result } = renderHookWithQueryClient(() => useDeviceLogs('123', { limit: 50 }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.logs).toHaveLength(2);
    expect(result.current.data?.logs[0]?.id).toBe('log-1');
    expect(result.current.data?.logs[0]?.deviceId).toBe('123');
    expect(result.current.data?.logs[0]?.eventType).toBe('connection');
    expect(result.current.data?.hasMore).toBe(false);
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/dashboard/device/:imei/logs');
    registerGraphQLResponse('GetLogs', () => ({
      deviceLogs: {
        __typename: 'DeviceLogConnection',
        events: [
          {
            __typename: 'LogEntry',
            id: 'log-1',
            type: 'error',
            timestamp: 1704067200,
            data: { msg: 'boom' },
          },
        ],
        pagination: { limit: 50, hasMore: false, nextCursor: null },
      },
    }));
    const { result } = renderHookWithQueryClient(() => useDeviceLogs('123', { limit: 50 }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.logs).toHaveLength(1);
    expect(result.current.data?.logs[0]?.id).toBe('log-1');
    expect(result.current.data?.logs[0]?.eventType).toBe('error');
    expect(result.current.data?.logs[0]?.deviceId).toBe('123');
    expect(result.current.data?.hasMore).toBe(false);
  });
});

describe('useLog', () => {
  beforeEach(() => {
    resetApiState();
    setOrg('org-1');
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useLog(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches a single log with organizationId', async () => {
    const { result } = renderHookWithQueryClient(() => useLog('log-detail-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.id).toBe('log-detail-1');
    expect(result.current.data?.eventType).toBe('info');
    expect(result.current.data?.deviceId).toBe('123');
  });
});

describe('useLogStats', () => {
  beforeEach(() => {
    resetApiState();
    setOrg('org-1');
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useLogStats(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches stats with organizationId', async () => {
    const { result } = renderHookWithQueryClient(() => useLogStats('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.total).toBe(2);
    expect(result.current.data?.byType.connection).toBe(1);
    expect(result.current.data?.byType.command).toBe(1);
    expect(result.current.data?.byType.telemetry).toBe(0);
  });
});
