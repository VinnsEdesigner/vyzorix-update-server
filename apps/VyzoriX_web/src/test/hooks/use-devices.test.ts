/**
 * Integration tests for useDevices / useDevice / useDeviceCount / etc.
 *
 * These tests render the REAL hooks via React Testing Library. The hooks call
 * the REAL API client functions (devices.list, devices.get, etc.) which use
 * the REAL restClient (axios) + domain mappers. MSW intercepts the HTTP
 * requests and returns mock server responses.
 *
 * No vi.mock — the real code path runs end-to-end.
 */
import { describe, it, expect, beforeEach } from 'vitest';
import { waitFor } from '@testing-library/react';
import { renderHookWithQueryClient } from '../helpers/render-hook';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
import { useAuthStore } from '@/stores/auth-store';
import {
  useDevices,
  useDevice,
  useDeviceCount,
  useUpdateDeviceSettings,
  useDeviceThresholds,
  useUpdateDeviceThresholds,
  useDeregisterDevice,
} from '@/hooks/device/use-devices';

setupIntegrationTest();

function setOrg(orgId: string | null) {
  // Use authContext.setOrganization so both authContext and the store stay in sync.
  // Using useAuthStore.setState directly would be overwritten by the next onChange event.
  useAuthStore.getState().setOrganization(orgId);
}

beforeEach(() => {
  setOrg(null);
});

describe('useDevices', () => {
  it('is disabled when no organization is selected', () => {
    const { result } = renderHookWithQueryClient(() => useDevices());
    expect(result.current.isLoading).toBe(false);
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches devices when organization is set', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useDevices());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.devices ?? []).toHaveLength(3);
    expect(result.current.data?.total).toBe(3);
  });

  it('passes status param to the API', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useDevices({ search: 'online' }));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    // MSW returns all 3 devices regardless of filter, but the hook should succeed
    expect(result.current.data?.devices).toBeDefined();
  });

  it('maps device list items through the real mapper', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useDevices());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    const first = (result.current.data?.devices ?? [])[0];
    expect(first).toBeDefined();
    expect(first?.imei).toBeDefined();
    expect(['online', 'offline']).toContain(first?.status);
  });
});

describe('useDevice', () => {
  beforeEach(() => {
    setOrg('org-1');
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
    const { result } = renderHookWithQueryClient(() => useDevice('111111111111111'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.imei).toBe('111111111111111');
  });

  it('returns error for unknown imei', async () => {
    const { result } = renderHookWithQueryClient(() => useDevice('unknown-imei'));
    await waitFor(() => expect(result.current.isError).toBe(true));
    expect(result.current.data).toBeUndefined();
  });
});

describe('useDeviceCount', () => {
  it('is disabled without org', () => {
    const { result } = renderHookWithQueryClient(() => useDeviceCount());
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches count with org', async () => {
    setOrg('org-1');
    const { result } = renderHookWithQueryClient(() => useDeviceCount());
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.count).toBe(3);
  });
});

describe('useUpdateDeviceSettings', () => {
  beforeEach(() => {
    setOrg('org-1');
  });

  it('calls updateSettings and returns the result', async () => {
    const { result } = renderHookWithQueryClient(() => useUpdateDeviceSettings());
    result.current.mutate({ imei: '111111111111111', data: { customName: 'My Device' } });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.customName).toBe('My Device');
    expect(result.current.data?.deviceImei).toBe('111111111111111');
  });
});

describe('useDeviceThresholds', () => {
  beforeEach(() => {
    setOrg('org-1');
  });

  it('is disabled without imei', () => {
    const { result } = renderHookWithQueryClient(() => useDeviceThresholds(undefined));
    expect(result.current.fetchStatus).toBe('idle');
  });

  it('fetches thresholds with org context', async () => {
    const { result } = renderHookWithQueryClient(() => useDeviceThresholds('111111111111111'));
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.thresholds).toBeDefined();
    expect(result.current.data?.thresholds?.riskWarn).toBe(70);
  });
});

describe('useUpdateDeviceThresholds', () => {
  beforeEach(() => {
    setOrg('org-1');
  });

  it('calls updateSettingsThresholds and caches result', async () => {
    const { result } = renderHookWithQueryClient(() => useUpdateDeviceThresholds());
    result.current.mutate({ imei: '111111111111111', data: { riskWarn: 75 } });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.thresholds?.riskWarn).toBe(75);
  });
});

describe('useDeregisterDevice', () => {
  beforeEach(() => {
    setOrg('org-1');
  });

  it('calls deregister with imei and org', async () => {
    const { result } = renderHookWithQueryClient(() => useDeregisterDevice());
    result.current.mutate({ imei: '111111111111111' });
    await waitFor(() => expect(result.current.isSuccess).toBe(true));
    expect(result.current.data?.success).toBe(true);
  });
});
