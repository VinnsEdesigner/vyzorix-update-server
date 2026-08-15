import { describe, it, expect, beforeEach, vi } from 'vitest';
import { waitFor } from '@testing-library/react';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { useAuthStore } from '@/stores/auth-store';
import { useCommandDispatchStore } from '@/stores';

const { commandsMock, pollCommandStatusMock, queryPendingCommandsMock, queryCommandMock, authContextStub } = vi.hoisted(() => ({
  commandsMock: {
    getHistory: vi.fn(),
    getByDispatchId: vi.fn(),
    getPending: vi.fn(),
    send: vi.fn(),
    cancel: vi.fn(),
    retry: vi.fn(),
  },
  pollCommandStatusMock: vi.fn(),
  queryPendingCommandsMock: vi.fn(),
  queryCommandMock: vi.fn(),
  authContextStub: {
    getState: vi.fn(() => ({
      isAuthenticated: false,
      operator: null,
      organizationId: null,
      accessToken: null,
      refreshToken: null,
      tokenExpiresAt: null,
    })),
    getLockoutState: vi.fn(() => ({ isLocked: false, retryAfter: 0, lockedUntil: 0 })),
    onChange: vi.fn(() => () => {}),
    setOrganization: vi.fn(),
    setAccessToken: vi.fn(),
    setFromLoginWithTokens: vi.fn(),
    setFromMeResponse: vi.fn(),
    refreshTokens: vi.fn(async () => {}),
    setLockout: vi.fn(),
    clear: vi.fn(),
  },
}));

vi.mock('@vyzorix/api-client', () => ({
  commands: commandsMock,
  pollCommandStatus: (...args: unknown[]) => pollCommandStatusMock(...args),
  queryPendingCommands: (...args: unknown[]) => queryPendingCommandsMock(...args),
  queryCommand: (...args: unknown[]) => queryCommandMock(...args),
  authContext: authContextStub,
  getCurrentOrganizationId: vi.fn(() => null),
  initConnectivityMonitor: vi.fn(() => ({
    subscribe: vi.fn(() => () => {}),
    getState: vi.fn(() => ({
      isOnline: true,
      wasOnline: true,
      lastChecked: 0,
      effectiveType: '4g',
      downlink: 10,
      rtt: 50,
    })),
    getQueueSize: vi.fn(() => 0),
    getQueuedRequests: vi.fn(() => []),
    checkConnectivity: vi.fn(async () => true),
    flushQueue: vi.fn(async () => {}),
    clearQueue: vi.fn(),
  })),
  getConnectivityMonitor: vi.fn(() => ({
    getQueueSize: vi.fn(() => 0),
    getQueuedRequests: vi.fn(() => []),
  })),
}));

const { useCommandHistory, useCommand, usePendingCommands, useSendCommand, useCancelCommand, useRetryCommand } =
  await import('@/hooks/commands/use-commands');

describe('useCommandHistory', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
    useCommandDispatchStore.setState({ pendingCommands: {}, pendingCount: 0 });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useCommandHistory(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches history when imei is provided', async () => {
    commandsMock.getHistory.mockResolvedValue({ items: [], total: 0 });
    const { result } = renderHookWithQueryClient(() => useCommandHistory('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(commandsMock.getHistory).toHaveBeenCalled();
  });

  it('passes params including organizationId', async () => {
    commandsMock.getHistory.mockResolvedValue({ items: [], total: 0 });
    const { result } = renderHookWithQueryClient(() =>
      useCommandHistory('123', { status: 'pending', page: 1, limit: 10 }),
    );
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(commandsMock.getHistory).toHaveBeenCalledWith(
      '123',
      expect.objectContaining({ status: 'pending', page: 1, limit: 10, organizationId: 'org-1' }),
    );
  });

  it('is disabled when organizationId is null', () => {
    useAuthStore.setState({ organizationId: null });
    const { result } = renderHookWithQueryClient(() => useCommandHistory('123'));
    expect(result.current.fetchStatus).toBe('idle');
    expect(commandsMock.getHistory).not.toHaveBeenCalled();
  });
});

describe('useCommand', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when dispatchId is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useCommand(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches command by dispatchId', async () => {
    commandsMock.getByDispatchId.mockResolvedValue({ dispatchId: 'd1', status: 'completed' });
    const { result } = renderHookWithQueryClient(() => useCommand('d1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(commandsMock.getByDispatchId).toHaveBeenCalledWith('d1', 'org-1');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    commandsMock.getByDispatchId.mockRejectedValue(new Error('REST down'));
    queryCommandMock.mockResolvedValue({
      command: { dispatchId: 'd1', commandId: 'c1', deviceId: '123', command: 'FORCE_SPEAKER', status: 'pending', createdAt: '2024-01-01T00:00:00Z' },
    });
    const { result } = renderHookWithQueryClient(() => useCommand('d1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryCommandMock).toHaveBeenCalledWith({ organizationId: 'org-1', dispatchId: 'd1' });
    expect(result.current.data?.dispatchId).toBe('d1');
    expect(result.current.data?.status).toBe('pending');
  });
});

describe('usePendingCommands', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => usePendingCommands(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches pending commands', async () => {
    commandsMock.getPending.mockResolvedValue([]);
    const { result } = renderHookWithQueryClient(() => usePendingCommands('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(commandsMock.getPending).toHaveBeenCalledWith('123', 'org-1');
  });

  it('falls back to GraphQL when REST rejects', async () => {
    commandsMock.getPending.mockRejectedValue(new Error('REST down'));
    queryPendingCommandsMock.mockResolvedValue({
      pendingCommands: [
        { dispatchId: 'd1', commandId: 'c1', deviceId: '123', command: 'FORCE_SPEAKER', status: 'pending', createdAt: '2024-01-01T00:00:00Z' },
      ],
    });
    const { result } = renderHookWithQueryClient(() => usePendingCommands('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryPendingCommandsMock).toHaveBeenCalledWith({ organizationId: 'org-1', deviceId: '123' });
    expect(result.current.data).toHaveLength(1);
    expect(result.current.data?.[0]?.dispatchId).toBe('d1');
  });
});

describe('useSendCommand', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
    useCommandDispatchStore.setState({ pendingCommands: {}, pendingCount: 0 });
  });

  it('calls commands.send and adds to pending store', async () => {
    const sentCommand = { dispatchId: 'disp-1', deviceId: '123', status: 'queued' };
    commandsMock.send.mockResolvedValue(sentCommand);
    const { result } = renderHookWithQueryClient(() => useSendCommand());
    result.current.mutate({ deviceId: '123', commandType: 'FORCE_SPEAKER' } as never);
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(commandsMock.send).toHaveBeenCalledWith(
      { deviceId: '123', commandType: 'FORCE_SPEAKER' },
      'org-1',
    );
    expect(useCommandDispatchStore.getState().getPending('disp-1')).toBeDefined();
    expect(useCommandDispatchStore.getState().pendingCount).toBe(1);
  });
});

describe('useCancelCommand', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
    useCommandDispatchStore.setState({ pendingCommands: {}, pendingCount: 0 });
  });

  it('calls cancel and removes from pending', async () => {
    useCommandDispatchStore.getState().addPending({
      dispatchId: 'disp-1',
      imei: '123',
      commandType: 'FORCE_SPEAKER',
      createdAt: Date.now(),
    });
    commandsMock.cancel.mockResolvedValue({});
    const { result } = renderHookWithQueryClient(() => useCancelCommand());
    result.current.mutate('disp-1');
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(commandsMock.cancel).toHaveBeenCalledWith('disp-1', 'org-1');
    expect(useCommandDispatchStore.getState().getPending('disp-1')).toBeUndefined();
  });
});

describe('useRetryCommand', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls retry with dispatchId and org', async () => {
    commandsMock.retry.mockResolvedValue({ dispatchId: 'disp-2', deviceId: '123', status: 'queued' });
    const { result } = renderHookWithQueryClient(() => useRetryCommand());
    result.current.mutate('disp-1');
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(commandsMock.retry).toHaveBeenCalledWith('disp-1', 'org-1');
  });
});
