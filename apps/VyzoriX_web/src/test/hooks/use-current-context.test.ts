import { describe, it, expect, beforeEach } from 'vitest';
import { renderHook } from '@testing-library/react';
import {
  useCurrentOrganizationId,
  useRequiredOrganizationId,
  useSelectedImei,
  useRequiredImei,
} from '@/hooks/_shared/use-current-context';
import { useAuthStore } from '@/stores/auth-store';
import { useDeviceSelectorStore } from '@/stores/device-selector-store';

describe('useCurrentOrganizationId', () => {
  beforeEach(() => {
    useAuthStore.setState({ organizationId: null });
  });

  it('returns null when no org selected', () => {
    const { result } = renderHook(() => useCurrentOrganizationId());
    expect(result.current).toBeNull();
  });

  it('returns the org id when set', () => {
    useAuthStore.setState({ organizationId: 'org-1' });
    const { result } = renderHook(() => useCurrentOrganizationId());
    expect(result.current).toBe('org-1');
  });
});

describe('useRequiredOrganizationId', () => {
  beforeEach(() => {
    useAuthStore.setState({ organizationId: null });
  });

  it('throws when no org selected', () => {
    const { result } = renderHook(() => {
      try {
        useRequiredOrganizationId();
        return null;
      } catch (e) {
        return e as Error;
      }
    });
    expect(result.current?.message).toContain('No organization selected');
  });

  it('returns org id when set', () => {
    useAuthStore.setState({ organizationId: 'org-1' });
    const { result } = renderHook(() => {
      try {
        return useRequiredOrganizationId();
      } catch (e) {
        return e as Error;
      }
    });
    expect(result.current).toBe('org-1');
  });
});

describe('useSelectedImei', () => {
  beforeEach(() => {
    useDeviceSelectorStore.setState({ selectedDevice: null });
  });

  it('returns null when no device selected', () => {
    const { result } = renderHook(() => useSelectedImei());
    expect(result.current).toBeNull();
  });

  it('returns imei when device selected', () => {
    useDeviceSelectorStore.setState({
      selectedDevice: { id: 'dev-1', imei: '123456789012345' },
    });
    const { result } = renderHook(() => useSelectedImei());
    expect(result.current).toBe('123456789012345');
  });

  it('returns null when selected device has no imei', () => {
    useDeviceSelectorStore.setState({ selectedDevice: { id: 'dev-1' } });
    const { result } = renderHook(() => useSelectedImei());
    expect(result.current).toBeNull();
  });
});

describe('useRequiredImei', () => {
  beforeEach(() => {
    useDeviceSelectorStore.setState({ selectedDevice: null });
  });

  it('throws when no device selected', () => {
    const { result } = renderHook(() => {
      try {
        useRequiredImei();
        return null;
      } catch (e) {
        return e as Error;
      }
    });
    expect(result.current?.message).toContain('No device selected');
  });

  it('throws when device has no imei', () => {
    useDeviceSelectorStore.setState({ selectedDevice: { id: 'dev-1' } });
    const { result } = renderHook(() => {
      try {
        useRequiredImei();
        return null;
      } catch (e) {
        return e as Error;
      }
    });
    expect(result.current?.message).toContain('No device selected');
  });

  it('returns imei when set', () => {
    useDeviceSelectorStore.setState({
      selectedDevice: { id: 'dev-1', imei: '123456789012345' },
    });
    const { result } = renderHook(() => {
      try {
        return useRequiredImei();
      } catch (e) {
        return e as Error;
      }
    });
    expect(result.current).toBe('123456789012345');
  });
});
