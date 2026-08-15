import { describe, it, expect, beforeEach } from 'vitest';
import { useDiagnosticsStore } from '@/stores/diagnostics-store';
import type { DeviceInspection } from '@vyzorix/api-client';

function makeInspection(overrides: Partial<DeviceInspection> = {}): DeviceInspection {
  return {
    identity: { imei: '861234567890123' },
    software: {},
    registration: { status: 'online', fcmTokenValid: true, commandSecretSet: true },
    connection: {
      webSocketStatus: 'connected',
      connectedAt: new Date(),
      fcmStatus: 'valid',
      lastSeen: new Date(),
      clientIp: '10.0.0.1',
      protocol: 'WSS',
    },
    telemetry: { framesToday: 10, sessionsToday: 1 },
    ...overrides,
  };
}

describe('useDiagnosticsStore', () => {
  beforeEach(() => {
    useDiagnosticsStore.setState({
      snapshots: {},
      lastRefreshedAt: {},
      isRefreshing: {},
      refreshIntervalMs: 10_000,
      isPolling: false,
      activeOrganizationId: null,
    });
  });

  it('setSnapshot stores under orgId:imei key and stamps lastRefreshedAt', () => {
    const inspection = makeInspection();
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', inspection);
    expect(useDiagnosticsStore.getState().getSnapshot('org-1', '123')).toEqual(inspection);
    expect(useDiagnosticsStore.getState().lastRefreshedAt['org-1:123']).not.toBeNull();
    expect(useDiagnosticsStore.getState().isRefreshing['org-1:123']).toBe(false);
  });

  it('getSnapshot returns undefined for unknown key', () => {
    expect(useDiagnosticsStore.getState().getSnapshot('org-1', '999')).toBeUndefined();
  });

  it('org isolation: snapshots are keyed by orgId, not just imei', () => {
    const a = makeInspection({ identity: { imei: '123' } });
    const b = makeInspection({ identity: { imei: '123' }, software: { osVersion: '14' } });
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', a);
    useDiagnosticsStore.getState().setSnapshot('org-2', '123', b);
    expect(useDiagnosticsStore.getState().getSnapshot('org-1', '123')).toBe(a);
    expect(useDiagnosticsStore.getState().getSnapshot('org-2', '123')).toBe(b);
    expect(useDiagnosticsStore.getState().getSnapshot('org-2', '123')?.software.osVersion).toBe('14');
  });

  it('patchConnection merges into the stored snapshot', () => {
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', makeInspection());
    useDiagnosticsStore.getState().patchConnection('org-1', '123', {
      webSocketStatus: 'disconnected',
      clientIp: '192.168.1.1',
    });
    const snap = useDiagnosticsStore.getState().getSnapshot('org-1', '123')!;
    expect(snap.connection.webSocketStatus).toBe('disconnected');
    expect(snap.connection.clientIp).toBe('192.168.1.1');
    // Unpatched fields preserved.
    expect(snap.connection.protocol).toBe('WSS');
  });

  it('patchConnection is a no-op when no snapshot exists', () => {
    useDiagnosticsStore.getState().patchConnection('org-1', '123', { clientIp: 'x' });
    expect(useDiagnosticsStore.getState().snapshots['org-1:123']).toBeUndefined();
  });

  it('patchTelemetry merges into the stored snapshot', () => {
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', makeInspection());
    useDiagnosticsStore.getState().patchTelemetry('org-1', '123', {
      avgLatencyMs: 42,
      framesToday: 99,
    });
    const snap = useDiagnosticsStore.getState().getSnapshot('org-1', '123')!;
    expect(snap.telemetry.avgLatencyMs).toBe(42);
    expect(snap.telemetry.framesToday).toBe(99);
  });

  it('setRefreshing toggles the per-key flag', () => {
    useDiagnosticsStore.getState().setRefreshing('org-1', '123', true);
    expect(useDiagnosticsStore.getState().isRefreshing['org-1:123']).toBe(true);
    useDiagnosticsStore.getState().setRefreshing('org-1', '123', false);
    expect(useDiagnosticsStore.getState().isRefreshing['org-1:123']).toBe(false);
  });

  it('setRefreshInterval updates the cadence', () => {
    useDiagnosticsStore.getState().setRefreshInterval(30_000);
    expect(useDiagnosticsStore.getState().refreshIntervalMs).toBe(30_000);
  });

  it('setActiveOrganization clears all snapshots on org switch (no cross-org leak)', () => {
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', makeInspection());
    useDiagnosticsStore.getState().setSnapshot('org-1', '456', makeInspection());
    useDiagnosticsStore.getState().setActiveOrganization('org-2');
    expect(useDiagnosticsStore.getState().snapshots).toEqual({});
    expect(useDiagnosticsStore.getState().lastRefreshedAt).toEqual({});
    expect(useDiagnosticsStore.getState().activeOrganizationId).toBe('org-2');
  });

  it('setActiveOrganization is a no-op when org is unchanged', () => {
    useDiagnosticsStore.getState().setActiveOrganization('org-1');
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', makeInspection());
    useDiagnosticsStore.getState().setActiveOrganization('org-1');
    expect(useDiagnosticsStore.getState().getSnapshot('org-1', '123')).toBeDefined();
  });

  it('clear removes a single orgId:imei entry', () => {
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', makeInspection());
    useDiagnosticsStore.getState().setSnapshot('org-1', '456', makeInspection());
    useDiagnosticsStore.getState().clear('org-1', '123');
    expect(useDiagnosticsStore.getState().getSnapshot('org-1', '123')).toBeUndefined();
    expect(useDiagnosticsStore.getState().getSnapshot('org-1', '456')).toBeDefined();
  });

  it('clearAll wipes every snapshot', () => {
    useDiagnosticsStore.getState().setSnapshot('org-1', '123', makeInspection());
    useDiagnosticsStore.getState().setSnapshot('org-2', '456', makeInspection());
    useDiagnosticsStore.getState().clearAll();
    expect(useDiagnosticsStore.getState().snapshots).toEqual({});
  });
});
