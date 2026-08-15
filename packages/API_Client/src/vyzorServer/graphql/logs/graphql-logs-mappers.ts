import type {
  LogEntry,
  LogEventType,
  LogListResult,
  LogStats,
} from "../../../domain/logs";
import type {
  RawLogEntry,
  RawLogListItem,
  RawLogConnection,
  RawLogStats,
} from "./graphql-logs-types";

function parseTimestamp(value?: number | null): Date {
  if (!value) return new Date();
  return new Date(value * 1000);
}

export function logEntryFromRaw(raw: RawLogEntry): LogEntry {
  return {
    id: raw.id,
    deviceId: raw.id,
    eventType: (raw.type as LogEventType) ?? "info",
    timestamp: parseTimestamp(raw.timestamp),
    data: raw.data,
  };
}

export function logListItemFromRaw(raw: RawLogListItem): LogEntry {
  return {
    id: raw.id,
    deviceId: raw.id,
    eventType: (raw.type as LogEventType) ?? "info",
    timestamp: parseTimestamp(raw.timestamp),
    data: raw.data,
  };
}

export function logListResultFromRaw(raw: RawLogConnection): LogListResult {
  return {
    logs: raw.events.map(logListItemFromRaw),
    hasMore: raw.pagination.hasMore,
    nextCursor: raw.pagination.nextCursor,
  };
}

export function logStatsFromRaw(raw: RawLogStats): LogStats {
  return {
    total: raw.total ?? 0,
    byType: {
      connection: 0,
      command: 0,
      telemetry: 0,
      error: raw.byType?.error ?? 0,
      warning: raw.byType?.warn ?? 0,
    },
  };
}
