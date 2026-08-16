import type {
  ConnectivityConfig,
  ConnectivityMonitor,
  ConnectivityCallback,
  NetworkState,
  QueuedRequest,
} from './connectivity.types';

import { DEFAULT_CONNECTIVITY_CONFIG } from './connectivity.types';

const STORAGE_KEY = 'vyzorix_offline_queue';

interface SerializedRequest {
  id: string;
  method: string;
  url: string;
  data?: unknown;
  params?: Record<string, unknown>;
  headers?: Record<string, string>;
  timestamp: number;
}

function generateRequestId(): string {
  return `req_${Date.now()}_${Math.random().toString(36).substring(2, 9)}`;
}

function saveQueueToStorage(queue: Map<string, QueuedRequest>): void {
  if (typeof localStorage === 'undefined') return;
  const serialized: SerializedRequest[] = [];
  queue.forEach((req) => {
    serialized.push({
      id: req.id,
      method: req.method,
      url: req.url,
      data: req.data,
      params: req.params,
      headers: req.headers,
      timestamp: req.timestamp,
    });
  });
  localStorage.setItem(STORAGE_KEY, JSON.stringify(serialized));
}

function loadQueueFromStorage(): SerializedRequest[] {
  if (typeof localStorage === 'undefined') return [];
  try {
    const stored = localStorage.getItem(STORAGE_KEY);
    if (!stored) return [];
    return JSON.parse(stored);
  } catch {
    return [];
  }
}

function clearStorage(): void {
  if (typeof localStorage === 'undefined') return;
  localStorage.removeItem(STORAGE_KEY);
}

function getBrowserOnlineStatus(): boolean {
  return typeof navigator !== 'undefined' ? navigator.onLine : true;
}

function isReactNative(): boolean {
  return typeof navigator !== 'undefined' && (navigator as { product?: string }).product === 'ReactNative';
}

function setupWebListeners(onChange: (online: boolean) => void): () => void {
  if (typeof window === 'undefined') return () => {};
  const handleOnline = (): void => onChange(true);
  const handleOffline = (): void => onChange(false);
  window.addEventListener('online', handleOnline);
  window.addEventListener('offline', handleOffline);
  return () => {
    window.removeEventListener('online', handleOnline);
    window.removeEventListener('offline', handleOffline);
  };
}

function setupReactNativeListeners(onChange: (online: boolean) => void): () => void {
  if (!isReactNative()) return () => {};
  try {
    const req = eval('require');
    const NetInfo = req('@react-native-community/netinfo').default;
    const unsubscribe = NetInfo.addEventListener((state: { isConnected: boolean | null }) => {
      onChange(state.isConnected ?? true);
    });
    return () => {
      if (typeof unsubscribe === 'function') unsubscribe();
    };
  } catch {
    return () => {};
  }
}

async function checkWebConnectivity(): Promise<boolean> {
  if (typeof window === 'undefined') return true;
  try {
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), 5000);
    const response = await fetch(`${window.location.origin}/health`, {
      method: 'HEAD',
      cache: 'no-cache',
      signal: controller.signal,
    });
    clearTimeout(timeoutId);
    return response.ok;
  } catch {
    return navigator.onLine;
  }
}

async function checkReactNativeConnectivity(): Promise<boolean> {
  if (!isReactNative()) return checkWebConnectivity();
  try {
    const req = eval('require');
    const NetInfo = req('@react-native-community/netinfo').default;
    const state = await NetInfo.fetch();
    return state.isConnected ?? true;
  } catch {
    return true;
  }
}

export function createConnectivityMonitor(
  config: Partial<ConnectivityConfig> = {}
): ConnectivityMonitor {
  const fullConfig: ConnectivityConfig = {
    ...DEFAULT_CONNECTIVITY_CONFIG,
    ...config,
  };

  let isOnline = getBrowserOnlineStatus();
  let state: NetworkState = {
    isOnline,
    wasOnline: isOnline,
    lastChecked: Date.now(),
  };

  const subscribers = new Set<ConnectivityCallback>();
  const requestQueue = new Map<string, QueuedRequest>();
  const pendingPromises = new Map<string, { resolve: (v: unknown) => void; reject: (e: Error) => void }>();
  let isFlushing = false;

  function notifySubscribers(): void {
    subscribers.forEach((callback) => {
      try {
        callback(state);
      } catch (error) {
        console.error('[Connectivity] Subscriber error:', error);
      }
    });
  }

  function updateState(updates: Partial<NetworkState>): void {
    const prevState = { ...state };
    state = { ...state, ...updates, lastChecked: Date.now() };

    if (prevState.isOnline !== state.isOnline) {
      console.log(`[Connectivity] ${prevState.isOnline ? 'ONLINE' : 'OFFLINE'} → ${state.isOnline ? 'ONLINE' : 'OFFLINE'}`);
      notifySubscribers();

      if (state.isOnline && !prevState.isOnline) {
        flushQueueInternal();
      }
    }
  }

  function handleNetworkChange(online: boolean): void {
    isOnline = online;
    updateState({ isOnline, wasOnline: online });
  }

  async function executeRequest(request: QueuedRequest): Promise<unknown> {
    console.log(`[Connectivity] Executing ${request.method} ${request.url}`);
    const axios = (await import('axios')).default;
    const response = await axios({
      method: request.method,
      url: request.url,
      data: request.data,
      params: request.params,
      headers: request.headers,
    });
    return response.data;
  }

  async function flushQueueInternal(): Promise<void> {
    if (isFlushing || !isOnline || requestQueue.size === 0) return;

    isFlushing = true;
    const now = Date.now();
    const requests = Array.from(requestQueue.values()).sort((a, b) => a.timestamp - b.timestamp);

    console.log(`[Connectivity] Flushing ${requests.length} queued requests`);

    for (const request of requests) {
      if (!isOnline) break;

      const age = now - request.timestamp;
      if (fullConfig.enableStaleDetection && age > fullConfig.maxRequestAge) {
        const pending = pendingPromises.get(request.id);
        if (pending) {
          pending.reject(new Error(`Request stale (${Math.round(age / 1000)}s old)`));
          pendingPromises.delete(request.id);
        }
        requestQueue.delete(request.id);
        continue;
      }

      try {
        const result = await executeRequest(request);
        const pending = pendingPromises.get(request.id);
        if (pending) {
          pending.resolve(result);
          pendingPromises.delete(request.id);
        }
        requestQueue.delete(request.id);
      } catch (error) {
        const pending = pendingPromises.get(request.id);
        if (pending) {
          pending.reject(error as Error);
          pendingPromises.delete(request.id);
        }
        requestQueue.delete(request.id);
      }
    }

    saveQueueToStorage(requestQueue);
    isFlushing = false;
  }

  setupWebListeners(handleNetworkChange);
  setupReactNativeListeners(handleNetworkChange);

  const stored = loadQueueFromStorage();
  if (stored.length > 0) {
    console.log(`[Connectivity] Restored ${stored.length} queued requests from storage`);
    stored.forEach((sr: SerializedRequest) => {
      requestQueue.set(sr.id, { ...sr, resolve: () => {}, reject: () => {} });
    });
    if (isOnline) flushQueueInternal();
  }

  const monitor: ConnectivityMonitor = {
    getState(): NetworkState {
      return { ...state };
    },

    isOnline(): boolean {
      return isOnline;
    },

    subscribe(callback: ConnectivityCallback): () => void {
      subscribers.add(callback);
      callback(state);
      return () => subscribers.delete(callback);
    },

    queueRequest(request: Omit<QueuedRequest, 'id' | 'timestamp' | 'resolve' | 'reject'>): Promise<unknown> {
      if (!fullConfig.enableOfflineQueue) {
        return Promise.reject(new Error('Offline queueing is disabled'));
      }

      if (requestQueue.size >= fullConfig.maxQueueSize) {
        return Promise.reject(new Error(`Queue full (max: ${fullConfig.maxQueueSize})`));
      }

      const id = generateRequestId();

      return new Promise<unknown>((resolve, reject) => {
        pendingPromises.set(id, { resolve, reject });

        const queuedRequest: QueuedRequest = {
          ...request,
          id,
          timestamp: Date.now(),
          resolve,
          reject,
        };

        requestQueue.set(id, queuedRequest);
        saveQueueToStorage(requestQueue);

        console.log(`[Connectivity] Queued ${request.method} ${request.url} (${requestQueue.size} total)`);

        if (isOnline) flushQueueInternal();
      });
    },

    getQueuedRequests(): QueuedRequest[] {
      return Array.from(requestQueue.values()).sort((a, b) => a.timestamp - b.timestamp);
    },

    clearQueue(): void {
      const count = requestQueue.size;
      pendingPromises.forEach((p) => p.reject(new Error('Queue cleared')));
      pendingPromises.clear();
      requestQueue.clear();
      clearStorage();
      console.log(`[Connectivity] Cleared ${count} queued requests`);
    },

    async flushQueue(): Promise<void> {
      return flushQueueInternal();
    },

    getQueueSize(): number {
      return requestQueue.size;
    },

    async checkConnectivity(): Promise<boolean> {
      const connected = isReactNative()
        ? await checkReactNativeConnectivity()
        : await checkWebConnectivity();
      updateState({ isOnline: connected });
      return connected;
    },
  };

  return monitor;
}

let defaultMonitor: ConnectivityMonitor | null = null;

export function getConnectivityMonitor(): ConnectivityMonitor {
  if (!defaultMonitor) {
    defaultMonitor = createConnectivityMonitor();
  }
  return defaultMonitor;
}

export function initConnectivityMonitor(config?: Partial<ConnectivityConfig>): ConnectivityMonitor {
  if (defaultMonitor) return defaultMonitor;
  defaultMonitor = createConnectivityMonitor(config);
  return defaultMonitor;
}
