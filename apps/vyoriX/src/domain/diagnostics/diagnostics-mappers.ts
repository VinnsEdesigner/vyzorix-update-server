import type {
  DeviceInspection,
  TimelineEvent,
  TimelineConnection,
  IdentityInfo,
  SoftwareInfo,
  RegistrationInfo,
  ConnectionInfo,
  TelemetryInfo,
  TimelineEventType,
  DeviceStatus,
  FCMStatus,
  WebSocketStatus,
} from "./diagnostics-entity";

export type RawIdentityInfo = {
  imei: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
};

export type RawSoftwareInfo = {
  os_version?: string;
  app_version?: string;
  security_patch?: string;
  build_id?: string;
};

export type RawRegistrationInfo = {
  status: string;
  registered_at?: number;
  fcm_token_valid: boolean;
  fcm_token_refreshed_at?: number;
  command_secret_set: boolean;
};

export type RawConnectionInfo = {
  web_socket_status?: string;
  connected_at?: number;
  fcm_status?: string;
  last_seen?: number;
  client_ip?: string;
  protocol?: string;
};

export type RawTelemetryInfo = {
  last_timestamp?: number;
  frames_today?: number;
  avg_latency_ms?: number;
  total_bytes_today?: number;
  sessions_today?: number;
};

export type RawDeviceInspection = {
  identity: RawIdentityInfo;
  software: RawSoftwareInfo;
  registration: RawRegistrationInfo;
  connection: RawConnectionInfo;
  telemetry: RawTelemetryInfo;
};

export type RawTimelineEvent = {
  id: string;
  type: string;
  timestamp: number;
  data: Record<string, unknown>;
};

export type RawTimelineConnection = {
  events: RawTimelineEvent[];
  pagination: {
    limit: number;
    has_more: boolean;
    next_cursor?: string;
  };
};

export type RawTimelineResponse = {
  events: RawTimelineEvent[];
  has_more: boolean;
  next_cursor?: string;
};

function identityFromRaw(raw: RawIdentityInfo): IdentityInfo {
  return {
    imei: raw.imei,
    deviceName: raw.device_name,
    model: raw.model,
    manufacturer: raw.manufacturer,
  };
}

function softwareFromRaw(raw: RawSoftwareInfo): SoftwareInfo {
  return {
    osVersion: raw.os_version,
    appVersion: raw.app_version,
    securityPatch: raw.security_patch,
    buildId: raw.build_id,
  };
}

function registrationFromRaw(raw: RawRegistrationInfo): RegistrationInfo {
  return {
    status: (raw.status as DeviceStatus) ?? "offline",
    registeredAt: raw.registered_at ? new Date(raw.registered_at) : undefined,
    fcmTokenValid: raw.fcm_token_valid,
    fcmTokenRefreshedAt: raw.fcm_token_refreshed_at ? new Date(raw.fcm_token_refreshed_at) : undefined,
    commandSecretSet: raw.command_secret_set,
  };
}

function connectionFromRaw(raw: RawConnectionInfo): ConnectionInfo {
  return {
    webSocketStatus: (raw.web_socket_status as WebSocketStatus) ?? "disconnected",
    connectedAt: raw.connected_at ? new Date(raw.connected_at) : undefined,
    fcmStatus: (raw.fcm_status as FCMStatus) ?? "not_set",
    lastSeen: raw.last_seen ? new Date(raw.last_seen) : undefined,
    clientIp: raw.client_ip,
    protocol: raw.protocol,
  };
}

function telemetryFromRaw(raw: RawTelemetryInfo): TelemetryInfo {
  return {
    lastTimestamp: raw.last_timestamp ? new Date(raw.last_timestamp) : undefined,
    framesToday: raw.frames_today ?? 0,
    avgLatencyMs: raw.avg_latency_ms,
    totalBytesToday: raw.total_bytes_today,
    sessionsToday: raw.sessions_today ?? 0,
  };
}

function timelineEventFromRaw(raw: RawTimelineEvent): TimelineEvent {
  return {
    id: raw.id,
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

export function timelineConnectionFromRaw(raw: RawTimelineResponse): TimelineConnection {
  return {
    events: raw.events.map(timelineEventFromRaw),
    hasMore: raw.has_more,
    nextCursor: raw.next_cursor,
  };
}

export function timelineEventsFromRaw(raw: RawTimelineEvent[]): TimelineEvent[] {
  return raw.map(timelineEventFromRaw);
}
