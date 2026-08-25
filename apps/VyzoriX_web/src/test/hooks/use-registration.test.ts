/**
 * Integration tests for registration hooks.
 *
 * These tests render the REAL hooks via React Testing Library. The hooks call
 * the REAL `registration` API client (real restClient/axios + domain mappers)
 * and the REAL GraphQL fallback functions (real Apollo Client). MSW intercepts
 * the HTTP requests and returns mock server responses mirroring the Go server
 * contract.
 *
 * GraphQL fallback paths are exercised by making MSW return 500 for the REST
 * endpoint and registering a GraphQL response handler for the fallback
 * query/mutation (the operation name the real Apollo request carries).
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
import {
  useInbox,
  useInboxEntry,
  useAcknowledgeInbox,
  useCreateInboxRequest,
  useConfirmDevice,
  useDismissInbox,
  useResendInboxApproval,
  useRegisteredDevices,
  useRegisteredDevice,
  useDeregisterRegisteredDevice,
  useRegistrationStatus,
  useRegisterDevice,
} from '@/hooks/registration';

const { server } = setupIntegrationTest();

function setOrg(orgId: string | null) {
  useAuthStore.getState().setOrganization(orgId);
  if (orgId) {
    graphqlClient.setOrganization(orgId);
  }
}

// --- REST response fixtures (raw server shapes; domain mappers normalize) ---

const now = Date.now();

const RAW_INBOX_ENTRY = {
  id: 'e1',
  imei: '123',
  deviceName: 'Dev',
  deviceClass: 'phone',
  model: 'X',
  manufacturer: 'Acme',
  osVersion: '14',
  appVersion: '1.0',
  fcmToken: 'tok',
  firebaseInstallId: 'fid',
  status: 'pending',
  acknowledgedAt: null,
  approvingAt: null,
  approvedAt: null,
  rejectedAt: null,
  notes: null,
  operatorId: null,
  createdAt: now,
};

const RAW_REGISTERED_DEVICE = {
  id: 'dev-1',
  imei: '123',
  deviceName: 'Dev',
  model: 'X',
  manufacturer: 'Acme',
  osVersion: '14',
  appVersion: '1.0',
  status: 'online',
  registeredAt: now,
  lastSeen: now,
  online: true,
};

/** Make MSW return 500 for a REST path, triggering the GraphQL fallback. */
function makeRestFail(method: 'get' | 'post' | 'delete', path: string) {
  server.use(http[method](path, () => HttpResponse.json({ error: 'REST down' }, { status: 500 })));
}

/**
 * Override the shared `/v1/devices*` handlers (used by the devices client) with
 * the registered-device raw shape the registration client expects. The
 * registration REST client reuses these paths but consumes a different raw
 * shape than the devices client, so registration tests scope them per-test.
 */
function overrideRegisteredDeviceHandlers() {
  server.use(
    http.get('/v1/devices', async ({ request }) => {
      const url = new URL(request.url);
      return HttpResponse.json({
        devices: [RAW_REGISTERED_DEVICE],
        pagination: {
          page: Number(url.searchParams.get('page') ?? '1'),
          limit: Number(url.searchParams.get('limit') ?? '20'),
          total: 1,
          totalPages: 1,
        },
      });
    }),
    http.get('/v1/devices/:imei', async ({ params }) =>
      HttpResponse.json({ ...RAW_REGISTERED_DEVICE, imei: params.imei as string }),
    ),
    http.delete('/v1/devices/:imei', async ({ params }) =>
      HttpResponse.json({
        imei: params.imei as string,
        status: 'deregistered',
        deregisteredAt: now,
        retentionUntil: now + 30 * 24 * 60 * 60 * 1000,
      }),
    ),
  );
}

describe('useInbox', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when no organization is selected', () => {
    setOrg(null);
    const { result } = renderHookWithQueryClient(() => useInbox());
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches inbox via REST when organization is set', async () => {
    const { result } = renderHookWithQueryClient(() => useInbox({ status: 'pending' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.requests ?? []).toHaveLength(1);
    expect((result.current.data?.requests ?? [])[0]?.imei).toBe('123');
    expect((result.current.data?.requests ?? [])[0]?.status).toBe('pending');
    expect(result.current.data?.pagination?.total).toBe(1);
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/device/inbox');
    registerGraphQLResponse('GetInboxEntries', () => ({
      inbox: {
        __typename: 'InboxConnection',
        requests: [{ ...RAW_INBOX_ENTRY, __typename: 'InboxEntry' }],
        pagination: { total: 1, limit: 20, offset: 0, hasMore: false },
      },
    }));
    const { result } = renderHookWithQueryClient(() => useInbox());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.requests ?? []).toHaveLength(1);
    expect((result.current.data?.requests ?? [])[0]?.imei).toBe('123');
  });
});

describe('useInboxEntry', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useInboxEntry(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches entry via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useInboxEntry('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('pending');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/inbox/:imei');
    registerGraphQLResponse('GetInboxEntry', () => ({
      inboxEntry: { ...RAW_INBOX_ENTRY, __typename: 'InboxEntry' },
    }));
    const { result } = renderHookWithQueryClient(() => useInboxEntry('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('pending');
  });

  it('is disabled (idle) when no org is selected, even if REST would reject', () => {
    setOrg(null);
    makeRestFail('get', '/v1/inbox/:imei');
    const { result } = renderHookWithQueryClient(() => useInboxEntry('123'));
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useAcknowledgeInbox', () => {
  beforeEach(() => setOrg('org-1'));

  it('calls acknowledgeInbox via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useAcknowledgeInbox());
    result.current.mutate({ imei: '123', data: { action: 'approve' } });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('approved');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('post', '/v1/device/inbox/:imei/ack');
    registerGraphQLResponse('AckInbox', () => ({
      ackInbox: {
        __typename: 'AckResult',
        id: '1',
        imei: '123',
        status: 'rejected',
        acknowledgedAt: null,
        approvingAt: null,
        approvedAt: null,
        rejectedAt: now,
        commandSecret: null,
        fcmPushSent: false,
        notes: 'no',
      },
    }));
    const { result } = renderHookWithQueryClient(() => useAcknowledgeInbox());
    result.current.mutate({ imei: '123', data: { action: 'reject', notes: 'no' } });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('rejected');
    expect(result.current.data?.notes).toBe('no');
  });
});

describe('useCreateInboxRequest', () => {
  beforeEach(() => setOrg('org-1'));

  it('calls createInboxRequest with org', async () => {
    const { result } = renderHookWithQueryClient(() => useCreateInboxRequest());
    const request = { imei: '123', fcmToken: 't', firebaseInstallId: 'f' };
    result.current.mutate({ data: request });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('pending');
  });
});

describe('useConfirmDevice', () => {
  beforeEach(() => setOrg('org-1'));

  it('calls confirmDevice with imei, commandSecret and org', async () => {
    const { result } = renderHookWithQueryClient(() => useConfirmDevice());
    result.current.mutate({ imei: '123', commandSecret: 'secret' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.device_id).toBe('d1');
    expect(result.current.data?.confirmed).toBe(true);
    expect(result.current.data?.imei).toBe('123');
  });
});

describe('useDismissInbox', () => {
  beforeEach(() => setOrg('org-1'));

  it('calls dismissInbox with org', async () => {
    const { result } = renderHookWithQueryClient(() => useDismissInbox());
    result.current.mutate({ imei: '123', data: { status: 'rejected' } });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.status).toBe('rejected');
  });
});

describe('useResendInboxApproval', () => {
  beforeEach(() => setOrg('org-1'));

  it('calls resendInboxApproval with org', async () => {
    const { result } = renderHookWithQueryClient(() => useResendInboxApproval());
    result.current.mutate({ imei: '123' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.fcmPushSent).toBe(true);
    expect(result.current.data?.status).toBe('approved');
  });
});

describe('useRegisteredDevices', () => {
  beforeEach(() => {
    setOrg('org-1');
    overrideRegisteredDeviceHandlers();
  });

  it('fetches devices via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useRegisteredDevices({ status: 'online' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.devices).toHaveLength(1);
    expect(result.current.data?.devices[0]?.imei).toBe('123');
    expect(result.current.data?.devices[0]?.online).toBe(true);
    expect(result.current.data?.devices[0]?.status).toBe('online');
  });

  it('errors when REST rejects (no GraphQL fallback for devices)', async () => {
    makeRestFail('get', '/v1/devices');
    const { result } = renderHookWithQueryClient(() => useRegisteredDevices());
    await waitFor(() => expect(result.current.isError).toBe(true));
  });
});

describe('useRegisteredDevice', () => {
  beforeEach(() => {
    setOrg('org-1');
    overrideRegisteredDeviceHandlers();
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useRegisteredDevice(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches device via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useRegisteredDevice('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.online).toBe(true);
    expect(result.current.data?.status).toBe('online');
  });
});

describe('useDeregisterRegisteredDevice', () => {
  beforeEach(() => {
    setOrg('org-1');
    overrideRegisteredDeviceHandlers();
  });

  it('calls deregisterDevice via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useDeregisterRegisteredDevice());
    result.current.mutate({ imei: '123' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('deregistered');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('delete', '/v1/devices/:imei');
    registerGraphQLResponse('DeregisterDevice', () => ({
      deregisterDevice: {
        __typename: 'DeregisterResult',
        imei: '123',
        status: 'deregistered',
        deregisteredAt: now,
        retentionUntil: now + 30 * 24 * 60 * 60 * 1000,
      },
    }));
    const { result } = renderHookWithQueryClient(() => useDeregisterRegisteredDevice());
    result.current.mutate({ imei: '123', hard: true });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('deregistered');
  });
});

describe('useRegistrationStatus', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useRegistrationStatus(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('returns status from the REST inbox entry', async () => {
    const { result } = renderHookWithQueryClient(() => useRegistrationStatus('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('pending');
    expect(result.current.data?.entry?.imei).toBe('123');
  });

  it('returns null when no entry exists', async () => {
    server.use(http.get('/v1/device/inbox/:imei', () => HttpResponse.json(null)));
    const { result } = renderHookWithQueryClient(() => useRegistrationStatus('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBeNull();
  });
});

describe('useRegisterDevice', () => {
  beforeEach(() => setOrg('org-1'));

  it('calls createInboxRequest with org', async () => {
    const { result } = renderHookWithQueryClient(() => useRegisterDevice());
    const request = { imei: '123', fcmToken: 't', firebaseInstallId: 'f' };
    result.current.mutate({ data: request });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('123');
    expect(result.current.data?.status).toBe('pending');
  });
});
