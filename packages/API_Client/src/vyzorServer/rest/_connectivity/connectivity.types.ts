/**
 * Connectivity Monitor Types
 * 
 * Feature 4: Connectivity & Status Monitor Logic
 */

export interface QueuedRequest {
  id: string;
  method: string;
  url: string;
  data?: unknown;
  params?: Record<string, unknown>;
  headers?: Record<string, string>;
  timestamp: number;
  resolve: (value: unknown) => void;
  reject: (error: Error) => void;
}

export interface ConnectivityConfig {
  enableOfflineQueue: boolean;
  maxQueueSize: number;
  maxRequestAge: number;
  reconnectDelay: number;
  enableStaleDetection: boolean;
}

export interface NetworkState {
  isOnline: boolean;
  wasOnline: boolean;
  lastChecked: number;
  effectiveType?: string;
  downlink?: number;
  rtt?: number;
}

export type ConnectivityCallback = (state: NetworkState) => void;

export interface ConnectivityMonitor {
  getState(): NetworkState;
  isOnline(): boolean;
  subscribe(callback: ConnectivityCallback): () => void;
  queueRequest(request: Omit<QueuedRequest, 'id' | 'timestamp' | 'resolve' | 'reject'>): Promise<unknown>;
  getQueuedRequests(): QueuedRequest[];
  clearQueue(): void;
  flushQueue(): Promise<void>;
  getQueueSize(): number;
  checkConnectivity(): Promise<boolean>;
}

export const DEFAULT_CONNECTIVITY_CONFIG: ConnectivityConfig = {
  enableOfflineQueue: true,
  maxQueueSize: 100,
  maxRequestAge: 5 * 60 * 1000,
  reconnectDelay: 300,
  enableStaleDetection: true,
};
