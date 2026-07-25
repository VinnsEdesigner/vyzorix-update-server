

import { restClient, getOrganizationContext } from "../_shared/rest-client";
import type { TelemetryFrame } from "@/domain/telemetry";

const TELEMETRY_PATHS = {
  history: "/v1/telemetry/history",
  latest: (deviceId: string) => `/v1/telemetry/latest/${deviceId}`,
  stats: (deviceId: string) => `/v1/telemetry/stats/${deviceId}`,
  cleanup: "/v1/telemetry/cleanup",
} as const;

function parseTimestamp(value: string | number): Date {
  if (typeof value === "number") {
    return new Date(value > 1e12 ? value : value * 1000);
  }
  return new Date(value);
}

// Raw response from GET /v1/telemetry/history
export interface RawTelemetryEntry {
  id: string;
  deviceId: string;
  receivedAt: string;
  payload?: string;
  riskScore: number;
  bufferLevel: number;
  thermalTemp: number;
}

export interface RawTelemetryHistoryResponse {
  deviceId: string;
  entries: RawTelemetryEntry[];
  totalCount: number;
  startTime: number;
  endTime: number;
  queryTime: number;
}

// Raw response from GET /v1/telemetry/latest/:deviceId
export interface RawLatestTelemetry {
  id: string;
  deviceId: string;
  receivedAt: string;
  riskScore: number;
  bufferLevel: number;
  thermalTemp: number;
}

// Raw response from GET /v1/telemetry/stats/:deviceId
export interface RawTelemetryStats {
  deviceId: string;
  sampleCount: number;
  latestEntry: string;
  oldestEntry: string;
  riskScore: { avg: number; min: number; max: number };
  bufferLevel: { avg: number };
  thermalTemp: { avg: number; min: number; max: number };
}

// Parsed response types
export interface TelemetryEntry {
  id: string;
  deviceId: string;
  receivedAt: Date;
  payload?: string;
  riskScore: number;
  bufferLevel: number;
  thermalTemp: number;
}

export interface TelemetryHistoryResponse {
  deviceId: string;
  entries: TelemetryEntry[];
  totalCount: number;
  startTime: number;
  endTime: number;
  queryTime: number;
}

export interface LatestTelemetry {
  id: string;
  deviceId: string;
  receivedAt: Date;
  riskScore: number;
  bufferLevel: number;
  thermalTemp: number;
}

export interface TelemetryStats {
  deviceId: string;
  sampleCount: number;
  latestEntry: Date;
  oldestEntry: Date;
  riskScore: { avg: number; min: number; max: number };
  bufferLevel: { avg: number };
  thermalTemp: { avg: number; min: number; max: number };
}

export interface TelemetryHistoryParams {
  deviceId: string;
  startTime?: number;
  endTime?: number;
  limit?: number;
  format?: "json" | "csv";
}

// Mapper functions
function telemetryEntryFromRaw(raw: RawTelemetryEntry): TelemetryEntry {
  return {
    id: raw.id,
    deviceId: raw.deviceId,
    receivedAt: parseTimestamp(raw.receivedAt),
    payload: raw.payload,
    riskScore: raw.riskScore,
    bufferLevel: raw.bufferLevel,
    thermalTemp: raw.thermalTemp,
  };
}

function telemetryHistoryResponseFromRaw(raw: RawTelemetryHistoryResponse): TelemetryHistoryResponse {
  return {
    deviceId: raw.deviceId,
    entries: raw.entries.map(telemetryEntryFromRaw),
    totalCount: raw.totalCount,
    startTime: raw.startTime,
    endTime: raw.endTime,
    queryTime: raw.queryTime,
  };
}

function latestTelemetryFromRaw(raw: RawLatestTelemetry): LatestTelemetry {
  return {
    id: raw.id,
    deviceId: raw.deviceId,
    receivedAt: parseTimestamp(raw.receivedAt),
    riskScore: raw.riskScore,
    bufferLevel: raw.bufferLevel,
    thermalTemp: raw.thermalTemp,
  };
}

function telemetryStatsFromRaw(raw: RawTelemetryStats): TelemetryStats {
  return {
    deviceId: raw.deviceId,
    sampleCount: raw.sampleCount,
    latestEntry: parseTimestamp(raw.latestEntry),
    oldestEntry: parseTimestamp(raw.oldestEntry),
    riskScore: raw.riskScore,
    bufferLevel: raw.bufferLevel,
    thermalTemp: raw.thermalTemp,
  };
}

// Export mappers for use elsewhere
export {
  telemetryEntryFromRaw,
  telemetryHistoryResponseFromRaw,
  latestTelemetryFromRaw,
  telemetryStatsFromRaw,
};

// API functions
export async function queryTelemetryHistory(
  params: TelemetryHistoryParams & { organizationId?: string }
): Promise<TelemetryHistoryResponse> {
  const queryParams = new URLSearchParams();
  const orgId = params.organizationId || getOrganizationContext();
  queryParams.set("deviceId", params.deviceId);
  if (orgId) queryParams.set("organization_id", orgId);
  
  if (params.startTime) queryParams.set("startTime", params.startTime.toString());
  if (params.endTime) queryParams.set("endTime", params.endTime.toString());
  if (params.limit) queryParams.set("limit", params.limit.toString());
  if (params.format) queryParams.set("format", params.format);

  const response = await restClient.get<RawTelemetryHistoryResponse>(
    `${TELEMETRY_PATHS.history}?${queryParams.toString()}`
  );
  
  return telemetryHistoryResponseFromRaw(response);
}

export async function getLatestTelemetry(deviceId: string, organizationId?: string): Promise<LatestTelemetry> {
  const orgId = organizationId || getOrganizationContext();
  const response = await restClient.get<RawLatestTelemetry>(TELEMETRY_PATHS.latest(deviceId), {
    params: { organization_id: orgId },
  });
  return latestTelemetryFromRaw(response);
}

export async function getTelemetryStats(deviceId: string, organizationId?: string): Promise<TelemetryStats> {
  const orgId = organizationId || getOrganizationContext();
  const response = await restClient.get<RawTelemetryStats>(TELEMETRY_PATHS.stats(deviceId), {
    params: { organization_id: orgId },
  });
  return telemetryStatsFromRaw(response);
}

export async function exportTelemetry(params: TelemetryHistoryParams & { organizationId?: string }): Promise<string> {
  const queryParams = new URLSearchParams();
  const orgId = params.organizationId || getOrganizationContext();
  queryParams.set("deviceId", params.deviceId);
  if (orgId) queryParams.set("organization_id", orgId);
  
  if (params.startTime) queryParams.set("startTime", params.startTime.toString());
  if (params.endTime) queryParams.set("endTime", params.endTime.toString());
  if (params.limit) queryParams.set("limit", params.limit.toString());
  if (params.format) queryParams.set("format", params.format);

  return restClient.get<string>(
    `${TELEMETRY_PATHS.history}?${queryParams.toString()}`
  );
}

export async function cleanupOldTelemetry(olderThanTimestamp: number, organizationId?: string): Promise<{
  deleted: number;
  olderThan: number;
}> {
  const orgId = organizationId || getOrganizationContext();
  return restClient.delete<{ deleted: number; olderThan: number }>(
    `${TELEMETRY_PATHS.cleanup}?olderThan=${olderThanTimestamp}`,
    { params: { organization_id: orgId } }
  );
}
