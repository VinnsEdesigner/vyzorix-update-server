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
  hasMore: boolean;
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

export function telemetryFromRaw(raw: RawTelemetryInfo): TelemetryInfo {
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
    telemetry: telemetryFromRaw(raw.telemetry),
  };
}

export function timelineResultFromRaw(raw: RawTimelineResult): TimelineResult {
  return {
    events: raw.events.map(timelineEventFromRaw),
    hasMore: raw.hasMore,
    nextCursor: raw.nextCursor,
  };
}
