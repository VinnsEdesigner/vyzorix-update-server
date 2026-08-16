export const queryKeys = {
  me: ['me'] as const,
  myOrganizations: ['me', 'organizations'] as const,
  myInvitations: ['me', 'invitations'] as const,

  devices: (params?: Record<string, unknown>) => ['devices', params ?? {}] as const,
  device: (imei: string) => ['devices', imei] as const,
  deviceStatus: (imei: string) => ['devices', imei, 'status'] as const,
  deviceSettings: (imei: string) => ['devices', imei, 'settings'] as const,
  deviceThresholds: (imei: string) => ['devices', imei, 'thresholds'] as const,
  deviceConnectionStatus: (imei: string) => ['devices', imei, 'connection'] as const,
  deviceCount: ['devices', 'count'] as const,
  deviceStats: ['devices', 'stats'] as const,

  dashboardStats: (organizationId: string) => ['dashboard', 'stats', organizationId] as const,
  deviceMetrics: (
    organizationId: string,
    imei: string,
    range?: string,
  ) => ['metrics', organizationId, imei, range ?? {}] as const,

  logs: (imei: string, params?: Record<string, unknown>) => ['logs', imei, params ?? {}] as const,
  logStats: (imei: string, params?: Record<string, unknown>) => ['logs', imei, 'stats', params ?? {}] as const,

  inspection: (organizationId: string, imei: string) => ['diagnostics', organizationId, imei, 'inspection'] as const,
  timeline: (organizationId: string, imei: string, params?: Record<string, unknown>) => ['diagnostics', organizationId, imei, 'timeline', params ?? {}] as const,

  events: (params?: Record<string, unknown>) => ['events', params ?? {}] as const,
  deviceEvents: (imei: string, params?: Record<string, unknown>) => ['events', 'device', imei, params ?? {}] as const,
  recentEvents: (limit?: number) => ['events', 'recent', limit ?? 20] as const,

  commands: (imei: string, params?: Record<string, unknown>) => ['commands', imei, params ?? {}] as const,
  command: (dispatchId: string) => ['commands', 'dispatch', dispatchId] as const,
  pendingCommands: (imei: string) => ['commands', imei, 'pending'] as const,

  apiKeys: (params?: Record<string, unknown>) => ['api-keys', params ?? {}] as const,
  apiKey: (id: string) => ['api-keys', id] as const,

  // Hierarchical API-key query keys (spec §5.1). Use these for invalidation
  // so partial keys (e.g. ['api-keys', 'list']) don't accidentally match detail
  // queries and vice-versa. The flat apiKeys()/apiKey() above are retained for
  // backward compatibility with existing call sites.
  apiKeysQueryKeys: {
    all: ['api-keys'] as const,
    lists: () => [...queryKeys.apiKeysQueryKeys.all, 'list'] as const,
    list: (params?: Record<string, unknown>) =>
      [...queryKeys.apiKeysQueryKeys.lists(), params ?? {}] as const,
    details: () => [...queryKeys.apiKeysQueryKeys.all, 'detail'] as const,
    detail: (keyId: string) => [...queryKeys.apiKeysQueryKeys.details(), keyId] as const,
  },

  // Admin API-key query keys (spec §12.5). Separate hierarchy so super-admin
  // invalidation never touches operator-scoped queries.
  adminApiKeysQueryKeys: {
    all: ['admin', 'api-keys'] as const,
    lists: () => [...queryKeys.adminApiKeysQueryKeys.all, 'list'] as const,
    list: (filters?: Record<string, unknown>) =>
      [...queryKeys.adminApiKeysQueryKeys.lists(), filters ?? {}] as const,
    operatorKeys: (operatorId: string, page?: number, limit?: number) =>
      [...queryKeys.adminApiKeysQueryKeys.all, 'operator', operatorId, { page, limit }] as const,
    stats: () => [...queryKeys.adminApiKeysQueryKeys.all, 'stats'] as const,
    operatorStats: (operatorId: string) =>
      [...queryKeys.adminApiKeysQueryKeys.all, 'stats', 'operator', operatorId] as const,
  },

  sessions: ['sessions'] as const,
  concurrentSessions: ['sessions', 'concurrent'] as const,

  settings: ['settings'] as const,
  thresholds: ['settings', 'thresholds'] as const,
  notifications: ['settings', 'notifications'] as const,

  registrationInbox: (params?: Record<string, unknown>) => ['registration', 'inbox', params ?? {}] as const,
  registrationInboxEntry: (imei: string) => ['registration', 'inbox', 'entry', imei] as const,
  registrationDevices: (params?: Record<string, unknown>) => ['registration', 'devices', params ?? {}] as const,
  registrationDevice: (imei: string) => ['registration', 'device', imei] as const,
  registrationStatus: (imei: string) => ['registration', 'status', imei] as const,

  telemetryHistory: (
    deviceId: string,
    params?: Record<string, unknown>,
  ) => ['telemetry', 'history', deviceId, params ?? {}] as const,
  latestTelemetry: (organizationId: string, deviceId: string) =>
    ['telemetry', 'latest', organizationId, deviceId] as const,
  telemetryStats: (organizationId: string, deviceId: string) =>
    ['telemetry', 'stats', organizationId, deviceId] as const,

  updatesStatus: (organizationId: string) => ['updates', 'status', organizationId] as const,
  updateVersions: (params?: Record<string, unknown>) => ['updates', 'versions', params ?? {}] as const,
  updateChangelog: (version?: string) => ['updates', 'changelog', version ?? 'all'] as const,
  updateHistory: (params?: Record<string, unknown>) => ['updates', 'history', params ?? {}] as const,
  updatePushDetail: (pushId: string) => ['updates', 'push', pushId] as const,

  organizations: ['organizations'] as const,
  organization: (id: string) => ['organizations', id] as const,
  orgSettings: (orgId: string) => ['organizations', orgId, 'settings'] as const,
  orgThresholds: (orgId: string) => ['organizations', orgId, 'thresholds'] as const,
  organizationMembers: (orgId: string) => ['organizations', orgId, 'members'] as const,
  organizationInvitations: (orgId: string, status?: string) =>
    ['organizations', orgId, 'invitations', status ?? 'all'] as const,
} as const;

export type QueryKeyFactory = typeof queryKeys;
