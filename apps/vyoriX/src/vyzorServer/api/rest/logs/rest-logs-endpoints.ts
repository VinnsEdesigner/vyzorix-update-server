import { apiGet } from "../_shared/rest-client";

export const LOGS_PATHS = {
  logs: (imei: string) => `/v1/logs/${imei}`,
  log: (id: string) => `/v1/logs/detail/${id}`,
  stats: (imei: string) => `/v1/logs/${imei}/stats`,
} as const;

export type RawLogEntry = {
  id: string;
  timestamp: number;
  type: string;
  data: string;
  device_imei: string;
};

export type RawLogConnection = {
  logs: RawLogEntry[];
  pagination: {
    limit: number;
    has_more: boolean;
    next_cursor?: string;
  };
};

export type RawLogStats = {
  total: number;
  by_type: {
    connection: number;
    command: number;
    telemetry: number;
    error: number;
    warning: number;
  };
};

export async function fetchLogs(
  imei: string,
  params?: { type?: string; limit?: number; cursor?: string }
) {
  const data = await apiGet<RawLogConnection>(LOGS_PATHS.logs(imei), {
    type: params?.type,
    limit: params?.limit,
    cursor: params?.cursor,
  });
  
  return {
    logs: data.logs.map(logFromRaw),
    pagination: {
      limit: data.pagination.limit,
      hasMore: data.pagination.has_more,
      nextCursor: data.pagination.next_cursor,
    },
  };
}

export async function fetchLog(id: string) {
  const data = await apiGet<RawLogEntry>(LOGS_PATHS.log(id));
  return logFromRaw(data);
}

export async function fetchLogStats(imei: string, params?: { start_time?: number; end_time?: number }) {
  const data = await apiGet<RawLogStats>(LOGS_PATHS.stats(imei), {
    start_time: params?.start_time,
    end_time: params?.end_time,
  });
  
  return {
    total: data.total,
    byType: {
      connection: data.by_type.connection,
      command: data.by_type.command,
      telemetry: data.by_type.telemetry,
      error: data.by_type.error,
      warning: data.by_type.warning,
    },
  };
}

function logFromRaw(raw: RawLogEntry) {
  return {
    id: raw.id,
    timestamp: new Date(raw.timestamp > 1e12 ? raw.timestamp : raw.timestamp * 1000),
    type: raw.type,
    data: raw.data,
    deviceImei: raw.device_imei,
  };
}
