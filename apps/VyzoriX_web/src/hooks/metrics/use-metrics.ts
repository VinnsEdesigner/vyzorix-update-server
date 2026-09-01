import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import {
  useGetDashboardDeviceImeiMetrics,
  getDashboardDeviceImeiMetricsExport,
} from '@/generated-rq/devices/device-management';
import { useGetDashboardStats } from '@/generated-rq/dashboard/dashboard-stats';
import type { TelemetryFrame } from '@vyzorix/api-client';
import { useDashboardStore, useMetricsRealtimeStore } from '@/stores';

export interface DeviceMetricsParams {
  window?: string;
}

export function useDeviceMetrics(imei: string | undefined, params?: DeviceMetricsParams) {
  const organizationId = useCurrentOrganizationId();
  return useGetDashboardDeviceImeiMetrics(
    imei ?? '',
    { window: params?.window },
    {
      query: {
        queryKey: ['metrics', organizationId ?? '', imei ?? '', params?.window] as const,
        enabled: imei !== undefined && imei !== '' && organizationId !== null,
      },
    },
  );
}

export function useDashboardStats() {
  const organizationId = useCurrentOrganizationId();
  const setStats = useDashboardStore((s) => s.setStats);
  const setRefreshing = useDashboardStore((s) => s.setRefreshing);
  const setActiveOrganization = useDashboardStore((s) => s.setActiveOrganization);

  const query = useGetDashboardStats({
    query: {
      queryKey: ['dashboard', 'stats', organizationId ?? ''] as const,
      enabled: organizationId !== null,
      refetchInterval: 30_000,
    },
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
      if (deviceId) push(deviceId,frame);
    },
  };
}

export function useExportMetrics() {
  const organizationId = useCurrentOrganizationId();
  return {
    export: (imei: string, params?: { format?: 'json' | 'csv' }) =>
      getDashboardDeviceImeiMetricsExport(imei, {
        format: params?.format,
      }),
    organizationId,
  };
}

export type { GetTelemetryResponse, DashboardStats } from '@vyzorix/api-client';
