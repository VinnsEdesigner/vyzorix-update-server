import { restClient } from "../_shared/rest-client";
import {
  logEntryFromRaw,
  logListResultFromRaw,
  logStatsFromRaw,
  type RawLogEntry,
  type RawLogListResult,
  type RawLogStats,
} from "@/domain/logs";
import type {
  LogEntry,
  LogListResult,
  LogStats,
  LogEventType,
} from "@/domain/logs";

const PATHS = {
  logs: (imei: string) => `/v1/logs/${imei}`,
  detail: (id: string) => `/v1/logs/detail/${id}`,
  stats: (imei: string) => `/v1/logs/${imei}/stats`,
} as const;

export interface LogParams {
  eventType?: LogEventType;
  limit?: number;
  cursor?: string;
}

export interface StatsParams {
  startTime?: number;
  endTime?: number;
}

export const logs = {
  async list(imei: string, params?: LogParams): Promise<LogListResult> {
    const response = await restClient.get<RawLogListResult>(PATHS.logs(imei), {
      params: {
        event_type: params?.eventType,
        limit: params?.limit,
        cursor: params?.cursor,
      },
    });
    return logListResultFromRaw(response);
  },

  async get(id: string): Promise<LogEntry> {
    const response = await restClient.get<RawLogEntry>(PATHS.detail(id));
    return logEntryFromRaw(response);
  },

  async stats(imei: string, params?: StatsParams): Promise<LogStats> {
    const response = await restClient.get<RawLogStats>(PATHS.stats(imei), {
      params: {
        start_time: params?.startTime,
        end_time: params?.endTime,
      },
    });
    return logStatsFromRaw(response);
  },
};
