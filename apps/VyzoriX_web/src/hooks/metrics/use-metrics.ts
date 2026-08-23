import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  getDevices,
  getDashboard,
  type GetTelemetryResponse,
  type DashboardStats,
  type TelemetryFrame,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { useDashboardStore, useMetricsRealtimeStore } from '@/stores';

export interface DeviceMetricsParams {
  window?: string;
}

export function useDeviceMetrics(
  imei: string | undefined,
  params?: DeviceMetricsParams,
  options?: Omit<UseQueryOptions<GetTelemetryResponse>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceMetrics(organizationId ?? '', imei ?? '', params?.window),
    queryFn: () =>
      getDevices().getDashboardDeviceImeiMetrics(imei!, {
        window: params?.window,
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
    queryFn: () => getDashboard().getDashboardStats(),
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
    export: (imei: string, params?: { format?: 'json' | 'csv' }) =>
      getDevices().getDashboardDeviceImeiMetricsExport(imei, {
        format: params?.format,
      }),
    organizationId,
  };
}

export type { GetTelemetryResponse, DashboardStats, DeviceMetricsParams as MetricsParams };
