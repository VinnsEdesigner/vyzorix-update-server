import { restClient, getOrganizationContext } from "../_shared/rest-client";
import type { DeviceMetrics, DashboardStats, TimeRange, MetricResolution } from "../../../domain/metrics";
import { deviceMetricsFromRaw, dashboardStatsFromRaw } from "../../../domain/metrics";
import { getMetricsConfig } from "../../config";

export const METRICS_PATHS = {
  metrics: (imei: string) => `/v1/dashboard/device/${imei}/metrics`,
  metricsExport: (imei: string) => `/v1/dashboard/device/${imei}/metrics/export`,
  telemetry: (imei: string) => `/v1/dashboard/device/${imei}/telemetry`,
  dashboardStats: "/v1/dashboard/stats",
} as const;

export async function fetchDeviceMetrics(
  imei: string,
  params?: {
    range?: TimeRange;
    startTime?: number;
    endTime?: number;
    resolution?: MetricResolution;
    organizationId?: string;
  }
): Promise<DeviceMetrics> {
  const data = await restClient.get<Parameters<typeof deviceMetricsFromRaw>[0]>(
    METRICS_PATHS.metrics(imei),
    {
      params: {
        range: params?.range,
        start_time: params?.startTime,
        end_time: params?.endTime,
        resolution: params?.resolution,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
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
    organizationId?: string;
  }
): Promise<Blob> {
  const searchParams = new URLSearchParams();
  const orgId = params?.organizationId || getOrganizationContext();
  if (orgId) searchParams.set("organization_id", orgId);
  if (params?.format) searchParams.set("format", params.format);
  if (params?.range) searchParams.set("range", params.range);
  if (params?.metrics) searchParams.set("metrics", params.metrics.join(","));

  const response = await restClient.get<Blob>(
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
    organizationId?: string;
  }
): Promise<unknown> {
  const metricsConfig = getMetricsConfig();
  const defaultLimit = params?.limit ?? metricsConfig.defaultLimit;
  const orgId = params?.organizationId || getOrganizationContext();
  
  const data = await restClient.get<{
    frames: {
      timestamp: number;
      risk_score: number;
      thermal_temp: number;
      buffer_level: number;
      uptime: number;
    }[];
    stats: {
      risk_score: { current: number; avg: number; min: number; max: number };
      thermal_temp: { current: number; avg: number; min: number; max: number };
      buffer_level: { current: number; avg: number; min: number; max: number };
    };
  }>(METRICS_PATHS.telemetry(imei), {
    params: {
      start_time: params?.startTime,
      end_time: params?.endTime,
      limit: defaultLimit,
      organization_id: orgId,
    },
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

export async function fetchDashboardStats(organizationId?: string): Promise<DashboardStats> {
  const orgId = organizationId || getOrganizationContext();
  const data = await restClient.get<Parameters<typeof dashboardStatsFromRaw>[0]>(
    METRICS_PATHS.dashboardStats,
    { params: { organization_id: orgId } }
  );
  return dashboardStatsFromRaw(data);
}
