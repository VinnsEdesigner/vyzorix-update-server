/**
 * Logs Domain Index
 * 
 * Re-exports all logs domain types and mappers.
 */

// Types
export type {
  LogEntry,
  LogListItem,
  LogLevel,
  LogLevelInfo,
  LogFilters,
  LogLevelCounts,
  LogStats,
} from "./logs-entity";

export {
  LOG_LEVELS,
  getLogLevelInfo,
  isErrorLevel,
  isWarningLevel,
  formatLogMessage,
  filterLogsByLevel,
  calculateLogStats,
} from "./logs-entity";

// Mappers
export type {
  RawLogEntry,
  RawLogStats,
  RawPaginatedLogs,
} from "./logs-mappers";

export {
  logEntryFromRaw,
  logListItemFromRaw,
  logStatsFromRaw,
  logEntriesFromRaw,
  logListItemsFromRaw,
  logFiltersToParams,
  paramsToLogFilters,
} from "./logs-mappers";
