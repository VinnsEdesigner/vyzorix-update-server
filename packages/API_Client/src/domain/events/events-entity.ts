

export type EventType =
  | "DEVICE_CONNECTED"
  | "DEVICE_DISCONNECTED"
  | "DEVICE_RECONNECTED"
  | "TELEMETRY_RECEIVED"
  | "THRESHOLD_BREACH"
  | "RISK_SCORE_ALERT"
  | "THERMAL_ALERT"
  | "BUFFER_LEVEL_ALERT"
  | "THRESHOLD_RESOLVED"
  | "COMMAND_SENT"
  | "COMMAND_DELIVERED"
  | "COMMAND_FAILED"
  | "COMMAND_ACKNOWLEDGED"
  | "FCM_FALLBACK"
  | "DEVICE_OFFLINE"
  | "DEVICE_ONLINE"
  | "ERROR"
  | "REGISTRATION"
  | "DEREGISTRATION";

export type Severity = "info" | "warning" | "critical";


export interface Event {
  id: string;
  deviceId: string;
  operatorId?: string;
  type: EventType;
  severity: Severity;
  timestamp: Date;
  data?: Record<string, unknown>;
  source: "device" | "server" | "dashboard";
}


export interface EventResult {
  events: Event[];
  hasMore: boolean;
  count: number;
  totalCount?: number;
}


export interface EventFilter {
  types?: EventType[];
  severities?: Severity[];
  limit?: number;
  offset?: number;
  startTime?: Date;
  endTime?: Date;
}


export interface EventParams {
  types?: string[];
  severities?: string[];
  limit?: number;
  offset?: number;
  startTime?: number;
  endTime?: number;
}


export function isConnectivityEvent(type: EventType): boolean {
  return [
    "DEVICE_CONNECTED",
    "DEVICE_DISCONNECTED",
    "DEVICE_RECONNECTED",
    "DEVICE_OFFLINE",
    "DEVICE_ONLINE",
  ].includes(type);
}


export function isTelemetryEvent(type: EventType): boolean {
  return [
    "TELEMETRY_RECEIVED",
    "THRESHOLD_BREACH",
    "RISK_SCORE_ALERT",
    "THERMAL_ALERT",
    "BUFFER_LEVEL_ALERT",
    "THRESHOLD_RESOLVED",
  ].includes(type);
}


export function isCommandEvent(type: EventType): boolean {
  return [
    "COMMAND_SENT",
    "COMMAND_DELIVERED",
    "COMMAND_FAILED",
    "COMMAND_ACKNOWLEDGED",
  ].includes(type);
}


export function getDefaultSeverity(type: EventType): Severity {
  const criticalTypes: EventType[] = [
    "THRESHOLD_BREACH",
    "RISK_SCORE_ALERT",
    "THERMAL_ALERT",
    "COMMAND_FAILED",
    "FCM_FALLBACK",
    "DEVICE_OFFLINE",
    "ERROR",
    "DEREGISTRATION",
  ];
  const warningTypes: EventType[] = [
    "BUFFER_LEVEL_ALERT",
    "DEVICE_DISCONNECTED",
  ];

  if (criticalTypes.includes(type)) return "critical";
  if (warningTypes.includes(type)) return "warning";
  return "info";
}
