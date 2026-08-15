import { describe, it, expect, beforeEach } from 'vitest';
import { useMetricsRealtimeStore } from '@/stores/metrics-realtime-store';
import type { TelemetryFrame } from '@vyzorix/api-client';

function makeFrame(overrides: Partial<TelemetryFrame> = {}): TelemetryFrame {
  return {
    timestamp: new Date(),
    riskScore: 10,
    thermalTemp: 40,
    bufferLevel: 50,
    ...overrides,
  };
}

describe('useMetricsRealtimeStore', () => {
  beforeEach(() => {
    useMetricsRealtimeStore.setState({ byDevice: {}, windowMs: 6 * 60 * 60 * 1000, lastFrame: {}, activeOrganizationId: null });
  });

  it('push adds a point and sets lastFrame', () => {
    const frame = makeFrame({ riskScore: 25 });
    useMetricsRealtimeStore.getState().push('123', frame);
    expect(useMetricsRealtimeStore.getState().byDevice['123']).toHaveLength(1);
    expect(useMetricsRealtimeStore.getState().getLastFrame('123')?.riskScore).toBe(25);
  });

  it('getSeries returns the requested metric', () => {
    useMetricsRealtimeStore.getState().push('123', makeFrame({ thermalTemp: 60 }));
    const series = useMetricsRealtimeStore.getState().getSeries('123', 'thermalTemp');
    expect(series).toHaveLength(1);
    expect(series[0]?.value).toBe(60);
  });

  it('sliding window evicts old points', () => {
    const old = new Date(Date.now() - 10 * 60 * 60 * 1000);
    useMetricsRealtimeStore.getState().push('123', makeFrame({ timestamp: old, riskScore: 1 }));
    useMetricsRealtimeStore.getState().push('123', makeFrame({ riskScore: 2 }));
    const entries = useMetricsRealtimeStore.getState().byDevice['123']!;
    expect(entries).toHaveLength(1);
    expect(entries[0]?.riskScore).toBe(2);
  });

  it('setWindow changes the eviction window', () => {
    useMetricsRealtimeStore.getState().setWindow(1000);
    const recent = new Date(Date.now() - 500);
    useMetricsRealtimeStore.getState().push('123', makeFrame({ timestamp: recent, riskScore: 5 }));
    expect(useMetricsRealtimeStore.getState().byDevice['123']).toHaveLength(1);
  });

  it('clear removes a single device', () => {
    useMetricsRealtimeStore.getState().push('123', makeFrame());
    useMetricsRealtimeStore.getState().push('456', makeFrame());
    useMetricsRealtimeStore.getState().clear('123');
    expect(useMetricsRealtimeStore.getState().byDevice['123']).toBeUndefined();
    expect(useMetricsRealtimeStore.getState().byDevice['456']).toHaveLength(1);
  });

  it('setActiveOrganization clears on org switch', () => {
    useMetricsRealtimeStore.getState().setActiveOrganization('org-1');
    useMetricsRealtimeStore.getState().push('123', makeFrame());
    useMetricsRealtimeStore.getState().setActiveOrganization('org-2');
    expect(Object.keys(useMetricsRealtimeStore.getState().byDevice)).toHaveLength(0);
  });
});
