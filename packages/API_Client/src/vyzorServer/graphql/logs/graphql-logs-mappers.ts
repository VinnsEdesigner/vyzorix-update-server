import type {
  LogEntry,
  LogEventType,
  LogListResult,
  LogStats,
} from "@/domain/logs";
import type {
  RawLogEntry,
  RawLogListItem,
  RawLogConnection,
  RawLogStats,
} from "./graphql-logs-types";

function parseTimestamp(value?: string | null): Date {
  if (!value) return new Date();
  return new Date(value);
}

export function logEntryFromRaw(raw: RawLogEntry): LogEntry {
  return {
    id: raw.id,
    deviceId: raw.deviceImei,
    eventType: (raw.eventType as LogEventType) ?? "info",
    timestamp: parseTimestamp(raw.timestamp),
    data: raw.data ? JSON.parse(raw.data) : undefined,
  };
}

export function logListItemFromRaw(raw: RawLogListItem): LogEntry {
  return logEntryFromRaw(raw as RawLogEntry);
}

export function logListResultFromRaw(raw: RawLogConnection): LogListResult {
  return {
    logs: raw.logs.map(logListItemFromRaw),
    hasMore: raw.pagination.hasMore,
    nextCursor: raw.pagination.nextCursor,
  };
}

export function logStatsFromRaw(raw: RawLogStats): LogStats {
  return {
    total: raw.total ?? 0,
    byType: {
      connection: raw.byType?.connection ?? 0,
      command: raw.byType?.command ?? 0,
      telemetry: raw.byType?.telemetry ?? 0,
      error: raw.byType?.error ?? 0,
      warning: raw.byType?.warning ?? 0,
    },
  };
}
