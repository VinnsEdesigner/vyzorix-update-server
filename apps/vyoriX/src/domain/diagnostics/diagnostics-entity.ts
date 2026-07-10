export type DeviceStatus = "online" | "offline" | "deregistered";

export type FCMStatus = "valid" | "invalid" | "not_set";

export type WebSocketStatus = "connected" | "disconnected";

export type TimelineEventType =
  | "TELEMETRY"
  | "COMMAND_SENT"
  | "COMMAND_ACK"
  | "COMMAND_FAILED"
  | "CONNECTION_OPEN"
  | "CONNECTION_LOST"
  | "FCM_FALLBACK"
  | "RECONNECTED"
  | "THRESHOLD_BREACH"
  | "REGISTERED"
  | "DEREGISTERED"
  | "ERROR";

export type TimelineEventCategory = "telemetry" | "command" | "connection" | "error";

export interface IdentityInfo {
  imei: string;
  deviceName?: string;
  model?: string;
  manufacturer?: string;
}

export interface SoftwareInfo {
  osVersion?: string;
  appVersion?: string;
  securityPatch?: string;
  buildId?: string;
}

export interface RegistrationInfo {
  status: DeviceStatus;
  registeredAt?: Date;
  fcmTokenValid: boolean;
  fcmTokenRefreshedAt?: Date;
  commandSecretSet: boolean;
}

export interface ConnectionInfo {
  webSocketStatus: WebSocketStatus;
  connectedAt?: Date;
  fcmStatus: FCMStatus;
  lastSeen?: Date;
  clientIp?: string;
  protocol?: string;
}

export interface TelemetryInfo {
  lastTimestamp?: Date;
  framesToday: number;
  avgLatencyMs?: number;
  totalBytesToday?: number;
  sessionsToday: number;
}

export interface DeviceInspection {
  identity: IdentityInfo;
  software: SoftwareInfo;
  registration: RegistrationInfo;
  connection: ConnectionInfo;
  telemetry: TelemetryInfo;
}

export interface TimelineEvent {
  id: string;
  type: TimelineEventType;
  timestamp: Date;
  data: Record<string, unknown>;
}

export interface TimelineConnection {
  events: TimelineEvent[];
  hasMore: boolean;
  nextCursor?: string;
}

export interface TimelineFilters {
  eventType?: TimelineEventType;
  startTime?: Date;
  endTime?: Date;
  limit?: number;
}

export function getEventCategory(type: TimelineEventType): TimelineEventCategory {
  switch (type) {
    case "TELEMETRY":
      return "telemetry";
    case "COMMAND_SENT":
    case "COMMAND_ACK":
    case "COMMAND_FAILED":
      return "command";
    case "CONNECTION_OPEN":
    case "CONNECTION_LOST":
    case "FCM_FALLBACK":
    case "RECONNECTED":
      return "connection";
    case "THRESHOLD_BREACH":
    case "ERROR":
    case "REGISTERED":
    case "DEREGISTERED":
      return "error";
    default:
      return "error";
  }
}

export function getEventTypeLabel(type: TimelineEventType): string {
  const labels: Record<TimelineEventType, string> = {
    TELEMETRY: "Telemetry",
    COMMAND_SENT: "Command Sent",
    COMMAND_ACK: "Command Ack",
    COMMAND_FAILED: "Command Failed",
    CONNECTION_OPEN: "Connected",
    CONNECTION_LOST: "Disconnected",
    FCM_FALLBACK: "FCM Fallback",
    RECONNECTED: "Reconnected",
    THRESHOLD_BREACH: "Threshold Breach",
    REGISTERED: "Registered",
    DEREGISTERED: "Deregistered",
    ERROR: "Error",
  };
  return labels[type] ?? type;
}
