/**
 * Diagnostics REST Endpoints
 * 
 * REST API client for device diagnostics operations.
 * Based on SERVER_BACKEND_DIAGNOSTICS_API.md specification.
 * Uses session-based authentication (cookies).
 */

import { restClient } from "../_shared/rest-client";
import type { PaginatedResult } from "@/domain/_shared";
import { offsetPaginationFromRaw } from "@/domain/_shared";

// ============================================================================
// API Paths
// ============================================================================

/**
 * Diagnostics API paths
 */
export const DIAGNOSTICS_PATHS = {
  // GET /v1/device/:imei/inspect - Full device inspection
  inspect: (imei: string) => `/v1/device/${imei}/inspect`,
  // GET /v1/device/:imei/timeline - Device timeline events
  timeline: (imei: string) => `/v1/device/${imei}/timeline`,
} as const;

// ============================================================================
// Raw Types
// ============================================================================

/**
 * Raw device identity from API
 */
export interface RawDeviceIdentity {
  imei: string;
  device_name: string;
  model: string;
  manufacturer: string;
}

/**
 * Raw device software from API
 */
export interface RawDeviceSoftware {
  os_version: string;
  app_version: string;
  security_patch: string;
  build_id: string;
}

/**
 * Raw device registration from API
 */
export interface RawDeviceRegistration {
  status: string;
  registered_at: number;
  fcm_token_valid: boolean;
  fcm_token_refreshed_at?: number;
  command_secret_set: boolean;
}

/**
 * Raw device connection from API
 */
export interface RawDeviceConnection {
  web_socket_status: string;
  connected_at?: number;
  fcm_status: string;
  last_seen: number;
  client_ip: string;
  protocol: string;
}

/**
 * Raw telemetry stats from API
 */
export interface RawTelemetryStats {
  last_timestamp: number;
  frames_today: number;
  avg_latency_ms: number;
  total_bytes_today: number;
  sessions_today: number;
}

/**
 * Raw device inspection from API
 */
export interface RawDeviceInspection {
  identity: RawDeviceIdentity;
  software: RawDeviceSoftware;
  registration: RawDeviceRegistration;
  connection: RawDeviceConnection;
  telemetry: RawTelemetryStats;
}

/**
 * Raw timeline event from API
 */
export interface RawTimelineEvent {
  id: string;
  timestamp: number;
  type: string;
  description: string;
  data?: Record<string, unknown>;
}

// ============================================================================
// Transformed Types
// ============================================================================

/**
 * Device identity
 */
export interface DeviceIdentity {
  imei: string;
  deviceName: string;
  model: string;
  manufacturer: string;
}

/**
 * Device software
 */
export interface DeviceSoftware {
  osVersion: string;
  appVersion: string;
  securityPatch: string;
  buildId: string;
}

/**
 * Device registration status
 */
export interface DeviceRegistration {
  status: "pending" | "approved" | "rejected" | "registered" | "deregistered";
  registeredAt: Date;
  fcmTokenValid: boolean;
  fcmTokenRefreshedAt: Date | null;
  commandSecretSet: boolean;
}

/**
 * WebSocket status
 */
export type WebSocketStatus = "connected" | "disconnected" | "reconnecting";

/**
 * FCM status
 */
export type FcmStatus = "valid" | "invalid" | "refreshed";

/**
 * Device connection
 */
export interface DeviceConnection {
  webSocketStatus: WebSocketStatus;
  connectedAt: Date | null;
  fcmStatus: FcmStatus;
  lastSeen: Date;
  clientIp: string;
  protocol: string;
}

/**
 * Telemetry stats
 */
export interface TelemetryStats {
  lastTimestamp: Date;
  framesToday: number;
  avgLatencyMs: number;
  totalBytesToday: number;
  sessionsToday: number;
}

/**
 * Device inspection
 */
export interface DeviceInspection {
  identity: DeviceIdentity;
  software: DeviceSoftware;
  registration: DeviceRegistration;
  connection: DeviceConnection;
  telemetry: TelemetryStats;
}

/**
 * Timeline event type
 */
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

/**
 * Timeline event
 */
export interface TimelineEvent {
  id: string;
  timestamp: Date;
  type: TimelineEventType;
  description: string;
  data?: Record<string, unknown>;
}

// ============================================================================
// Transform Functions
// ============================================================================

function identityFromRaw(raw: RawDeviceIdentity): DeviceIdentity {
  return {
    imei: raw.imei,
    deviceName: raw.device_name,
    model: raw.model,
    manufacturer: raw.manufacturer,
  };
}

function softwareFromRaw(raw: RawDeviceSoftware): DeviceSoftware {
  return {
    osVersion: raw.os_version,
    appVersion: raw.app_version,
    securityPatch: raw.security_patch,
    buildId: raw.build_id,
  };
}

function registrationFromRaw(raw: RawDeviceRegistration): DeviceRegistration {
  return {
    status: (raw.status as DeviceRegistration["status"]) ?? "registered",
    registeredAt: new Date(raw.registered_at),
    fcmTokenValid: raw.fcm_token_valid,
    fcmTokenRefreshedAt: raw.fcm_token_refreshed_at ? new Date(raw.fcm_token_refreshed_at) : null,
    commandSecretSet: raw.command_secret_set,
  };
}

function connectionFromRaw(raw: RawDeviceConnection): DeviceConnection {
  return {
    webSocketStatus: (raw.web_socket_status as WebSocketStatus) ?? "disconnected",
    connectedAt: raw.connected_at ? new Date(raw.connected_at) : null,
    fcmStatus: (raw.fcm_status as FcmStatus) ?? "valid",
    lastSeen: new Date(raw.last_seen),
    clientIp: raw.client_ip,
    protocol: raw.protocol,
  };
}

function telemetryStatsFromRaw(raw: RawTelemetryStats): TelemetryStats {
  return {
    lastTimestamp: new Date(raw.last_timestamp),
    framesToday: raw.frames_today,
    avgLatencyMs: raw.avg_latency_ms,
    totalBytesToday: raw.total_bytes_today,
    sessionsToday: raw.sessions_today,
  };
}

function inspectionFromRaw(raw: RawDeviceInspection): DeviceInspection {
  return {
    identity: identityFromRaw(raw.identity),
    software: softwareFromRaw(raw.software),
    registration: registrationFromRaw(raw.registration),
    connection: connectionFromRaw(raw.connection),
    telemetry: telemetryStatsFromRaw(raw.telemetry),
  };
}

function timelineEventFromRaw(raw: RawTimelineEvent): TimelineEvent {
  return {
    id: raw.id,
    timestamp: new Date(raw.timestamp),
    type: (raw.type as TimelineEventType) ?? "ERROR",
    description: raw.description,
    data: raw.data,
  };
}

// ============================================================================
// Fetch Operations
// ============================================================================

/**
 * Fetch full device inspection
 * GET /v1/device/:imei/inspect
 */
export async function fetchDeviceInspection(imei: string): Promise<DeviceInspection> {
  const data = await restClient.get<RawDeviceInspection>(DIAGNOSTICS_PATHS.inspect(imei));
  return inspectionFromRaw(data);
}

/**
 * Fetch device timeline events
 * GET /v1/device/:imei/timeline
 */
export async function fetchDeviceTimeline(
  imei: string,
  params?: {
    types?: TimelineEventType[];
    startTime?: number;
    endTime?: number;
    page?: number;
    limit?: number;
  }
): Promise<PaginatedResult<TimelineEvent[]>> {
  const data = await restClient.get<{
    events: RawTimelineEvent[];
    pagination: {
      page: number;
      limit: number;
      total: number;
      total_pages: number;
      has_more: boolean;
    };
  }>(DIAGNOSTICS_PATHS.timeline(imei), {
    types: params?.types?.join(","),
    start_time: params?.startTime,
    end_time: params?.endTime,
    page: params?.page,
    limit: params?.limit,
  });

  return {
    items: data.events.map(timelineEventFromRaw),
    pagination: offsetPaginationFromRaw(data.pagination),
  };
}