import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import {
  useIsOnline,
  useOfflineQueueSize,
  useCheckConnectivity,
  useFlushOfflineQueue,
  useClearOfflineQueue,
  useConnectivity,
  useConnectivityState,
  useInitConnectivity,
} from '@/hooks/connectivity/use-connectivity';
import { useConnectivityStore } from '@/stores/connectivity-store';

vi.mock('@vyzorix/api-client', () => ({
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

describe('connectivity hooks', () => {
  beforeEach(() => {
    useConnectivityStore.setState({
      isOnline: true,
      wasOnline: true,
      lastChecked: 0,
      effectiveType: '4g',
      downlink: 10,
      rtt: 50,
      queueSize: 0,
      queuedRequests: [],
      isChecking: false,
      isFlushing: false,
      checkConnectivity: vi.fn(async () => true),
      flushQueue: vi.fn(async () => {}),
      clearQueue: vi.fn(),
    });
  });

  describe('useIsOnline', () => {
    it('returns isOnline from store', () => {
      const { result } = renderHook(() => useIsOnline());
      expect(result.current).toBe(true);
      act(() => useConnectivityStore.setState({ isOnline: false }));
      expect(result.current).toBe(false);
    });
  });

  describe('useOfflineQueueSize', () => {
    it('returns queueSize from store', () => {
      const { result } = renderHook(() => useOfflineQueueSize());
      expect(result.current).toBe(0);
      act(() => useConnectivityStore.setState({ queueSize: 5 }));
      expect(result.current).toBe(5);
    });
  });

  describe('useCheckConnectivity', () => {
    it('returns check fn and checking flag', () => {
      const { result } = renderHook(() => useCheckConnectivity());
      expect(typeof result.current.check).toBe('function');
      expect(result.current.checking).toBe(false);
    });

    it('reflects isChecking state', () => {
      const { result } = renderHook(() => useCheckConnectivity());
      act(() => useConnectivityStore.setState({ isChecking: true }));
      expect(result.current.checking).toBe(true);
    });
  });

  describe('useFlushOfflineQueue', () => {
    it('returns flush fn and flushing flag', () => {
      const { result } = renderHook(() => useFlushOfflineQueue());
      expect(typeof result.current.flush).toBe('function');
      expect(result.current.flushing).toBe(false);
    });
  });

  describe('useClearOfflineQueue', () => {
    it('returns clearQueue fn', () => {
      const { result } = renderHook(() => useClearOfflineQueue());
      expect(typeof result.current).toBe('function');
    });
  });

  describe('useConnectivityState', () => {
    it('returns network state slice', () => {
      const { result } = renderHook(() => useConnectivityState());
      expect(result.current).toEqual({
        isOnline: true,
        wasOnline: true,
        lastChecked: 0,
        effectiveType: '4g',
        downlink: 10,
        rtt: 50,
      });
    });
  });

  describe('useConnectivity', () => {
    it('returns aggregated state with all fields', () => {
      const { result } = renderHook(() => useConnectivity());
      expect(result.current.isOnline).toBe(true);
      expect(result.current.queueSize).toBe(0);
      expect(result.current.isChecking).toBe(false);
      expect(result.current.isFlushing).toBe(false);
      expect(typeof result.current.checkConnectivity).toBe('function');
      expect(typeof result.current.flushQueue).toBe('function');
      expect(typeof result.current.clearQueue).toBe('function');
    });
  });

  describe('useInitConnectivity', () => {
    it('returns a callback that calls checkConnectivity', () => {
      const { result } = renderHook(() => useInitConnectivity());
      expect(typeof result.current).toBe('function');
      act(() => result.current());
      expect(useConnectivityStore.getState().checkConnectivity).toHaveBeenCalled();
    });
  });
});
