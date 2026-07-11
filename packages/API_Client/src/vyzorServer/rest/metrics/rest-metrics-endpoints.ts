import { restClient } from "../_shared/rest-client";
import type { DeviceMetrics, DashboardStats, TimeRange, MetricResolution } from "@/domain/metrics";
import { deviceMetricsFromRaw, dashboardStatsFromRaw } from "@/domain/metrics";
import { getMetricsConfig } from "../../config";

export const METRICS_PATHS = {
  metrics: (imei: string) => `/v1/device/${imei}/metrics`,
  metricsExport: (imei: string) => `/v1/device/${imei}/metrics/export`,
  telemetry: (imei: string) => `/v1/device/${imei}/telemetry`,
  dashboardStats: "/v1/dashboard/stats",
} as const;

export async function fetchDeviceMetrics(
  imei: string,
  params?: {
    range?: TimeRange;
    startTime?: number;
    endTime?: number;
    resolution?: MetricResolution;
  }
): Promise<DeviceMetrics> {
  const data = await restClient.get<Parameters<typeof deviceMetricsFromRaw>[0]>(
    METRICS_PATHS.metrics(imei),
    {
      range: params?.range,
      start_time: params?.startTime,
      end_time: params?.endTime,
      resolution: params?.resolution,
    }
  );

  return deviceMetricsFromRaw(data);
}

export async function exportMetrics(
  imei: string,
  params?: {
    format?: "json" | "csv";
    range?: TimeRange;
    metrics?: string[];
  }
): Promise<Blob> {
  const searchParams = new URLSearchParams();
  if (params?.format) searchParams.set("format", params.format);
  if (params?.range) searchParams.set("range", params.range);
  if (params?.metrics) searchParams.set("metrics", params.metrics.join(","));

  const response = await restClient.get(
    `${METRICS_PATHS.metricsExport(imei)}?${searchParams}`,
    { responseType: 'blob' }
  );

  return response;
}

export async function fetchTelemetryHistory(
  imei: string,
  params?: {
    startTime?: number;
    endTime?: number;
    limit?: number;
  }
) {
  const metricsConfig = getMetricsConfig();
  const defaultLimit = params?.limit ?? metricsConfig.defaultLimit;
  
  const data = await restClient.get<{
    frames: Array<{
      timestamp: number;
      risk_score: number;
      thermal_temp: number;
      buffer_level: number;
      uptime: number;
    }>;
    stats: {
      risk_score: { current: number; avg: number; min: number; max: number };
      thermal_temp: { current: number; avg: number; min: number; max: number };
      buffer_level: { current: number; avg: number; min: number; max: number };
    };
  }>(METRICS_PATHS.telemetry(imei), {
    start_time: params?.startTime,
    end_time: params?.endTime,
    limit: defaultLimit,
  });

  return {
    frames: data.frames.map((frame) => ({
      timestamp: new Date(frame.timestamp),
      riskScore: frame.risk_score,
      thermalTemp: frame.thermal_temp,
      bufferLevel: frame.buffer_level,
      uptime: frame.uptime,
    })),
    stats: {
      riskScore: data.stats.risk_score,
      thermalTemp: data.stats.thermal_temp,
      bufferLevel: data.stats.buffer_level,
    },
  };
}

export async function fetchDashboardStats(): Promise<DashboardStats> {
  const data = await restClient.get<Parameters<typeof dashboardStatsFromRaw>[0]>(
    METRICS_PATHS.dashboardStats
  );
  return dashboardStatsFromRaw(data);
}
