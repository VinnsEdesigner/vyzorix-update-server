import { describe, it, expect, beforeEach, vi } from 'vitest';
import { waitFor } from '@testing-library/react';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { useAuthStore } from '@/stores/auth-store';

const { logsMock, queryLogsMock, authContextStub } = vi.hoisted(() => ({
  logsMock: {
    list: vi.fn(),
    get: vi.fn(),
    stats: vi.fn(),
  },
  queryLogsMock: vi.fn(),
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
  logs: logsMock,
  queryLogs: (...args: unknown[]) => queryLogsMock(...args),
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

const { useDeviceLogs, useLog, useLogStats } = await import('@/hooks/logs/use-logs');

describe('useDeviceLogs', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useDeviceLogs(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled when organizationId is null', () => {
    useAuthStore.setState({ organizationId: null });
    const { result } = renderHookWithQueryClient(() => useDeviceLogs('123'));
    expect(result.current.fetchStatus).toBe('idle');
    expect(logsMock.list).not.toHaveBeenCalled();
  });

  it('fetches logs via REST with organizationId', async () => {
    logsMock.list.mockResolvedValue({ logs: [], hasMore: false });
    const { result } = renderHookWithQueryClient(() => useDeviceLogs('123', { limit: 50 }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(logsMock.list).toHaveBeenCalledWith('123', expect.objectContaining({ limit: 50, organizationId: 'org-1' }));
    expect(queryLogsMock).not.toHaveBeenCalled();
  });

  it('falls back to GraphQL when REST rejects', async () => {
    logsMock.list.mockRejectedValue(new Error('REST down'));
    queryLogsMock.mockResolvedValue({
      deviceLogs: {
        events: [{ id: 'log-1', type: 'error', timestamp: 1704067200 }],
        pagination: { limit: 50, hasMore: false },
      },
    });
    const { result } = renderHookWithQueryClient(() => useDeviceLogs('123', { limit: 50 }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(queryLogsMock).toHaveBeenCalledWith(expect.objectContaining({ organizationId: 'org-1', imei: '123', limit: 50 }));
    expect(result.current.data?.logs).toHaveLength(1);
    expect(result.current.data?.logs[0]?.id).toBe('log-1');
    expect(result.current.data?.logs[0]?.eventType).toBe('error');
  });
});

describe('useLog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when id is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useLog(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches a single log with organizationId', async () => {
    logsMock.get.mockResolvedValue({ id: 'log-1', eventType: 'info' });
    const { result } = renderHookWithQueryClient(() => useLog('log-1'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(logsMock.get).toHaveBeenCalledWith('log-1', 'org-1');
  });
});

describe('useLogStats', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useLogStats(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches stats with organizationId', async () => {
    logsMock.stats.mockResolvedValue({ total: 0, byType: { connection: 0, command: 0, telemetry: 0, error: 0, warning: 0 } });
    const { result } = renderHookWithQueryClient(() => useLogStats('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(logsMock.stats).toHaveBeenCalledWith('123', expect.objectContaining({ organizationId: 'org-1' }));
  });
});
