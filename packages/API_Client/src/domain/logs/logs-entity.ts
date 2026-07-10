/**
 * Logs Domain Types
 * 
 * Domain types for device logs.
 * Pure TypeScript - no external imports.
 */

// ============================================================================
// Log Levels
// ============================================================================

/**
 * Log severity levels
 */
export type LogLevel = "debug" | "info" | "warn" | "error";

/**
 * Log level metadata
 */
export interface LogLevelInfo {
  level: LogLevel;
  label: string;
  color: string;
  priority: number;
}

/**
 * All log levels with metadata
 */
export const LOG_LEVELS: Record<LogLevel, LogLevelInfo> = {
  debug: {
    level: "debug",
    label: "Debug",
    color: "gray",
    priority: 0,
  },
  info: {
    level: "info",
    label: "Info",
    color: "blue",
    priority: 1,
  },
  warn: {
    level: "warn",
    label: "Warning",
    color: "yellow",
    priority: 2,
  },
  error: {
    level: "error",
    label: "Error",
    color: "red",
    priority: 3,
  },
} as const;

// ============================================================================
// Log Entry
// ============================================================================

/**
 * Device log entry
 */
export interface LogEntry {
  id: string;
  timestamp: Date;
  level: LogLevel;
  deviceImei: string;
  message: string;
  source?: string;
  metadata?: Record<string, unknown>;
}

// ============================================================================
// Log List Item
// ============================================================================

/**
 * Lightweight log for list views
 */
export interface LogListItem {
  id: string;
  timestamp: Date;
  level: LogLevel;
  deviceImei: string;
  message: string;
  source?: string;
}

// ============================================================================
// Log Filters
// ============================================================================

/**
 * Log filter options
 */
export interface LogFilters {
  levels?: LogLevel[];
  imei?: string;
  search?: string;
  startDate?: Date;
  endDate?: Date;
}

// ============================================================================
// Log Statistics
// ============================================================================

/**
 * Log count by level
 */
export interface LogLevelCounts {
  debug: number;
  info: number;
  warn: number;
  error: number;
}

/**
 * Log statistics
 */
export interface LogStats {
  total: number;
  byLevel: LogLevelCounts;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Get log level info
 */
export function getLogLevelInfo(level: LogLevel): LogLevelInfo {
  return LOG_LEVELS[level];
}

/**
 * Check if log level is error or above
 */
export function isErrorLevel(level: LogLevel): boolean {
  return level === "error";
}

/**
 * Check if log level is warning or above
 */
export function isWarningLevel(level: LogLevel): boolean {
  return level === "warn" || level === "error";
}

/**
 * Format log message with timestamp
 */
export function formatLogMessage(entry: LogEntry | LogListItem): string {
  const timestamp = entry.timestamp.toISOString();
  const level = entry.level.toUpperCase().padEnd(5);
  return `[${timestamp}] ${level} ${entry.message}`;
}

/**
 * Filter logs by level
 */
export function filterLogsByLevel(
  logs: LogEntry[],
  levels: LogLevel[]
): LogEntry[] {
  if (!levels || levels.length === 0) return logs;
  return logs.filter((log) => levels.includes(log.level));
}

/**
 * Calculate log stats
 */
export function calculateLogStats(logs: LogEntry[]): LogStats {
  const counts: LogLevelCounts = {
    debug: 0,
    info: 0,
    warn: 0,
    error: 0,
  };
  
  for (const log of logs) {
    counts[log.level]++;
  }
  
  return {
    total: logs.length,
    byLevel: counts,
  };
}
