import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  fetchDeviceMetrics,
  fetchDashboardStats,
  exportMetrics,
  type DeviceMetrics,
  type DashboardStats,
  type TimeRange,
  type MetricResolution,
  type TelemetryFrame,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { useDashboardStore, useMetricsRealtimeStore } from '@/stores';

export interface DeviceMetricsParams {
  range?: TimeRange;
  startTime?: number;
  endTime?: number;
  resolution?: MetricResolution;
}

export function useDeviceMetrics(
  imei: string | undefined,
  params?: DeviceMetricsParams,
  options?: Omit<UseQueryOptions<DeviceMetrics>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceMetrics(organizationId ?? '', imei ?? '', params?.range),
    queryFn: () =>
      fetchDeviceMetrics(imei!, {
        range: params?.range,
        startTime: params?.startTime,
        endTime: params?.endTime,
        resolution: params?.resolution,
        organizationId: organizationId ?? undefined,
      }),
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export function useDashboardStats(
  options?: Omit<UseQueryOptions<DashboardStats>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  const setStats = useDashboardStore((s) => s.setStats);
  const setRefreshing = useDashboardStore((s) => s.setRefreshing);
  const setActiveOrganization = useDashboardStore((s) => s.setActiveOrganization);

  const query = useQuery({
    queryKey: queryKeys.dashboardStats(organizationId ?? ''),
    queryFn: () => fetchDashboardStats(organizationId ?? undefined),
    enabled: organizationId !== null,
    refetchInterval: 30_000,
    ...options,
  });

  setActiveOrganization(organizationId);
  setRefreshing(query.isFetching);
  if (query.data) setStats(query.data);

  return query;
}

export function useLiveMetrics(deviceId: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  const push = useMetricsRealtimeStore((s) => s.push);
  const setActiveOrganization = useMetricsRealtimeStore((s) => s.setActiveOrganization);
  const series = useMetricsRealtimeStore((s) =>
    deviceId ? s.byDevice[deviceId] ?? [] : [],
  );

  setActiveOrganization(organizationId);

  return {
    series,
    push: (frame: TelemetryFrame) => {
      if (deviceId) push(deviceId, frame);
    },
  };
}

export function useExportMetrics() {
  const organizationId = useCurrentOrganizationId();
  return {
    export: (imei: string, params: {
      format?: 'json' | 'csv';
      range?: TimeRange;
      metrics?: string[];
    }) =>
      exportMetrics(imei, {
        format: params.format,
        range: params.range,
        metrics: params.metrics,
        organizationId: organizationId ?? undefined,
      }),
  };
}

export type { DeviceMetrics, DashboardStats, TimeRange, MetricResolution };
