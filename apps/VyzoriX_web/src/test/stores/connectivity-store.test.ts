import { describe, it, expect, beforeEach, vi } from 'vitest';

const subscribers = new Set<(s: unknown) => void>();
const mockState = {
  isOnline: true,
  wasOnline: true,
  lastChecked: Date.now(),
  effectiveType: '4g' as string,
  downlink: 10,
  rtt: 50,
};

const monitorMock = {
  subscribe: vi.fn((cb: (s: unknown) => void) => {
    subscribers.add(cb);
    return () => subscribers.delete(cb);
  }),
  getState: vi.fn(() => ({ ...mockState })),
  getQueueSize: vi.fn(() => 0),
  getQueuedRequests: vi.fn(() => []),
  checkConnectivity: vi.fn(async () => true),
  flushQueue: vi.fn(async () => {}),
  clearQueue: vi.fn(),
};

vi.mock('@vyzorix/api-client', () => ({
  initConnectivityMonitor: vi.fn(() => monitorMock),
  getConnectivityMonitor: vi.fn(() => monitorMock),
}));

const { useConnectivityStore } = await import('@/stores/connectivity-store');

function emit(state: Partial<typeof mockState>) {
  Object.assign(mockState, state);
  for (const cb of subscribers) cb({ ...mockState });
}

describe('useConnectivityStore', () => {
  beforeEach(() => {
    Object.assign(mockState, {
      isOnline: true,
      wasOnline: true,
      lastChecked: Date.now(),
      effectiveType: '4g',
      downlink: 10,
      rtt: 50,
    });
    vi.clearAllMocks();
    useConnectivityStore.setState({
      isOnline: true,
      wasOnline: true,
      lastChecked: mockState.lastChecked,
      effectiveType: '4g',
      downlink: 10,
      rtt: 50,
      queueSize: 0,
      queuedRequests: [],
      isChecking: false,
      isFlushing: false,
    });
  });

  it('initializes with monitor state', () => {
    const state = useConnectivityStore.getState();
    expect(state.isOnline).toBe(true);
    expect(state.queueSize).toBe(0);
    expect(state.isChecking).toBe(false);
  });

  it('syncs when monitor emits offline state', () => {
    emit({ isOnline: false, wasOnline: true });
    expect(useConnectivityStore.getState().isOnline).toBe(false);
  });

  it('checkConnectivity toggles isChecking and returns result', async () => {
    monitorMock.checkConnectivity.mockResolvedValueOnce(true);
    const promise = useConnectivityStore.getState().checkConnectivity();
    expect(useConnectivityStore.getState().isChecking).toBe(true);
    const result = await promise;
    expect(result).toBe(true);
    expect(useConnectivityStore.getState().isChecking).toBe(false);
  });

  it('flushQueue toggles isFlushing', async () => {
    monitorMock.flushQueue.mockResolvedValueOnce(undefined);
    const promise = useConnectivityStore.getState().flushQueue();
    expect(useConnectivityStore.getState().isFlushing).toBe(true);
    await promise;
    expect(useConnectivityStore.getState().isFlushing).toBe(false);
  });

  it('clearQueue delegates to monitor', () => {
    useConnectivityStore.getState().clearQueue();
    expect(monitorMock.clearQueue).toHaveBeenCalledOnce();
  });

  it('checkConnectivity resets isChecking even on error', async () => {
    monitorMock.checkConnectivity.mockRejectedValueOnce(new Error('fail'));
    await expect(useConnectivityStore.getState().checkConnectivity()).rejects.toThrow('fail');
    expect(useConnectivityStore.getState().isChecking).toBe(false);
  });
});
