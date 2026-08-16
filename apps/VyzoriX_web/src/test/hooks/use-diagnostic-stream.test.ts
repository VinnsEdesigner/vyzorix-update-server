import { describe, it, expect, beforeEach, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useWebSocketStore } from '@/stores/websocket-store';
import { useDiagnosticsStore } from '@/stores/diagnostics-store';
import { useTimelineStreamStore } from '@/stores/timeline-stream-store';
import { useAuthStore } from '@/stores/auth-store';

/**
 * Real hook + real stores. The only thing stubbed is the WebSocket transport:
 * we inject a fake `subscribe` into `useWebSocketStore` state so no real socket
 * is opened. This exercises the real hook wiring (store patching, subscriptions,
 * cleanup) without mocking the API client or the stores module.
 */
let subscribeMock: ReturnType<typeof vi.fn>;

const { useDiagnosticStream } = await import('@/hooks/diagnostics/use-diagnostic-stream');

describe('useDiagnosticStream', () => {
  beforeEach(() => {
    subscribeMock = vi.fn();
    subscribeMock.mockImplementation(() => vi.fn());
    useWebSocketStore.setState({ subscribe: subscribeMock as never });
    useDiagnosticsStore.setState({
      snapshots: {},
      lastRefreshedAt: {},
      isRefreshing: {},
      refreshIntervalMs: 10_000,
      isPolling: false,
      activeOrganizationId: null,
    });
    useTimelineStreamStore.setState({
      byDevice: {},
      filters: {},
      autoScroll: true,
      activeOrganizationId: null,
    });
    useAuthStore.setState({ organizationId: 'org-1' });
  });

  it('does not subscribe when imei is undefined', () => {
    renderHook(() => useDiagnosticStream(undefined));
    expect(subscribeMock).not.toHaveBeenCalled();
  });

  it('does not subscribe when organizationId is null', () => {
    useAuthStore.setState({ organizationId: null });
    renderHook(() => useDiagnosticStream('123'));
    expect(subscribeMock).not.toHaveBeenCalled();
  });

  it('sets the active org on both stores', () => {
    renderHook(() => useDiagnosticStream('123'));
    expect(useDiagnosticsStore.getState().activeOrganizationId).toBe('org-1');
    expect(useTimelineStreamStore.getState().activeOrganizationId).toBe('org-1');
  });

  it('subscribes to device + telemetry and patches stores on next', () => {
    const handlersList: { next: (d: unknown) => void }[] = [];
    subscribeMock.mockImplementation((_payload, handlers) => {
      handlersList.push(handlers);
      return vi.fn();
    });

    renderHook(() => useDiagnosticStream('123'));
    // Two subscriptions: device-updated + telemetry-received.
    expect(subscribeMock).toHaveBeenCalledTimes(2);
    expect(handlersList).toHaveLength(2);

    // Seed a snapshot so patchConnection has something to merge into.
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', {
      identity: { imei: '123' },
      software: {},
      registration: { status: 'online', fcmTokenValid: true, commandSecretSet: true },
      connection: { webSocketStatus: 'connected', fcmStatus: 'valid' },
      telemetry: { framesToday: 0, sessionsToday: 0 },
    });

    // device-updated handler (first subscription).
    act(() => {
      handlersList[0]!.next({ id: '1', imei: '123', status: 'disconnected', lastSeen: '2024-06-20T12:00:00Z' });
    });
    expect(useDiagnosticsStore.getState().getSnapshot('org-1', '123')?.connection.webSocketStatus).toBe('disconnected');

    // telemetry-received handler (second subscription).
    act(() => {
      handlersList[1]!.next({
        id: 'evt-1',
        deviceId: '123',
        receivedAt: '2024-06-20T12:01:00Z',
        riskScore: 5,
        bufferLevel: 42,
        thermalTemp: 38,
      });
    });
    const snap = useDiagnosticsStore.getState().getSnapshot('org-1', '123')!;
    expect(snap.telemetry.avgLatencyMs).toBe(42);
    expect(snap.telemetry.lastTimestamp).toBeInstanceOf(Date);

    const events = useTimelineStreamStore.getState().getEvents('123');
    expect(events).toHaveLength(1);
    expect(events[0]?.id).toBe('evt-1');
    expect(events[0]?.type).toBe('TELEMETRY');
    expect(events[0]?.deviceId).toBe('123');
  });

  it('unsubscribes and clears the stream on unmount', () => {
    const unsub1 = vi.fn();
    const unsub2 = vi.fn();
    let call = 0;
    subscribeMock.mockImplementation(() => {
      call += 1;
      return call === 1 ? unsub1 : unsub2;
    });

    const { unmount } = renderHook(() => useDiagnosticStream('123'));
    unmount();
    expect(unsub1).toHaveBeenCalledTimes(1);
    expect(unsub2).toHaveBeenCalledTimes(1);
  });
});
