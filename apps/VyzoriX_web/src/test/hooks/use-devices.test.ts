import { describe, it, expect, beforeEach, vi } from 'vitest';
import { waitFor } from '@testing-library/react';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { useAuthStore } from '@/stores/auth-store';

const { devicesMock, authContextStub } = vi.hoisted(() => ({
  devicesMock: {
    list: vi.fn(),
    get: vi.fn(),
    getConnectionStatus: vi.fn(),
    getSettings: vi.fn(),
    updateSettings: vi.fn(),
    count: vi.fn(),
    stats: vi.fn(),
    deregister: vi.fn(),
    disconnect: vi.fn(),
  },
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
  devices: devicesMock,
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

const { useDevices, useDevice, useDeviceCount, useUpdateDeviceSettings, useDeregisterDevice } =
  await import('@/hooks/device/use-devices');

describe('useDevices', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: null });
  });

  it('is disabled when no organization is selected', () => {
    const { result } = renderHookWithQueryClient(() => useDevices());
    expect(result.current.isLoading).toBe(false);
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches devices when organization is set', async () => {
    useAuthStore.setState({ organizationId: 'org-1' });
    devicesMock.list.mockResolvedValue({ items: [], total: 0 });
    const { result } = renderHookWithQueryClient(() => useDevices());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(devicesMock.list).toHaveBeenCalledWith({ organizationId: 'org-1' });
  });

  it('passes params to the API', async () => {
    useAuthStore.setState({ organizationId: 'org-1' });
    devicesMock.list.mockResolvedValue({ items: [], total: 0 });
    const { result } = renderHookWithQueryClient(() => useDevices({ status: 'online' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(devicesMock.list).toHaveBeenCalledWith({ status: 'online', organizationId: 'org-1' });
  });
});

describe('useDevice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('is disabled when imei is undefined', () => {
    const { result } = renderHookWithQueryClient(() => useDevice(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('is disabled when imei is empty string', () => {
    const { result } = renderHookWithQueryClient(() => useDevice(''));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches device when imei is provided', async () => {
    devicesMock.get.mockResolvedValue({ imei: '123', name: 'Dev' });
    const { result } = renderHookWithQueryClient(() => useDevice('123'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(devicesMock.get).toHaveBeenCalledWith('123', 'org-1');
  });
});

describe('useDeviceCount', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: null });
  });

  it('is disabled without org', () => {
    const { result } = renderHookWithQueryClient(() => useDeviceCount());
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches count with org', async () => {
    useAuthStore.setState({ organizationId: 'org-1' });
    devicesMock.count.mockResolvedValue(42);
    const { result } = renderHookWithQueryClient(() => useDeviceCount());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data).toBe(42);
  });
});

describe('useUpdateDeviceSettings', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls updateSettings and returns the result', async () => {
    const updated = { imei: '123', fcmEnabled: true };
    devicesMock.updateSettings.mockResolvedValue(updated);
    const { result } = renderHookWithQueryClient(() => useUpdateDeviceSettings('123'));
    result.current.mutate({ fcmEnabled: true });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(devicesMock.updateSettings).toHaveBeenCalledWith('123', { fcmEnabled: true }, 'org-1');
    expect(result.current.data).toEqual(updated);
  });
});

describe('useDeregisterDevice', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('calls deregister with imei and org', async () => {
    devicesMock.deregister.mockResolvedValue({});
    const { result } = renderHookWithQueryClient(() => useDeregisterDevice());
    result.current.mutate('123');
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(devicesMock.deregister).toHaveBeenCalledWith('123', 'org-1');
  });
});
