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
  const r = response as Record<string, unknown> | null;
  if (!r) return null;
  const value = r[key];
  return (value as T | undefined) ?? null;
}

export interface LogsViaGraphQLParams {
  type?: string;
  startTime?: number;
  endTime?: number;
  limit?: number;
  cursor?: string;
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
