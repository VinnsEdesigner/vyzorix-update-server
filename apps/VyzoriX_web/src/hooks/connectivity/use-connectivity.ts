import { useCallback } from 'react';
import { useShallow } from 'zustand/react/shallow';
import { useConnectivityStore } from '@/stores/connectivity-store';
import type { NetworkState } from '@vyzorix/api-client';

export type { NetworkState };

export interface UseConnectivityReturn extends NetworkState {
  queueSize: number;
  queuedRequests: ReturnType<typeof useConnectivityStore.getState>['queuedRequests'];
  checkConnectivity: () => Promise<boolean>;
  flushQueue: () => Promise<void>;
  clearQueue: () => void;
  isChecking: boolean;
  isFlushing: boolean;
}

export function useConnectivityState(): NetworkState {
  return useConnectivityStore(
    useShallow((s) => ({
      isOnline: s.isOnline,
      wasOnline: s.wasOnline,
      lastChecked: s.lastChecked,
      effectiveType: s.effectiveType,
      downlink: s.downlink,
      rtt: s.rtt,
    })),
  );
}

export function useIsOnline(): boolean {
  return useConnectivityStore((s) => s.isOnline);
}

export function useOfflineQueueSize(): number {
  return useConnectivityStore((s) => s.queueSize);
}

export function useOfflineQueue() {
  return useConnectivityStore((s) => s.queuedRequests);
}

export function useCheckConnectivity() {
  const checkConnectivity = useConnectivityStore((s) => s.checkConnectivity);
  const isChecking = useConnectivityStore((s) => s.isChecking);
  return { check: checkConnectivity, checking: isChecking };
}

export function useFlushOfflineQueue() {
  const flushQueue = useConnectivityStore((s) => s.flushQueue);
  const isFlushing = useConnectivityStore((s) => s.isFlushing);
  return { flush: flushQueue, flushing: isFlushing };
}

export function useClearOfflineQueue() {
  return useConnectivityStore((s) => s.clearQueue);
}

export function useConnectivity(): UseConnectivityReturn {
  const state = useConnectivityState();
  const queueSize = useOfflineQueueSize();
  const queuedRequests = useOfflineQueue();
  const checkConnectivity = useConnectivityStore((s) => s.checkConnectivity);
  const flushQueue = useConnectivityStore((s) => s.flushQueue);
  const clearQueue = useConnectivityStore((s) => s.clearQueue);
  const isChecking = useConnectivityStore((s) => s.isChecking);
  const isFlushing = useConnectivityStore((s) => s.isFlushing);

  return {
    ...state,
    queueSize,
    queuedRequests,
    checkConnectivity,
    flushQueue,
    clearQueue,
    isChecking,
    isFlushing,
  };
}

export function useInitConnectivity() {
  const checkConnectivity = useConnectivityStore((s) => s.checkConnectivity);
  return useCallback(() => {
    void checkConnectivity();
  }, [checkConnectivity]);
}
