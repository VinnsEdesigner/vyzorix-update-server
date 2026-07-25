import { restClient, getOrganizationContext } from "../_shared/rest-client";
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
  logs: (imei: string) => `/v1/dashboard/device/${imei}/logs`,
  detail: (id: string) => `/v1/dashboard/logs/detail/${id}`,
  stats: (imei: string) => `/v1/dashboard/device/${imei}/logs/stats`,
} as const;

export interface LogParams {
  eventType?: LogEventType;
  limit?: number;
  cursor?: string;
  organizationId?: string;
}

export interface StatsParams {
  startTime?: number;
  endTime?: number;
  organizationId?: string;
}

export const logs = {
  async list(imei: string, params?: LogParams): Promise<LogListResult> {
    const response = await restClient.get<RawLogListResult>(PATHS.logs(imei), {
      params: {
        event_type: params?.eventType,
        limit: params?.limit,
        cursor: params?.cursor,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return logListResultFromRaw(response);
  },

  async get(id: string, organizationId?: string): Promise<LogEntry> {
    const response = await restClient.get<RawLogEntry>(PATHS.detail(id), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return logEntryFromRaw(response);
  },

  async stats(imei: string, params?: StatsParams): Promise<LogStats> {
    const response = await restClient.get<RawLogStats>(PATHS.stats(imei), {
      params: {
        start_time: params?.startTime,
        end_time: params?.endTime,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return logStatsFromRaw(response);
  },
};
