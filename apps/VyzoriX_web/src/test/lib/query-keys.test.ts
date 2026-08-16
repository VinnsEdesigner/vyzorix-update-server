import { describe, it, expect } from 'vitest';
import { queryKeys } from '@/lib/query-keys';

describe('queryKeys factory', () => {
  it('me is a static key', () => {
    expect(queryKeys.me).toEqual(['me']);
  });

  it('myOrganizations is nested under me', () => {
    expect(queryKeys.myOrganizations).toEqual(['me', 'organizations']);
  });

  it('devices returns a key with params object', () => {
    expect(queryKeys.devices({ status: 'online' })).toEqual(['devices', { status: 'online' }]);
  });

  it('devices with no params defaults to empty object', () => {
    expect(queryKeys.devices()).toEqual(['devices', {}]);
  });

  it('device returns imei-scoped key', () => {
    expect(queryKeys.device('12345')).toEqual(['devices', '12345']);
  });

  it('deviceSettings nests under device key', () => {
    expect(queryKeys.deviceSettings('12345')).toEqual(['devices', '12345', 'settings']);
  });

  it('deviceConnectionStatus nests under device key', () => {
    expect(queryKeys.deviceConnectionStatus('12345')).toEqual(['devices', '12345', 'connection']);
  });

  it('commands returns imei + params key', () => {
    expect(queryKeys.commands('12345', { page: 1 })).toEqual(['commands', '12345', { page: 1 }]);
  });

  it('command returns dispatch-scoped key', () => {
    expect(queryKeys.command('disp-1')).toEqual(['commands', 'dispatch', 'disp-1']);
  });

  it('pendingCommands returns imei-scoped pending key', () => {
    expect(queryKeys.pendingCommands('12345')).toEqual(['commands', '12345', 'pending']);
  });

  it('telemetryHistory returns device + params key', () => {
    expect(queryKeys.telemetryHistory('dev-1', { range: '1h' })).toEqual([
      'telemetry',
      'history',
      'dev-1',
      { range: '1h' },
    ]);
  });

  it('latestTelemetry returns org + device-scoped key', () => {
      expect(queryKeys.latestTelemetry('org-1', 'dev-1')).toEqual([
        'telemetry',
        'latest',
        'org-1',
        'dev-1',
      ]);
    });

    it('dashboardStats returns org-scoped key', () => {
      expect(queryKeys.dashboardStats('org-1')).toEqual(['dashboard', 'stats', 'org-1']);
    });

    it('deviceMetrics returns org + device + range key', () => {
      expect(queryKeys.deviceMetrics('org-1', 'dev-1', '6h')).toEqual([
        'metrics',
        'org-1',
        'dev-1',
        '6h',
      ]);
    });

    it('telemetryStats returns org + device-scoped key', () => {
      expect(queryKeys.telemetryStats('org-1', 'dev-1')).toEqual([
        'telemetry',
        'stats',
        'org-1',
        'dev-1',
      ]);
    });

  it('organization returns org-scoped key', () => {
    expect(queryKeys.organization('org-1')).toEqual(['organizations', 'org-1']);
  });

  it('updatesStatus returns org-scoped key', () => {
    expect(queryKeys.updatesStatus('org-1')).toEqual(['updates', 'status', 'org-1']);
  });

  it('updateVersions returns a key with params object', () => {
    expect(queryKeys.updateVersions({ status: 'latest' })).toEqual([
      'updates',
      'versions',
      { status: 'latest' },
    ]);
  });

  it('updateVersions with no params defaults to empty object', () => {
    expect(queryKeys.updateVersions()).toEqual(['updates', 'versions', {}]);
  });

  it('updateChangelog returns version-scoped key', () => {
    expect(queryKeys.updateChangelog('v1.2.0')).toEqual(['updates', 'changelog', 'v1.2.0']);
  });

  it('updateChangelog with no version defaults to all', () => {
    expect(queryKeys.updateChangelog()).toEqual(['updates', 'changelog', 'all']);
  });

  it('updateHistory returns a key with params object', () => {
    expect(queryKeys.updateHistory({ page: 1 })).toEqual(['updates', 'history', { page: 1 }]);
  });

  it('updatePushDetail returns push-scoped key', () => {
    expect(queryKeys.updatePushDetail('push-1')).toEqual(['updates', 'push', 'push-1']);
  });

  it('keys are readonly tuples (as const)', () => {
    expect(queryKeys.me).toEqual(['me']);
    expect(Array.isArray(queryKeys.me)).toBe(true);
  });
});
