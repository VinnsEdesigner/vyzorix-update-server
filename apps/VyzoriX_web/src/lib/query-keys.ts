export const queryKeys = {
  me: ['me'] as const,
  myOrganizations: ['me', 'organizations'] as const,
  myInvitations: ['me', 'invitations'] as const,

  devices: (params?: Record<string, unknown>) => ['devices', params ?? {}] as const,
  device: (imei: string) => ['devices', imei] as const,
  deviceStatus: (imei: string) => ['devices', imei, 'status'] as const,
  deviceSettings: (imei: string) => ['devices', imei, 'settings'] as const,
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

  organizations: ['organizations'] as const,
  organization: (id: string) => ['organizations', id] as const,
  organizationMembers: (orgId: string) => ['organizations', orgId, 'members'] as const,
  organizationInvitations: (orgId: string, status?: string) =>
    ['organizations', orgId, 'invitations', status ?? 'all'] as const,
} as const;

export type QueryKeyFactory = typeof queryKeys;
