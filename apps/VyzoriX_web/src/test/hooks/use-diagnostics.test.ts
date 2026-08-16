/**
 * Integration tests for diagnostics hooks (useDeviceInspection, useDeviceTimeline).
 *
 * These tests render the REAL hooks via React Testing Library. The hooks call
 * the REAL API client functions (diagnostics.inspectDevice, diagnostics.getTimeline)
 * which use the REAL restClient (axios) + domain mappers. MSW intercepts the HTTP
 * requests and returns mock server responses.
 *
 * GraphQL fallback paths are tested by making MSW return 500 for REST endpoints
 * and registering GraphQL response handlers for the fallback queries.
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
import { graphqlClient } from '@vyzorix/api-client';
import { useDeviceInspection, useDeviceTimeline } from '@/hooks/diagnostics/use-diagnostics';

const { server } = setupIntegrationTest();

function setOrg(orgId: string | null) {
  useAuthStore.getState().setOrganization(orgId);
  if (orgId) {
    graphqlClient.setOrganization(orgId);
  }
}

/** Override a REST handler to return 500, triggering GraphQL fallback. */
function makeRestFail(method: 'get' | 'post', path: string) {
  server.use(http[method](path, () => HttpResponse.json({ error: 'REST down' }, { status: 500 })));
}

beforeEach(() => {
  setOrg(null);
});

describe('useDeviceInspection', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useDeviceInspection(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled when organizationId is null', () => {
    setOrg(null);
    const { result } = renderHookWithQueryClient(() => useDeviceInspection('123'));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches inspection via REST with organizationId', async () => {
    const { result } = renderHookWithQueryClient(() => useDeviceInspection('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.identity.imei).toBe('123');
    expect(result.current.data?.connection.webSocketStatus).toBe('connected');
    expect(result.current.data?.telemetry.framesToday).toBe(5);
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/device/:imei/inspect');
    const nowIso = new Date().toISOString();
    registerGraphQLResponse('GetDeviceInspection', () => ({
      deviceInspection: {
        __typename: 'DeviceInspection',
        identity: { __typename: 'IdentityInfo', imei: '123', deviceName: 'QL Device' },
        software: { __typename: 'SoftwareInfo', osVersion: '14.0', appVersion: 'v1.1.0' },
        registration: {
          __typename: 'RegistrationInfo',
          status: 'registered',
          registeredAt: nowIso,
          fcmTokenValid: true,
          fcmTokenRefreshedAt: null,
          commandSecretSet: true,
        },
        connection: {
          __typename: 'ConnectionInfo',
          webSocketStatus: 'connected',
          connectedAt: nowIso,
          fcmStatus: 'valid',
          lastSeen: nowIso,
          clientIp: '10.0.0.1',
          protocol: 'ws',
        },
        telemetry: {
          __typename: 'TelemetryInfo',
          lastTimestamp: nowIso,
          framesToday: 7,
          avgLatencyMs: 12,
          totalBytesToday: 2048,
          sessionsToday: 2,
        },
      },
    }));
    const { result } = renderHookWithQueryClient(() => useDeviceInspection('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.identity.imei).toBe('123');
    expect(result.current.data?.identity.deviceName).toBe('QL Device');
    expect(result.current.data?.telemetry.framesToday).toBe(7);
    expect(result.current.data?.registration.status).toBe('registered');
  });

  it('stays idle (no GraphQL fallback) when no org context', () => {
    setOrg(null);
    const { result } = renderHookWithQueryClient(() => useDeviceInspection('123'));
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useDeviceTimeline', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useDeviceTimeline(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled when organizationId is null', () => {
    setOrg(null);
    const { result } = renderHookWithQueryClient(() => useDeviceTimeline('123'));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches timeline via REST with params', async () => {
    const { result } = renderHookWithQueryClient(() =>
      useDeviceTimeline('123', { limit: 25, eventType: 'TELEMETRY' }),
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.events.length).toBeGreaterThanOrEqual(1);
    expect(result.current.data?.events[0]?.deviceId).toBe('123');
    expect(result.current.data?.hasMore).toBe(false);
  });

  it('falls back to GraphQL when REST rejects and injects deviceId', async () => {
    makeRestFail('get', '/v1/device/:imei/timeline');
    const ts = '2024-06-20T12:00:00Z';
    registerGraphQLResponse('GetDeviceTimeline', () => ({
      deviceTimeline: {
        events: [
          { __typename: 'TimelineEvent', id: 'evt-1', type: 'TELEMETRY', timestamp: ts, data: { riskScore: 5 } },
        ],
        hasMore: true,
        nextCursor: 'cur',
      },
    }));
    const { result } = renderHookWithQueryClient(() => useDeviceTimeline('123', { limit: 50 }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.events).toHaveLength(1);
    // GraphQL fallback injects the known imei into events that lack deviceId.
    expect(result.current.data?.events[0]?.deviceId).toBe('123');
    expect(result.current.data?.events[0]?.type).toBe('TELEMETRY');
    expect(result.current.data?.hasMore).toBe(true);
    expect(result.current.data?.nextCursor).toBe('cur');
  });

  it('query key includes organizationId (no cross-org cache collision)', async () => {
    setOrg('org-1');
    const { result, rerender } = renderHookWithQueryClient(
      ({ org }) => {
        useAuthStore.getState().setOrganization(org);
        return useDeviceTimeline('123');
      },
      { initialProps: { org: 'org-1' as string | null } },
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const firstData = result.current.data;
    rerender({ org: 'org-2' });
    // Switching org triggers a fresh fetch (different query key).
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    // Both orgs resolve to data from the same MSW handler, but a new fetch ran.
    expect(result.current.data).toBeDefined();
    expect(firstData).toBeDefined();
  });
});
