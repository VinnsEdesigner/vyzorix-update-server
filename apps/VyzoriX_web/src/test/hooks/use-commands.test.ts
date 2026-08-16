/**
 * Integration tests for commands hooks.
 *
 * These tests render the REAL hooks via React Testing Library. The hooks call
 * the REAL API client functions (commands.getHistory, commands.send, etc.)
 * which use the REAL restClient (axios) + domain mappers. MSW intercepts the
 * HTTP requests and returns mock server responses.
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
import { useCommandDispatchStore } from '@/stores';
import { graphqlClient } from '@vyzorix/api-client';
import {
  useCommandHistory,
  useCommand,
  usePendingCommands,
  useSendCommand,
  useCancelCommand,
  useRetryCommand,
} from '@/hooks/commands/use-commands';

const { server } = setupIntegrationTest();

const IMEI = '123';
const now = Date.now();

function setOrg(orgId: string | null) {
  useAuthStore.getState().setOrganization(orgId);
  if (orgId) {
    graphqlClient.setOrganization(orgId);
  }
}

function resetDispatchStore() {
  useCommandDispatchStore.setState({ pendingCommands: {}, pendingCount: 0 });
}

/** Override a REST handler to return 500, triggering GraphQL fallback. */
function makeRestFail(method: 'get' | 'post' | 'delete', path: string) {
  server.use(http[method](path, () => HttpResponse.json({ error: 'REST down' }, { status: 500 })));
}

// --- GraphQL response fixtures ---

function makeGraphQLCommand(overrides: Record<string, unknown> = {}) {
  return {
    __typename: 'Command',
    dispatchId: 'd1',
    commandId: 'c1',
    deviceId: IMEI,
    command: 'FORCE_SPEAKER',
    args: {},
    status: 'pending',
    createdAt: '2024-01-01T00:00:00Z',
    deliveredAt: null,
    ...overrides,
  };
}

beforeEach(() => {
  setOrg(null);
  resetDispatchStore();
});

describe('useCommandHistory', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useCommandHistory(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches history when imei is provided', async () => {
    const { result } = renderHookWithQueryClient(() => useCommandHistory(IMEI));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.commands).toHaveLength(2);
    expect(result.current.data?.commands[0]?.dispatchId).toBe('disp-1');
  });

  it('passes params including organizationId', async () => {
    const { result } = renderHookWithQueryClient(() =>
      useCommandHistory(IMEI, { status: 'pending', page: 1, limit: 10 }),
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.pagination.page).toBe(1);
    expect(result.current.data?.pagination.limit).toBe(10);
  });

  it('is disabled when organizationId is null', () => {
    setOrg(null);
    const { result } = renderHookWithQueryClient(() => useCommandHistory(IMEI));
    expect(result.current.fetchStatus).toBe('idle');
  });
});

describe('useCommand', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when dispatchId is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useCommand(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches command by dispatchId', async () => {
    const { result } = renderHookWithQueryClient(() => useCommand('disp-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.dispatchId).toBe('disp-1');
    expect(result.current.data?.status).toBe('completed');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/command/:dispatchId/status');
    registerGraphQLResponse('GetCommand', () => ({
      command: makeGraphQLCommand({ dispatchId: 'd1', status: 'pending' }),
    }));
    const { result } = renderHookWithQueryClient(() => useCommand('d1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.dispatchId).toBe('d1');
    expect(result.current.data?.status).toBe('pending');
  });
});

describe('usePendingCommands', () => {
  beforeEach(() => setOrg('org-1'));

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => usePendingCommands(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches pending commands', async () => {
    const { result } = renderHookWithQueryClient(() => usePendingCommands(IMEI));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]?.dispatchId).toBe('disp-1');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    makeRestFail('get', '/v1/device/:imei/commands/pending');
    registerGraphQLResponse('GetPendingCommands', () => ({
      pendingCommands: [makeGraphQLCommand({ dispatchId: 'd1', status: 'pending' })],
    }));
    const { result } = renderHookWithQueryClient(() => usePendingCommands(IMEI));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]?.dispatchId).toBe('d1');
  });
});

describe('useSendCommand', () => {
  beforeEach(() => {
    setOrg('org-1');
    resetDispatchStore();
  });

  it('sends command via REST and adds to pending store', async () => {
    const { result } = renderHookWithQueryClient(() => useSendCommand());
    result.current.mutate({ imei: IMEI, commandType: 'FORCE_SPEAKER' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.dispatchId).toBe('disp-1');
    expect(result.current.data?.deviceId).toBe(IMEI);
    expect(useCommandDispatchStore.getState().getPending('disp-1')).toBeDefined();
    expect(useCommandDispatchStore.getState().pendingCount).toBe(1);
  });
});

describe('useCancelCommand', () => {
  beforeEach(() => {
    setOrg('org-1');
    resetDispatchStore();
  });

  it('cancels command via REST and removes from pending store', async () => {
    useCommandDispatchStore.getState().addPending({
      dispatchId: 'disp-1',
      imei: IMEI,
      commandType: 'FORCE_SPEAKER',
      createdAt: now,
    });
    expect(useCommandDispatchStore.getState().pendingCount).toBe(1);

    const { result } = renderHookWithQueryClient(() => useCancelCommand());
    result.current.mutate('disp-1');
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.success).toBe(true);
    expect(useCommandDispatchStore.getState().getPending('disp-1')).toBeUndefined();
    expect(useCommandDispatchStore.getState().pendingCount).toBe(0);
  });
});

describe('useRetryCommand', () => {
  beforeEach(() => setOrg('org-1'));

  it('retries command via REST', async () => {
    const { result } = renderHookWithQueryClient(() => useRetryCommand());
    result.current.mutate('disp-1');
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.dispatchId).toBe('disp-1');
    expect(result.current.data?.status).toBe('queued');
  });
});
