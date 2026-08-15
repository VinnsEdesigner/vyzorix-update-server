import type {
  IdentityInfo,
  SoftwareInfo,
  RegistrationInfo,
  ConnectionInfo,
  TelemetryInfo,
  DeviceInspection,
  TimelineEvent,
  TimelineResult,
  TimelineEventType,
  DeviceStatus,
  FCMStatus,
  WebSocketStatus,
} from "./diagnostics-entity";

export interface RawIdentityInfo {
  imei: string;
  deviceName?: string;
  model?: string;
  manufacturer?: string;
}

export interface RawSoftwareInfo {
  osVersion?: string;
  appVersion?: string;
  securityPatch?: string;
  buildId?: string;
}

export interface RawRegistrationInfo {
  status: string;
  registeredAt?: number;
  fcmTokenValid: boolean;
  fcmTokenRefreshedAt?: number;
  commandSecretSet: boolean;
}

export interface RawConnectionInfo {
  webSocketStatus?: string;
  connectedAt?: number;
  fcmStatus?: string;
  lastSeen?: number;
  clientIp?: string;
  protocol?: string;
}

export interface RawTelemetryInfo {
  lastTimestamp?: number;
  framesToday: number;
  avgLatencyMs?: number;
  totalBytesToday?: number;
  sessionsToday: number;
}

export interface RawDeviceInspection {
  identity: RawIdentityInfo;
  software: RawSoftwareInfo;
  registration: RawRegistrationInfo;
  connection: RawConnectionInfo;
  telemetry: RawTelemetryInfo;
}

export interface RawTimelineEvent {
  id: string;
  deviceId: string;
  type: string;
  timestamp: number;
  data?: Record<string, unknown>;
}

export interface RawTimelineResult {
  events: RawTimelineEvent[];
  // Server REST nests pagination under `pagination`; GraphQL is flat. Accept both.
  pagination?: { limit?: number; hasMore?: boolean; nextCursor?: string };
  hasMore?: boolean;
  nextCursor?: string;
}

export function identityFromRaw(raw: RawIdentityInfo): IdentityInfo {
  return {
    imei: raw.imei,
    deviceName: raw.deviceName,
    model: raw.model,
    manufacturer: raw.manufacturer,
  };
}

export function softwareFromRaw(raw: RawSoftwareInfo): SoftwareInfo {
  return {
    osVersion: raw.osVersion,
    appVersion: raw.appVersion,
    securityPatch: raw.securityPatch,
    buildId: raw.buildId,
  };
}

export function registrationFromRaw(raw: RawRegistrationInfo): RegistrationInfo {
  return {
    status: (raw.status as DeviceStatus) ?? "offline",
    registeredAt: raw.registeredAt ? new Date(raw.registeredAt) : undefined,
    fcmTokenValid: raw.fcmTokenValid,
    fcmTokenRefreshedAt: raw.fcmTokenRefreshedAt ? new Date(raw.fcmTokenRefreshedAt) : undefined,
    commandSecretSet: raw.commandSecretSet,
  };
}

export function connectionFromRaw(raw: RawConnectionInfo): ConnectionInfo {
  return {
    webSocketStatus: (raw.webSocketStatus as WebSocketStatus) ?? "disconnected",
    connectedAt: raw.connectedAt ? new Date(raw.connectedAt) : undefined,
    fcmStatus: (raw.fcmStatus as FCMStatus) ?? "not_set",
    lastSeen: raw.lastSeen ? new Date(raw.lastSeen) : undefined,
    clientIp: raw.clientIp,
    protocol: raw.protocol,
  };
}

export function diagnosticTelemetryFromRaw(raw: RawTelemetryInfo): TelemetryInfo {
  return {
    lastTimestamp: raw.lastTimestamp ? new Date(raw.lastTimestamp) : undefined,
    framesToday: raw.framesToday ?? 0,
    avgLatencyMs: raw.avgLatencyMs,
    totalBytesToday: raw.totalBytesToday,
    sessionsToday: raw.sessionsToday ?? 0,
  };
}

export function timelineEventFromRaw(raw: RawTimelineEvent): TimelineEvent {
  return {
    id: raw.id,
    deviceId: raw.deviceId,
    type: raw.type as TimelineEventType,
    timestamp: new Date(raw.timestamp),
    data: raw.data ?? {},
  };
}

export function deviceInspectionFromRaw(raw: RawDeviceInspection): DeviceInspection {
  return {
    identity: identityFromRaw(raw.identity),
    software: softwareFromRaw(raw.software),
    registration: registrationFromRaw(raw.registration),
    connection: connectionFromRaw(raw.connection),
    telemetry: diagnosticTelemetryFromRaw(raw.telemetry),
  };
}

export function timelineResultFromRaw(raw: RawTimelineResult): TimelineResult {
  // Server REST nests pagination under `pagination`; GraphQL returns flat. Prefer nested.
  const pagination = raw.pagination;
  return {
    events: raw.events.map(timelineEventFromRaw),
    hasMore: pagination?.hasMore ?? raw.hasMore ?? false,
    nextCursor: pagination?.nextCursor ?? raw.nextCursor,
  };
}

// GraphQL-specific raw types (with string timestamps)
export interface RawGraphQLRegistrationInfo {
  status: string;
  registeredAt?: string | null;
  fcmTokenValid: boolean;
  fcmTokenRefreshedAt?: string | null;
  commandSecretSet: boolean;
}

export interface RawGraphQLConnectionInfo {
  webSocketStatus?: string;
  connectedAt?: string | null;
  fcmStatus?: string;
  lastSeen?: string | null;
  clientIp?: string;
  protocol?: string;
}

export interface RawGraphQLTelemetryInfo {
  lastTimestamp?: string | null;
  framesToday: number;
  avgLatencyMs?: number;
  totalBytesToday?: number;
  sessionsToday: number;
}

export interface RawGraphQLDeviceInspection {
  identity: RawIdentityInfo;
  software: RawSoftwareInfo;
  registration: RawGraphQLRegistrationInfo;
  connection: RawGraphQLConnectionInfo;
  telemetry: RawGraphQLTelemetryInfo;
}

export interface RawGraphQLTimelineEvent {
  id: string;
  type: string;
  timestamp: string;
  data?: Record<string, unknown>;
}

export interface RawGraphQLTimelineConnection {
  events: RawGraphQLTimelineEvent[];
  hasMore: boolean;
  nextCursor?: string;
}

// GraphQL mappers that parse string timestamps
export function graphqlRegistrationFromRaw(raw: RawGraphQLRegistrationInfo): RegistrationInfo {
  return {
    status: (raw.status as DeviceStatus) ?? "offline",
    registeredAt: raw.registeredAt ? new Date(raw.registeredAt) : undefined,
    fcmTokenValid: raw.fcmTokenValid,
    fcmTokenRefreshedAt: raw.fcmTokenRefreshedAt ? new Date(raw.fcmTokenRefreshedAt) : undefined,
    commandSecretSet: raw.commandSecretSet,
  };
}

export function graphqlConnectionFromRaw(raw: RawGraphQLConnectionInfo): ConnectionInfo {
  return {
    webSocketStatus: (raw.webSocketStatus as WebSocketStatus) ?? "disconnected",
    connectedAt: raw.connectedAt ? new Date(raw.connectedAt) : undefined,
    fcmStatus: (raw.fcmStatus as FCMStatus) ?? "not_set",
    lastSeen: raw.lastSeen ? new Date(raw.lastSeen) : undefined,
    clientIp: raw.clientIp,
    protocol: raw.protocol,
  };
}

export function graphqlTelemetryFromRaw(raw: RawGraphQLTelemetryInfo): TelemetryInfo {
  return {
    lastTimestamp: raw.lastTimestamp ? new Date(raw.lastTimestamp) : undefined,
    framesToday: raw.framesToday ?? 0,
    avgLatencyMs: raw.avgLatencyMs,
    totalBytesToday: raw.totalBytesToday,
    sessionsToday: raw.sessionsToday ?? 0,
  };
}

export function graphqlTimelineEventFromRaw(raw: RawGraphQLTimelineEvent): TimelineEvent {
  return {
    id: raw.id,
    deviceId: "", // GraphQL response doesn't include deviceId in TimelineEvent
    type: raw.type as TimelineEventType,
    timestamp: new Date(raw.timestamp),
    data: raw.data ?? {},
  };
}

export function graphqlDeviceInspectionFromRaw(raw: RawGraphQLDeviceInspection): DeviceInspection {
  return {
    identity: identityFromRaw(raw.identity),
    software: softwareFromRaw(raw.software),
    registration: graphqlRegistrationFromRaw(raw.registration),
    connection: graphqlConnectionFromRaw(raw.connection),
    telemetry: graphqlTelemetryFromRaw(raw.telemetry),
  };
}

export function graphqlTimelineResultFromRaw(raw: RawGraphQLTimelineConnection): TimelineResult {
  return {
    events: raw.events.map(graphqlTimelineEventFromRaw),
    hasMore: raw.hasMore,
    nextCursor: raw.nextCursor,
  };
}
