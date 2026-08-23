import {
  queryLogs,
  type LogEntry,
  type LogListResult,
  type LogEventType,
} from '@vyzorix/api-client';

interface RawLogFields {
  id: string;
  type: string;
  timestamp: number;
  data?: Record<string, unknown>;
}

interface RawLogConnection {
  events: RawLogFields[];
  pagination: {
    limit: number;
    hasMore: boolean;
    nextCursor?: string;
  };
}

function toDate(value?: number | null): Date {
  if (!value) return new Date();
  return new Date(value > 1e12 ? value : value * 1000);
}

function normalizeLogEntry(raw: RawLogFields, deviceId: string): LogEntry {
  return {
    id: raw.id,
    deviceId,
    eventType: (raw.type as LogEventType) ?? 'info',
    timestamp: toDate(raw.timestamp),
    data: raw.data,
  };
}

function extractObject<T>(response: unknown, key: string): T | null {
  // Apollo's Client.query resolves to the full result envelope
  // ({ data: { [operation]: ... }, loading, networkStatus, ... }).
  // Unwrap `data` first, then look up the root field (e.g. `deviceLogs`).
  const envelope = response as { data?: Record<string, unknown> } | Record<string, unknown> | null;
  if (!envelope) return null;
  const payload = (envelope as { data?: Record<string, unknown> }).data ?? envelope;
  const value = (payload as Record<string, unknown>)[key];
  return (value as T | undefined) ?? null;
}

export interface LogsViaGraphQLParams {
  type?: string;
  startTime?: number;
  endTime?: number;
  limit?: number;
  cursor?: string;
}

/** Maps the REST wire shape (DeviceLogEventListResult) onto the domain
 * LogListResult so hook consumers get one type across REST and GraphQL. */
export function normalizeDeviceLogList(
  raw: {
    events?: { id?: string; type?: string; timestamp?: number; data?: Record<string, unknown> }[];
    pagination?: { hasMore?: boolean; nextCursor?: string };
  },
  deviceId: string,
): LogListResult {
  const logs = (raw.events ?? []).map((entry): LogEntry => ({
    id: entry.id ?? '',
    deviceId,
    eventType: (entry.type ?? 'info') as LogEventType,
    timestamp: toDate(entry.timestamp),
    data: entry.data,
  }));
  return {
    logs,
    hasMore: raw.pagination?.hasMore ?? false,
    nextCursor: raw.pagination?.nextCursor,
  };
}

export async function fetchDeviceLogsViaGraphQL(
  organizationId: string,
  imei: string,
  params?: LogsViaGraphQLParams,
): Promise<LogListResult> {
  const response = await queryLogs({
    organizationId,
    imei,
    type: params?.type,
    startTime: params?.startTime,
    endTime: params?.endTime,
    limit: params?.limit,
    cursor: params?.cursor,
  });
  const connection = extractObject<RawLogConnection>(response, 'deviceLogs');
  const events = connection?.events ?? [];
  return {
    logs: events.map((raw) => normalizeLogEntry(raw, imei)),
    hasMore: connection?.pagination.hasMore ?? false,
    nextCursor: connection?.pagination.nextCursor,
  };
}
