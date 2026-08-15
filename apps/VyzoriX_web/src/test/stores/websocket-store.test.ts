import { describe, it, expect, beforeEach, vi } from 'vitest';

let wsClientDispose = vi.fn();
let subscribeFn = vi.fn(() => vi.fn());

const createClientMock = vi.fn(() => ({
  dispose: wsClientDispose,
  subscribe: subscribeFn,
}));

vi.mock('graphql-ws', () => ({
  createClient: (...args: unknown[]) => createClientMock(...(args as [])),
}));

vi.mock('@vyzorix/api-client', () => ({
  getCurrentOrganizationId: vi.fn(() => 'org-1'),
}));

const { useWebSocketStore } = await import('@/stores/websocket-store');

describe('useWebSocketStore', () => {
  beforeEach(() => {
    useWebSocketStore.getState().disconnect();
    wsClientDispose = vi.fn();
    subscribeFn = vi.fn(() => vi.fn());
    createClientMock.mockReturnValue({
      dispose: wsClientDispose,
      subscribe: subscribeFn,
    });
    vi.clearAllMocks();
    useWebSocketStore.setState({
      connectionState: 'CLOSED',
      isConnected: false,
      isReconnecting: false,
      lastConnectedAt: null,
      lastError: null,
      reconnectAttempts: 0,
      activeSubscriptions: 0,
    });
  });

  it('starts disconnected', () => {
    const state = useWebSocketStore.getState();
    expect(state.connectionState).toBe('CLOSED');
    expect(state.isConnected).toBe(false);
    expect(state.activeSubscriptions).toBe(0);
  });

  it('connect creates a graphql-ws client', () => {
    useWebSocketStore.getState().connect();
    expect(createClientMock).toHaveBeenCalledOnce();
  });

  it('connect is idempotent (does not create a second client)', () => {
    useWebSocketStore.getState().connect();
    useWebSocketStore.getState().connect();
    expect(createClientMock).toHaveBeenCalledOnce();
  });

  it('disconnect disposes the client and resets state', () => {
    useWebSocketStore.getState().connect();
    useWebSocketStore.getState().disconnect();
    expect(wsClientDispose).toHaveBeenCalledOnce();
    const state = useWebSocketStore.getState();
    expect(state.connectionState).toBe('CLOSED');
    expect(state.isConnected).toBe(false);
  });

  it('disconnect is a no-op when not connected', () => {
    useWebSocketStore.getState().disconnect();
    expect(wsClientDispose).not.toHaveBeenCalled();
  });

  it('subscribe increments activeSubscriptions and returns unsubscribe', () => {
    useWebSocketStore.getState().connect();
    const unsub = useWebSocketStore.getState().subscribe(
      { query: 'subscription { s }' },
      { next: vi.fn() },
    );
    expect(useWebSocketStore.getState().activeSubscriptions).toBe(1);
    expect(subscribeFn).toHaveBeenCalledOnce();
    unsub();
  });

  it('subscribe auto-connects if no client exists', () => {
    expect(createClientMock).not.toHaveBeenCalled();
    useWebSocketStore.getState().subscribe(
      { query: 'subscription { s }' },
      {},
    );
    expect(createClientMock).toHaveBeenCalledOnce();
  });
});
