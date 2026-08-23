import type {
  UpdateVersion,
  UpdatePush,
  SyncState,
  UpdatePushRequest,
  Thresholds,
  ClientSettings,
  NotificationSettings,
  DeviceDetailResult,
  DeviceListItem,
  DashboardStats,
  OrganizationSettingsResult,
  DeviceSettingsResult,
} from '@vyzorix/api-client';

let counter = 0;
function nextId(prefix: string): string {
  counter += 1;
  return `${prefix}-test-${counter}`;
}

export function buildVersion(overrides: Partial<UpdateVersion> = {}): UpdateVersion {
  const now = new Date('2025-01-15T10:00:00Z');
  return {
    id: nextId('version'),
    version: 'v1.2.0',
    apkFilename: 'vyzorix-v1.2.0.apk',
    apkSize: 25_000_000,
    sha256: 'a'.repeat(64),
    releaseType: 'minor',
    releaseNotes: 'Bug fixes and performance improvements',
    releaseDate: now,
    isLatest: true,
    createdAt: now,
    updatedAt: now,
    ...overrides,
  };
}

export function buildSyncState(overrides: Partial<SyncState> = {}): SyncState {
  return {
    status: 'synced',
    lastSyncAt: new Date('2025-01-15T09:30:00Z'),
    versionsFound: 3,
    ...overrides,
  };
}

export function buildPush(overrides: Partial<UpdatePush> = {}): UpdatePush {
  const now = new Date('2025-01-15T10:30:00Z');
  return {
    id: nextId('push'),
    version: 'v1.2.0',
    installType: 'immediate',
    status: 'pending',
    initiatedBy: 'operator-test-1',
    initiatedAt: now,
    devices: {
      total: 5,
      pending: 5,
      sent: 0,
      acknowledged: 0,
      failed: 0,
    },
    ...overrides,
  };
}

export function buildPushRequest(
  overrides: Partial<UpdatePushRequest> = {},
): UpdatePushRequest {
  return {
    version: 'v1.2.0',
    deviceIds: ['device-test-1', 'device-test-2'],
    installType: 'immediate',
    ...overrides,
  };
}

export function buildDevice(overrides: Partial<DeviceDetailResult> = {}): DeviceDetailResult {
  const now = new Date('2025-01-15T10:00:00Z').getTime();
  return {
    id: nextId('device'),
    imei: '123456789012345',
    device_name: 'Test Device',
    model: 'VyzoriX Pro',
    manufacturer: 'VyzoriX',
    app_version: 'v1.1.0',
    status: 'online',
    registered_at: now - 86_400_000,
    last_seen: now,
    ...overrides,
  };
}

export function buildDeviceListItem(
  overrides: Partial<DeviceListItem> = {},
): DeviceListItem {
  return {
    id: nextId('device'),
    imei: '123456789012345',
    device_name: 'Test Device',
    model: 'VyzoriX Pro',
    manufacturer: 'VyzoriX',
    app_version: 'v1.1.0',
    status: 'online',
    online: true,
    last_seen: new Date('2025-01-15T10:00:00Z').getTime(),
    ...overrides,
  };
}

export function buildDeviceStats(overrides: Partial<DashboardStats> = {}): DashboardStats {
  return {
    total_devices: 10,
    online_devices: 7,
    offline_devices: 3,
    pending_devices: 0,
    ...overrides,
  };
}

export function buildThresholds(overrides: Partial<Thresholds> = {}): Thresholds {
  return {
    riskWarn: 70,
    riskCrit: 90,
    thermalWarn: 75,
    thermalCrit: 85,
    bufferWarn: 80,
    bufferCrit: 95,
    ...overrides,
  };
}

export function buildClientSettings(overrides: Partial<ClientSettings> = {}): ClientSettings {
  return {
    serverUrl: 'https://api.vyzorix.com',
    deviceId: 'device-test-1',
    requestTimeoutMs: 8000,
    logBufferLimit: 500,
    signalHistoryLimit: 100,
    autoReconnect: true,
    strictHmac: false,
    notificationsEnabled: true,
    ...overrides,
  };
}

export function buildNotificationSettings(
  overrides: Partial<NotificationSettings> = {},
): NotificationSettings {
  return {
    enabled: true,
    channels: ['email', 'push', 'webhook'],
    email: {
      thresholdBreach: true,
      deviceOffline: true,
      deviceOnline: false,
      updateAvailable: true,
      commandFailed: true,
      registrationRequest: true,
    },
    push: {
      thresholdBreach: true,
      deviceOffline: true,
      deviceOnline: false,
      updateAvailable: true,
      commandFailed: true,
      registrationRequest: false,
    },
    webhook: {
      enabled: true,
      url: 'https://hooks.example.com/vyzorix',
      secret: 'whsec_test_secret',
      types: ['threshold_breach', 'device_offline'],
    },
    ...overrides,
  };
}

export function buildOperatorSettings() {
  return {
    client: buildClientSettings(),
    thresholds: buildThresholds(),
    notifications: buildNotificationSettings(),
  };
}

export function buildOrgSettings(
  overrides: Partial<OrganizationSettingsResult> = {},
): OrganizationSettingsResult {
  const now = '2025-01-15T10:00:00Z';
  return {
    id: nextId('org_set'),
    organizationId: 'org-test-1',
    timezone: 'UTC',
    dateFormat: 'YYYY-MM-DD',
    alertCooldownMinutes: 15,
    defaultThresholds: buildThresholds(),
    createdAt: now,
    updatedAt: now,
    ...overrides,
  };
}

export function buildDeviceSettings(
  overrides: Partial<DeviceSettingsResult> = {},
): DeviceSettingsResult {
  const now = '2025-01-15T10:00:00Z';
  return {
    id: nextId('dev_set'),
    deviceImei: '123456789012345',
    customName: 'Test Device',
    location: 'Stockholm, SE',
    metadata: { firmware: 'v1.1.0', hardware: 'v2' },
    thresholds: buildThresholds(),
    createdAt: now,
    updatedAt: now,
    ...overrides,
  };
}

export function resetFixtureCounter(): void {
  counter = 0;
}
