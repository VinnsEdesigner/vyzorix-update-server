/**
 * Connectivity Hook
 * 
 * Feature 4: Connectivity & Status Monitor Logic
 * 
 * React hooks for connectivity monitoring.
 * These wrap the connectivity monitor from @vyzorix/api-client.
 */

import { useState, useEffect, useCallback } from 'react';
import { 
  initConnectivityMonitor, 
  getConnectivityMonitor, 
  type ConnectivityMonitor, 
  type NetworkState 
} from '@vyzorix/api-client';

export type { NetworkState };

/**
 * Initialize connectivity monitoring at app startup
 * Call this once in your app root
 */
export function useInitConnectivity() {
  useEffect(() => {
    initConnectivityMonitor();
  }, []);
}

/**
 * Hook to get current connectivity state
 */
export function useConnectivityState(): NetworkState {
  const monitor = getConnectivityMonitor();
  const [state, setState] = useState<NetworkState>(() => monitor.getState());

  useEffect(() => {
    return monitor.subscribe(setState);
  }, [monitor]);

  return state;
}

/**
 * Hook to check if currently online
 */
export function useIsOnline(): boolean {
  const state = useConnectivityState();
  return state.isOnline;
}

/**
 * Hook to get queued request count
 */
export function useOfflineQueueSize(): number {
  const monitor = getConnectivityMonitor();
  const [size, setSize] = useState(() => monitor.getQueueSize());

  useEffect(() => {
    return monitor.subscribe(() => {
      setSize(monitor.getQueueSize());
    });
  }, [monitor]);

  return size;
}

/**
 * Hook to get queued requests
 */
export function useOfflineQueue() {
  const monitor = getConnectivityMonitor();
  const [requests, setRequests] = useState(() => monitor.getQueuedRequests());

  useEffect(() => {
    return monitor.subscribe(() => {
      setRequests(monitor.getQueuedRequests());
    });
  }, [monitor]);

  return requests;
}

/**
 * Hook to manually trigger connectivity check
 */
export function useCheckConnectivity() {
  const monitor = getConnectivityMonitor();
  const [checking, setChecking] = useState(false);

  const check = useCallback(async () => {
    setChecking(true);
    try {
      return await monitor.checkConnectivity();
    } finally {
      setChecking(false);
    }
  }, [monitor]);

  return { check, checking };
}

/**
 * Hook to flush offline queue manually
 */
export function useFlushOfflineQueue() {
  const monitor = getConnectivityMonitor();
  const [flushing, setFlushing] = useState(false);

  const flush = useCallback(async () => {
    if (flushing) return;
    setFlushing(true);
    try {
      await monitor.flushQueue();
    } finally {
      setFlushing(false);
    }
  }, [monitor, flushing]);

  return { flush, flushing };
}

/**
 * Hook to clear offline queue
 */
export function useClearOfflineQueue() {
  const monitor = getConnectivityMonitor();

  const clear = useCallback(() => {
    monitor.clearQueue();
  }, [monitor]);

  return clear;
}

/**
 * Combined connectivity hook with all states and actions
 */
export interface UseConnectivityReturn extends NetworkState {
  queueSize: number;
  queuedRequests: ReturnType<ConnectivityMonitor['getQueuedRequests']>;
  checkConnectivity: () => Promise<boolean>;
  flushQueue: () => Promise<void>;
  clearQueue: () => void;
  checkingConnectivity: boolean;
  flushingQueue: boolean;
}

export function useConnectivity(): UseConnectivityReturn {
  const state = useConnectivityState();
  const queueSize = useOfflineQueueSize();
  const queuedRequests = useOfflineQueue();
  const { check, checking } = useCheckConnectivity();
  const { flush, flushing } = useFlushOfflineQueue();
  const clear = useClearOfflineQueue();

  return {
    ...state,
    queueSize,
    queuedRequests,
    checkConnectivity: check,
    flushQueue: flush,
    clearQueue: clear,
    checkingConnectivity: checking,
    flushingQueue: flushing,
  };
}
