import { describe, it, expect, beforeEach } from 'vitest';
import {
  buildVersion,
  buildPush,
  buildSyncState,
  buildDevice,
  buildDeviceListItem,
  buildDeviceStats,
  buildPushRequest,
  resetFixtureCounter,
} from '@/test/fixtures/vyzor-test-fixtures';

describe('vyzor-test-fixtures', () => {
  beforeEach(() => {
    resetFixtureCounter();
  });

  it('buildVersion returns valid UpdateVersion with defaults', () => {
    const v = buildVersion();
    expect(v.version).toBe('v1.2.0');
    expect(v.isLatest).toBe(true);
    expect(v.apkSize).toBeGreaterThan(0);
    expect(v.sha256).toHaveLength(64);
    expect(v.releaseDate).toBeInstanceOf(Date);
  });

  it('buildVersion accepts overrides', () => {
    const v = buildVersion({ version: 'v2.0.0', releaseType: 'major' });
    expect(v.version).toBe('v2.0.0');
    expect(v.releaseType).toBe('major');
  });

  it('buildPush returns valid UpdatePush', () => {
    const p = buildPush();
    expect(p.status).toBe('pending');
    expect(p.installType).toBe('immediate');
    expect(p.devices.total).toBe(5);
    expect(p.devices.pending).toBe(5);
  });

  it('buildSyncState returns valid SyncState', () => {
    const s = buildSyncState();
    expect(s.status).toBe('synced');
    expect(s.versionsFound).toBe(3);
  });

  it('buildDevice returns valid DeviceDetailResult', () => {
    const d = buildDevice();
    expect(d.imei).toHaveLength(15);
    expect(d.status).toBe('online');
    expect(typeof d.last_seen).toBe('number');
  });

  it('buildDeviceListItem returns valid DeviceListItem', () => {
    const d = buildDeviceListItem();
    expect(d.imei).toHaveLength(15);
    expect(d.status).toBe('online');
  });

  it('buildDeviceStats returns valid DashboardStats', () => {
    const s = buildDeviceStats();
    expect(s.total_devices).toBe(10);
    expect((s.online_devices ?? 0) + (s.offline_devices ?? 0)).toBe(s.total_devices);
  });

  it('buildPushRequest returns valid PushUpdateRequest', () => {
    const r = buildPushRequest();
    expect(r.version).toBe('v1.2.0');
    expect(r.deviceIds).toHaveLength(2);
    expect(r.installType).toBe('immediate');
  });

  it('generates unique IDs across calls', () => {
    const v1 = buildVersion();
    const v2 = buildVersion();
    expect(v1.id).not.toBe(v2.id);
  });
});
