import { createVyzorStore } from '@/lib/state';
import { createClient, type Client, type Sink, type SubscribePayload } from 'graphql-ws';
import { getCurrentOrganizationId } from '@vyzorix/api-client';

export type ConnectionState = 'CLOSED' | 'CONNECTING' | 'OPEN' | 'CLOSING' | 'RECONNECTING';

export interface ConnectionError {
  code: number;
  reason: string;
}

export interface SubscriptionHandlers<T = unknown> {
  next?: (data: T) => void;
  error?: (error: unknown) => void;
  complete?: () => void;
}

export interface WebSocketStoreState {
  connectionState: ConnectionState;
  isConnected: boolean;
  isReconnecting: boolean;
  lastConnectedAt: Date | null;
  lastError: ConnectionError | null;
  reconnectAttempts: number;
  activeSubscriptions: number;
  connect: () => void;
  disconnect: () => void;
  subscribe: <T = unknown>(
    payload: SubscribePayload,
    handlers: SubscriptionHandlers<T>,
  ) => () => void;
}

function buildWebSocketUrl(): string {
  const baseUrl =
    (import.meta.env as Record<string, string | undefined>).VITE_API_URL || '/api';
  const orgId = getCurrentOrganizationId();
  if (!orgId) {
    throw new Error('Cannot connect WebSocket: no organization selected');
  }
  const wsBase = baseUrl
    .replace(/^http:\/\//, 'ws://')
    .replace(/^https:\/\//, 'wss://');
  return `${wsBase}/${orgId}/graphql/ws`;
}

function toConnectionError(value: unknown): ConnectionError {
  if (value && typeof value === 'object') {
    const code = (value as { code?: number }).code;
    const reason = (value as { reason?: string }).reason;
    return {
      code: typeof code === 'number' ? code : 0,
      reason: typeof reason === 'string' ? reason : String(value),
    };
  }
  return { code: 0, reason: String(value) };
}

function toErrorMessage(value: unknown): string {
  if (value instanceof Error) return value.message;
  return String(value);
}

let wsClient: Client | null = null;

export const useWebSocketStore = createVyzorStore<WebSocketStoreState>('WebSocketStore', (set, get) => ({
  connectionState: 'CLOSED',
  isConnected: false,
  isReconnecting: false,
  lastConnectedAt: null,
  lastError: null,
  reconnectAttempts: 0,
  activeSubscriptions: 0,

  connect: () => {
    if (wsClient) return;

    const client = createClient({
      url: buildWebSocketUrl,
      lazy: false,
      retryAttempts: Infinity,
      keepAlive: 10_000,
      connectionAckWaitTimeout: 5_000,
      onNonLazyError: (error) => {
        set({ lastError: toConnectionError(error) });
      },
      on: {
        connecting: (isRetry) => {
          set((state) => ({
            connectionState: isRetry ? 'RECONNECTING' : 'CONNECTING',
            isReconnecting: isRetry,
            reconnectAttempts: isRetry ? state.reconnectAttempts + 1 : 0,
          }));
        },
        connected: () => {
          set({
            connectionState: 'OPEN',
            isConnected: true,
            isReconnecting: false,
            lastConnectedAt: new Date(),
            lastError: null,
            reconnectAttempts: 0,
          });
        },
        closed: (event) => {
          set((state) => ({
            connectionState: 'CLOSED',
            isConnected: false,
            isReconnecting: false,
            lastError: event ? toConnectionError(event) : state.lastError,
          }));
        },
        error: (error) => {
          set({ lastError: { code: 0, reason: toErrorMessage(error) } });
        },
      },
    });

    wsClient = client;
  },

  disconnect: () => {
    if (!wsClient) return;
    set({ connectionState: 'CLOSING' });
    wsClient.dispose();
    wsClient = null;
    set({
      connectionState: 'CLOSED',
      isConnected: false,
      isReconnecting: false,
      activeSubscriptions: 0,
    });
  },

  subscribe: <T = unknown>(payload: SubscribePayload, handlers: SubscriptionHandlers<T>) => {
    if (!wsClient) {
      get().connect();
    }
    const client = wsClient!;
    const sink: Sink = {
      next: (value) => {
        const data = (value as { data?: unknown }).data;
        if (data !== undefined) handlers.next?.(data as T);
      },
      error: (error) => handlers.error?.(error),
      complete: () => {
        set((state) => ({ activeSubscriptions: Math.max(0, state.activeSubscriptions - 1) }));
        handlers.complete?.();
      },
    };
    set((state) => ({ activeSubscriptions: state.activeSubscriptions + 1 }));
    return client.subscribe(payload, sink);
  },
}));
