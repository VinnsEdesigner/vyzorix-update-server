/**
 * Logs Mappers
 * 
 * Transformations from raw API response to domain types.
 * Raw API uses snake_case, domain uses camelCase.
 */

import type {
  LogEntry,
  LogListItem,
  LogLevel,
  LogFilters,
  LogStats,
  LogLevelCounts,
} from "./logs-entity";

// ============================================================================
// Raw API Types (snake_case)
// ============================================================================

/**
 * Raw log entry from API
 */
export interface RawLogEntry {
  id?: string;
  timestamp?: string | number;
  level?: string;
  device_imei?: string;
  message?: string;
  source?: string;
  metadata?: Record<string, unknown>;
}

/**
 * Raw log stats from API
 */
export interface RawLogStats {
  total?: number;
  debug?: number;
  info?: number;
  warn?: number;
  error?: number;
}

/**
 * Raw paginated logs response
 */
export interface RawPaginatedLogs {
  logs: RawLogEntry[];
  pagination: {
    page: number;
    limit: number;
    total: number;
    total_pages: number;
  };
}

// ============================================================================
// Transform Helpers
// ============================================================================

/**
 * Parse timestamp from various formats
 */
function parseTimestamp(value?: string | number | null): Date {
  if (!value) return new Date();
  
  if (typeof value === "number") {
    return new Date(value > 1e12 ? value : value * 1000);
  }
  
  return new Date(value);
}

// ============================================================================
// Transform Functions
// ============================================================================

/**
 * Transform raw log entry to domain
 */
export function logEntryFromRaw(raw: RawLogEntry): LogEntry {
  return {
    id: raw.id ?? "",
    timestamp: parseTimestamp(raw.timestamp),
    level: (raw.level as LogLevel) ?? "info",
    deviceImei: raw.device_imei ?? "",
    message: raw.message ?? "",
    source: raw.source,
    metadata: raw.metadata,
  };
}

/**
 * Transform raw log list item to domain
 */
export function logListItemFromRaw(raw: RawLogEntry): LogListItem {
  return {
    id: raw.id ?? "",
    timestamp: parseTimestamp(raw.timestamp),
    level: (raw.level as LogLevel) ?? "info",
    deviceImei: raw.device_imei ?? "",
    message: raw.message ?? "",
    source: raw.source,
  };
}

/**
 * Transform raw log stats to domain
 */
export function logStatsFromRaw(raw?: RawLogStats | null): LogStats {
  if (!raw) {
    return {
      total: 0,
      byLevel: { debug: 0, info: 0, warn: 0, error: 0 },
    };
  }
  
  return {
    total: raw.total ?? 0,
    byLevel: {
      debug: raw.debug ?? 0,
      info: raw.info ?? 0,
      warn: raw.warn ?? 0,
      error: raw.error ?? 0,
    },
  };
}

// ============================================================================
// Array Transformers
// ============================================================================

/**
 * Transform array of raw log entries
 */
export function logEntriesFromRaw(raw: RawLogEntry[]): LogEntry[] {
  return raw.map(logEntryFromRaw);
}

/**
 * Transform array of raw log list items
 */
export function logListItemsFromRaw(raw: RawLogEntry[]): LogListItem[] {
  return raw.map(logListItemFromRaw);
}

// ============================================================================
// Filter Transformers
// ============================================================================

/**
 * Transform log filters to API params
 */
export function logFiltersToParams(filters: LogFilters): Record<string, string | number | boolean | undefined> {
  const params: Record<string, string | number | boolean | undefined> = {};
  
  if (filters.levels && filters.levels.length > 0) {
    params.levels = filters.levels.join(",");
  }
  
  if (filters.imei) {
    params.imei = filters.imei;
  }
  
  if (filters.search) {
    params.search = filters.search;
  }
  
  if (filters.startDate) {
    params.start_date = filters.startDate.toISOString();
  }
  
  if (filters.endDate) {
    params.end_date = filters.endDate.toISOString();
  }
  
  return params;
}

/**
 * Transform API params to log filters
 */
export function paramsToLogFilters(params: Record<string, string>): LogFilters {
  const filters: LogFilters = {};
  
  if (params.levels) {
    filters.levels = params.levels.split(",") as LogLevel[];
  }
  
  if (params.imei) {
    filters.imei = params.imei;
  }
  
  if (params.search) {
    filters.search = params.search;
  }
  
  if (params.start_date) {
    filters.startDate = new Date(params.start_date);
  }
  
  if (params.end_date) {
    filters.endDate = new Date(params.end_date);
  }
  
  return filters;
}
