/**
 * Registration REST Client
 * 
 * REST API client for device registration/inbox operations.
 * Based on DEVICE_REGISTRATION_SYSTEM.md specification (frontend spec).
 * Uses session-based authentication (cookies).
 */

import { apiGet, apiPost, apiDelete } from "./_shared/rest-client";
import type { PaginatedResult } from "@/domain/_shared";
import { offsetPaginationFromRaw } from "@/domain/_shared";

// ============================================================================
// API Paths
// ============================================================================

export const REGISTRATION_PATHS = {
  inbox: "/v1/device/inbox",
  inboxEntry: (imei: string) => `/v1/device/inbox/${imei}`,
  inboxAck: (imei: string) => `/v1/device/inbox/${imei}/ack`,
  inboxDismiss: (imei: string) => `/v1/device/inbox/${imei}`,
  devices: "/v1/devices",
  device: (imei: string) => `/v1/devices/${imei}`,
  deregister: (imei: string) => `/v1/devices/${imei}`,
  register: "/v1/device/register",
  confirm: "/v1/device/confirm",
  telemetry: (imei: string) => `/v1/device/${imei}/telemetry`,
} as const;

// ============================================================================
// Raw Types (snake_case from API)
// ============================================================================

export interface RawInboxEntry {
  id?: string;
  imei?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  os_version?: string;
  app_version?: string;
  firmware?: string;
  security_patch?: string;
  build_id?: string;
  fcm_token?: string;
  firebase_install_id?: string;
  status?: string;
  received_at?: number;
  updated_at?: number;
  acknowledged_at?: number;
  approved_at?: number;
  rejected_at?: number;
  notes?: string;
}

export interface RawDevice {
  id?: string;
  imei?: string;
  device_name?: string;
  model?: string;
  manufacturer?: string;
  os_version?: string;
  app_version?: string;
  fcm_token?: string;
  status?: string;
  registered_at?: number;
  last_seen?: number;
  created_at?: number;
  updated_at?: number;
}

export interface RawTelemetryFrame {
  timestamp?: number;
  risk_score?: number;
  thermal_temp?: number;
  buffer_level?: number;
  uptime?: number;
}

// ============================================================================
// Domain Types
// ============================================================================

export type InboxStatus = "pending" | "acknowledged" | "approving" | "approved" | "rejected" | "expired";
export type DeviceStatus = "online" | "offline" | "deregistered";
export type AcknowledgeAction = "acknowledge" | "approve" | "reject";

export interface InboxEntry {
  id: string;
  imei: string;
  deviceName: string;
  model: string;
  manufacturer: string;
  osVersion: string;
  appVersion: string;
  firmware: string;
  securityPatch: string;
  buildId: string;
  fcmToken: string;
  firebaseInstallId: string;
  status: InboxStatus;
  receivedAt: Date;
  updatedAt: Date;
  acknowledgedAt: Date | null;
  approvedAt: Date | null;
  rejectedAt: Date | null;
  notes: string | null;
}

export interface Device {
  id: string;
  imei: string;
  deviceName: string;
  model: string;
  manufacturer: string;
  osVersion: string;
  appVersion: string;
  fcmToken: string;
  status: DeviceStatus;
  registeredAt: Date | null;
  lastSeen: Date | null;
}

export interface TelemetryFrame {
  timestamp: Date;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  uptime: number;
}

// ============================================================================
// Transform Functions
// ============================================================================

function parseTimestamp(value?: number | null): Date | null {
  if (!value) return null;
  return new Date(value > 1e12 ? value : value * 1000);
}

function inboxEntryFromRaw(raw: RawInboxEntry): InboxEntry {
  return {
    id: raw.id ?? "",
    imei: raw.imei ?? "",
    deviceName: raw.device_name ?? "",
    model: raw.model ?? "",
    manufacturer: raw.manufacturer ?? "",
    osVersion: raw.os_version ?? "",
    appVersion: raw.app_version ?? "",
    firmware: raw.firmware ?? "",
    securityPatch: raw.security_patch ?? "",
    buildId: raw.build_id ?? "",
    fcmToken: raw.fcm_token ?? "",
    firebaseInstallId: raw.firebase_install_id ?? "",
    status: (raw.status as InboxStatus) ?? "pending",
    receivedAt: parseTimestamp(raw.received_at) ?? new Date(),
    updatedAt: parseTimestamp(raw.updated_at) ?? new Date(),
    acknowledgedAt: parseTimestamp(raw.acknowledged_at),
    approvedAt: parseTimestamp(raw.approved_at),
    rejectedAt: parseTimestamp(raw.rejected_at),
    notes: raw.notes ?? null,
  };
}

function deviceFromRaw(raw: RawDevice): Device {
  return {
    id: raw.id ?? "",
    imei: raw.imei ?? "",
    deviceName: raw.device_name ?? "",
    model: raw.model ?? "",
    manufacturer: raw.manufacturer ?? "",
    osVersion: raw.os_version ?? "",
    appVersion: raw.app_version ?? "",
    fcmToken: raw.fcm_token ?? "",
    status: (raw.status as DeviceStatus) ?? "offline",
    registeredAt: parseTimestamp(raw.registered_at),
    lastSeen: parseTimestamp(raw.last_seen),
  };
}

function telemetryFrameFromRaw(raw: RawTelemetryFrame): TelemetryFrame {
  return {
    timestamp: parseTimestamp(raw.timestamp) ?? new Date(),
    riskScore: raw.risk_score ?? 0,
    thermalTemp: raw.thermal_temp ?? 0,
    bufferLevel: raw.buffer_level ?? 0,
    uptime: raw.uptime ?? 0,
  };
}

// ============================================================================
// Inbox Operations
// ============================================================================

export async function fetchInboxEntries(params?: {
  status?: InboxStatus | "all";
  page?: number;
  limit?: number;
}): Promise<PaginatedResult<InboxEntry[]>> {
  const data = await apiGet<{
    requests: RawInboxEntry[];
    pagination: { page: number; limit: number; total: number; total_pages: number; has_more?: boolean };
  }>(REGISTRATION_PATHS.inbox, { status: params?.status, page: params?.page, limit: params?.limit });

  return {
    items: data.requests.map(inboxEntryFromRaw),
    pagination: offsetPaginationFromRaw(data.pagination),
  };
}

export async function fetchInboxEntry(imei: string): Promise<InboxEntry | null> {
  const data = await apiGet<RawInboxEntry>(REGISTRATION_PATHS.inboxEntry(imei));
  if (!data || !data.imei) return null;
  return inboxEntryFromRaw(data);
}

export interface AcknowledgeRequest {
  action: AcknowledgeAction;
  notes?: string;
}

export interface AcknowledgeResponse {
  id: string;
  imei: string;
  status: InboxStatus;
  acknowledgedAt: Date | null;
  approvedAt: Date | null;
  commandSecret?: string;
  fcmPushSent?: boolean;
  notes: string | null;
}

export async function acknowledgeInbox(
  imei: string,
  request: AcknowledgeRequest
): Promise<AcknowledgeResponse> {
  const data = await apiPost<{
    id: string; imei: string; status: string;
    acknowledged_at?: number; approved_at?: number;
    command_secret?: string; fcm_push_sent?: boolean; notes?: string;
  }>(REGISTRATION_PATHS.inboxAck(imei), { action: request.action, notes: request.notes });

  return {
    id: data.id, imei: data.imei, status: (data.status as InboxStatus) ?? "pending",
    acknowledgedAt: parseTimestamp(data.acknowledged_at),
    approvedAt: parseTimestamp(data.approved_at),
    commandSecret: data.command_secret, fcmPushSent: data.fcm_push_sent, notes: data.notes ?? null,
  };
}

export async function dismissInboxEntry(imei: string): Promise<{ status: InboxStatus; updatedAt: Date }> {
  const data = await apiDelete<{ status: string; updated_at: number }>(REGISTRATION_PATHS.inboxDismiss(imei));
  return { status: (data.status as InboxStatus) ?? "rejected", updatedAt: parseTimestamp(data.updated_at) ?? new Date() };
}

// ============================================================================
// Device Operations
// ============================================================================

export async function fetchDevices(params?: {
  page?: number; limit?: number; status?: DeviceStatus | "all";
}): Promise<PaginatedResult<Device[]>> {
  const data = await apiGet<{
    devices: RawDevice[];
    pagination: { page: number; limit: number; total: number; total_pages: number; has_more?: boolean };
  }>(REGISTRATION_PATHS.devices, { page: params?.page, limit: params?.limit, status: params?.status });

  return { items: data.devices.map(deviceFromRaw), pagination: offsetPaginationFromRaw(data.pagination) };
}

export async function fetchDevice(imei: string): Promise<Device | null> {
  const data = await apiGet<RawDevice>(REGISTRATION_PATHS.device(imei));
  if (!data || !data.imei) return null;
  return deviceFromRaw(data);
}

export interface DeregisterResponse {
  imei: string; status: "deregistered"; deregisteredAt: Date; retentionUntil: Date;
}

export async function deregisterDevice(imei: string): Promise<DeregisterResponse> {
  const data = await apiDelete<{ imei: string; status: string; deregistered_at: number; retention_until: number }>(REGISTRATION_PATHS.deregister(imei));
  return { imei: data.imei, status: "deregistered", deregisteredAt: parseTimestamp(data.deregistered_at) ?? new Date(), retentionUntil: parseTimestamp(data.retention_until) ?? new Date() };
}

// ============================================================================
// Registration Operations
// ============================================================================

export interface RegisterDeviceRequest { imei: string; }
export interface RegisterDeviceResponse { status: "approving"; deviceId: string; message: string; }

export async function registerDevice(request: RegisterDeviceRequest): Promise<RegisterDeviceResponse> {
  const data = await apiPost<{ status: string; device_id: string; message: string }>(REGISTRATION_PATHS.register, { imei: request.imei });
  return { status: "approving", deviceId: data.device_id, message: data.message };
}

export interface ConfirmRegistrationRequest { imei: string; confirmed: boolean; }
export interface ConfirmRegistrationResponse { status: "registered"; deviceId: string; commandSecret: string; registeredAt: Date; }

export async function confirmRegistration(request: ConfirmRegistrationRequest): Promise<ConfirmRegistrationResponse> {
  const data = await apiPost<{ status: string; device_id: string; command_secret: string; registered_at: number }>(REGISTRATION_PATHS.confirm, { imei: request.imei, confirmed: request.confirmed });
  return { status: "registered", deviceId: data.device_id, commandSecret: data.command_secret, registeredAt: parseTimestamp(data.registered_at) ?? new Date() };
}

// ============================================================================
// Telemetry Operations
// ============================================================================

export interface TelemetryParams { startTime?: number; endTime?: number; limit?: number; }
export interface TelemetryResponse { frames: TelemetryFrame[]; pagination: { limit: number; hasMore: boolean }; }

export async function fetchDeviceTelemetry(imei: string, params?: TelemetryParams): Promise<TelemetryResponse> {
  const data = await apiGet<{ frames: RawTelemetryFrame[]; pagination: { limit: number; has_more: boolean } }>(REGISTRATION_PATHS.telemetry(imei), { start_time: params?.startTime, end_time: params?.endTime, limit: params?.limit });
  return { frames: data.frames.map(telemetryFrameFromRaw), pagination: { limit: data.pagination.limit, hasMore: data.pagination.has_more ?? false } };
}
