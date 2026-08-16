/**
 * Integration tests for useWebSocketStore.
 *
 * The store uses the REAL graphql-ws `createClient`. A real WebSocket cannot
 * be opened in jsdom, so we inject a fake WebSocket *implementation* via the
 * graphql-ws `webSocketImpl` option — this is the transport layer, not the API
 * client. The store's `getCurrentOrganizationId()` runs against the REAL
 * authContext. State transitions are driven by the fake socket emitting the
 * events graphql-ws expects, exercising real store + real client logic.
 */
import { describe, it, expect, beforeEach, afterEach, vi } from 'vitest';
import { setupIntegrationTest } from '../helpers/integration-test-setup';
import { useWebSocketStore } from '@/stores/websocket-store';
import { useAuthStore } from '@/stores/auth-store';
import { authContext } from '@vyzorix/api-client';

setupIntegrationTest();

type Listener = (event: Event) => void;

/**
 * Minimal fake WebSocket speaking the graphql-transport-ws protocol.
 * gql-ws opens the socket, waits for `open`, then sends a ConnectionInit
 * message and expects a ConnectionAck message back before emitting
 * `connected`. Tests drive the socket via `receiveMessage`.
 */
class FakeWebSocket {
  static instances: FakeWebSocket[] = [];
  static OPEN = 1;
  static CLOSED = 3;

  readonly url: string;
  readonly protocol: string;
  readyState = FakeWebSocket.OPEN;
  onopen: ((ev: Event) => void) | null = null;
  onclose: ((ev: CloseEvent) => void) | null = null;
  onerror: ((ev: Event) => void) | null = null;
  onmessage: ((ev: MessageEvent) => void) | null = null;
  sent: string[] = [];
  private listeners: { type: string; fn: Listener }[] = [];

  constructor(url: string, protocols?: string | string[]) {
    this.url = url;
    this.protocol = Array.isArray(protocols) ? protocols[0] ?? '' : protocols ?? '';
    FakeWebSocket.instances.push(this);
    // gql-ws registers listeners synchronously after `new WebSocketImpl(...)`,
    // then checks readyState === OPEN. Defer the open event to the next tick.
    queueMicrotask(() => {
      this.readyState = FakeWebSocket.OPEN;
      this.dispatchEvent(new Event('open'));
    });
  }

  addEventListener(type: string, listener: Listener) {
    this.listeners.push({ type, fn: listener });
  }
  removeEventListener(type: string, listener: Listener) {
    this.listeners = this.listeners.filter((l) => !(l.type === type && l.fn === listener));
  }
  dispatchEvent(event: Event) {
    const type = event.type;
    const handler = (this as unknown as Record<string, ((e: Event) => void) | null>)[`on${type}`];
    handler?.(event);
    for (const l of this.listeners) {
      if (l.type === type) l.fn(event);
    }
    return true;
  }
  send(data: string) {
    this.sent.push(data);
    const msg = JSON.parse(data) as { type: string };
    // Respond to ConnectionInit with ConnectionAck to drive the `connected` event.
    if (msg.type === 'connection_init') {
      this.receiveMessage({ type: 'connection_ack' });
    }
  }
  close(code = 1000, reason = '') {
    this.readyState = FakeWebSocket.CLOSED;
    const event = new CloseEvent('close', { code, reason, wasClean: true });
    this.onclose?.(event);
    this.dispatchEvent(event);
  }
  /** Feed an incoming graphql-transport-ws message to the client. */
  receiveMessage(message: unknown) {
    this.onmessage?.({ data: JSON.stringify(message) } as MessageEvent);
    this.dispatchEvent(new MessageEvent('message', { data: JSON.stringify(message) }));
  }
}

beforeEach(() => {
  FakeWebSocket.instances = [];
  vi.stubGlobal('WebSocket', FakeWebSocket);
  // Reset the module-level wsClient singleton between tests.
  useWebSocketStore.getState().disconnect();
});

afterEach(() => {
  useWebSocketStore.getState().disconnect();
  vi.unstubAllGlobals();
});

describe('useWebSocketStore', () => {
  beforeEach(() => {
    useAuthStore.getState().setOrganization('org-1');
  });

  it('starts disconnected', () => {
    const state = useWebSocketStore.getState();
    expect(state.connectionState).toBe('CLOSED');
    expect(state.isConnected).toBe(false);
    expect(state.activeSubscriptions).toBe(0);
  });

  it('connect opens a graphql-ws client over a real WebSocket', async () => {
    useWebSocketStore.getState().connect();
    await waitForState(() => useWebSocketStore.getState().isConnected);
    expect(useWebSocketStore.getState().connectionState).toBe('OPEN');
    expect(useWebSocketStore.getState().isConnected).toBe(true);
    expect(useWebSocketStore.getState().lastConnectedAt).toBeInstanceOf(Date);
  });

  it('connect is idempotent (does not create a second client)', async () => {
    useWebSocketStore.getState().connect();
    await waitForState(() => useWebSocketStore.getState().isConnected);
    useWebSocketStore.getState().connect();
    expect(FakeWebSocket.instances).toHaveLength(1);
  });

  it('disconnect disposes the client and resets state', async () => {
    useWebSocketStore.getState().connect();
    await waitForState(() => useWebSocketStore.getState().isConnected);
    useWebSocketStore.getState().disconnect();
    expect(useWebSocketStore.getState().connectionState).toBe('CLOSED');
    expect(useWebSocketStore.getState().isConnected).toBe(false);
    expect(useWebSocketStore.getState().activeSubscriptions).toBe(0);
  });

  it('disconnect is a no-op when not connected', () => {
    useWebSocketStore.getState().disconnect();
    expect(useWebSocketStore.getState().connectionState).toBe('CLOSED');
  });

  it('subscribe increments activeSubscriptions and delivers data', async () => {
    useWebSocketStore.getState().connect();
    await waitForState(() => useWebSocketStore.getState().isConnected);
    const next = vi.fn();
    const unsub = useWebSocketStore.getState().subscribe(
      { query: 'subscription { s }' },
      { next },
    );
    expect(useWebSocketStore.getState().activeSubscriptions).toBe(1);
    const socket = FakeWebSocket.instances[0]!;
    const subMsg = await waitForSubscribeMessage(socket);
    socket.receiveMessage({ id: subMsg.id, type: 'next', payload: { data: { s: 1 } } });
    expect(next).toHaveBeenCalledWith({ s: 1 });
    unsub();
  });

  it('subscribe auto-connects if no client exists', async () => {
    expect(useWebSocketStore.getState().connectionState).toBe('CLOSED');
    useWebSocketStore.getState().subscribe({ query: 'subscription { s }' }, {});
    await waitForState(() => useWebSocketStore.getState().isConnected);
    expect(FakeWebSocket.instances).toHaveLength(1);
    expect(useWebSocketStore.getState().activeSubscriptions).toBe(1);
  });

  it('connect does not open a socket when no organization is selected', async () => {
    authContext.clear();
    // gql-ws's non-lazy loop turns the url() throw into an unhandled rejection;
    // capture and swallow it so it doesn't fail the test run.
    const captured: unknown[] = [];
    const handler = (reason: unknown) => captured.push(reason);
    process.on('unhandledRejection', handler);
    useWebSocketStore.getState().connect();
    // buildWebSocketUrl throws inside gql-ws's async url resolver, before any
    // WebSocket is constructed. Give the microtask queue a chance to run.
    await new Promise((r) => setTimeout(r, 50));
    process.off('unhandledRejection', handler);
    expect(captured.some((e) => String(e).includes('organization'))).toBe(true);
    expect(FakeWebSocket.instances).toHaveLength(0);
    expect(useWebSocketStore.getState().isConnected).toBe(false);
  });

  it('completing a subscription decrements activeSubscriptions', async () => {
    useWebSocketStore.getState().connect();
    await waitForState(() => useWebSocketStore.getState().isConnected);
    const unsub = useWebSocketStore.getState().subscribe(
      { query: 'subscription { s }' },
      {},
    );
    expect(useWebSocketStore.getState().activeSubscriptions).toBe(1);
    const socket = FakeWebSocket.instances[0]!;
    const subMsg = await waitForSubscribeMessage(socket);
    socket.receiveMessage({ id: subMsg.id, type: 'complete' });
    // gql-ws invokes sink.complete() asynchronously after receiving `complete`.
    await waitForState(() => useWebSocketStore.getState().activeSubscriptions === 0);
    unsub();
  });
});

/** Wait for gql-ws to send the `subscribe` message (the 2nd frame after connection_init). */
function waitForSubscribeMessage(
  socket: FakeWebSocket,
  timeout = 1000,
): Promise<{ id: string }> {
  return new Promise((resolve, reject) => {
    const start = Date.now();
    const check = () => {
      const sub = socket.sent
        .map((raw) => JSON.parse(raw) as { type: string; id?: string })
        .find((m) => m.type === 'subscribe' && m.id);
      if (sub?.id) return resolve({ id: sub.id });
      if (Date.now() - start > timeout) {
        return reject(new Error(`Timed out waiting for subscribe message; sent=${JSON.stringify(socket.sent)}`));
      }
      setTimeout(check, 5);
    };
    check();
  });
}

function waitForState(predicate: () => boolean, timeout = 1000): Promise<void> {
  return new Promise((resolve, reject) => {
    const start = Date.now();
    const check = () => {
      if (predicate()) return resolve();
      if (Date.now() - start > timeout) {
        return reject(new Error('Timed out waiting for state'));
      }
      setTimeout(check, 5);
    };
    check();
  });
}
