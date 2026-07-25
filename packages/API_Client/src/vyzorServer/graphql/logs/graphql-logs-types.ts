export interface RawLogEntry {
  __typename?: "LogEntry";
  id: string;
  type: string;
  timestamp: number;
  data?: Record<string, unknown>;
}

export interface RawLogListItem {
  __typename?: "LogEntry";
  id: string;
  type: string;
  timestamp: number;
  data?: Record<string, unknown>;
}

export interface RawLogConnection {
  events: RawLogListItem[];
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
    timestamp: number;
    type: string;
    data?: Record<string, unknown>;
    deviceImei: string;
  }[];
}
