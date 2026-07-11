import type { LogEntry, LogEventType, LogListResult, LogStats } from "./logs-entity";

export interface RawLogEntry {
  id: string;
  deviceId: string;
  eventType: string;
  timestamp: number;
  data?: Record<string, unknown>;
}

export interface RawLogListResult {
  logs: RawLogEntry[];
  pagination: {
    limit: number;
    has_more: boolean;
    next_cursor?: string;
  };
}

export interface RawLogStats {
  total: number;
  by_type: {
    connection: number;
    command: number;
    telemetry: number;
    error: number;
    warning: number;
  };
}

function parseTimestamp(value?: number | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value > 1e12 ? value : value * 1000);
}

export function logEntryFromRaw(raw: RawLogEntry): LogEntry {
  return {
    id: raw.id,
    deviceId: raw.deviceId,
    eventType: (raw.eventType as LogEventType) ?? "info",
    timestamp: parseTimestamp(raw.timestamp) ?? new Date(),
    data: raw.data,
  };
}

export function logListResultFromRaw(raw: RawLogListResult): LogListResult {
  return {
    logs: raw.logs.map(logEntryFromRaw),
    hasMore: raw.pagination.has_more,
    nextCursor: raw.pagination.next_cursor,
  };
}

export function logStatsFromRaw(raw: RawLogStats): LogStats {
  return {
    total: raw.total ?? 0,
    byType: {
      connection: raw.by_type?.connection ?? 0,
      command: raw.by_type?.command ?? 0,
      telemetry: raw.by_type?.telemetry ?? 0,
      error: raw.by_type?.error ?? 0,
      warning: raw.by_type?.warning ?? 0,
    },
  };
}
