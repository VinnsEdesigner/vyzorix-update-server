import type { LogEntry, LogListItem, LogLevel, LogLevelCounts } from "@/domain/logs";

export type RawLogEntry = {
  __typename?: "LogEntry";
} & Omit<LogEntry, "timestamp"> & {
  timestamp: string;
};

export type RawLogListItem = {
  __typename?: "LogEntry";
} & Omit<LogListItem, "timestamp"> & {
  timestamp: string;
};

export interface RawLogConnection {
  logs: RawLogListItem[];
  pagination: {
    limit: number;
    hasMore: boolean;
    nextCursor?: string;
  };
}

export interface RawLogLevelCounts {
  debug: number;
  info: number;
  warn: number;
  error: number;
}

export interface RawLogStats {
  total: number;
  byType: RawLogLevelCounts;
}

export interface RawSubscribeToLogsResponse {
  logs: {
    id: string;
    timestamp: string;
    type: LogLevel;
    data: string;
    deviceImei: string;
  }[];
}
