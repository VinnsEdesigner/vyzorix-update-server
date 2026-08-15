import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  queryTelemetryHistory,
  getLatestTelemetry,
  getTelemetryStats,
  exportTelemetry,
  type TelemetryHistoryParams,
  type TelemetryHistoryResponse,
  type LatestTelemetry,
  type TelemetryStats,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useTelemetryHistory(
  deviceId: string | undefined,
  params?: Omit<TelemetryHistoryParams, 'deviceId'>,
  options?: Omit<UseQueryOptions<TelemetryHistoryResponse>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.telemetryHistory(deviceId ?? '', { ...params, organizationId }),
    queryFn: () =>
      queryTelemetryHistory({
        deviceId: deviceId!,
        startTime: params?.startTime,
        endTime: params?.endTime,
        limit: params?.limit,
        format: params?.format,
        organizationId: organizationId ?? undefined,
      }),
    enabled: deviceId !== undefined && deviceId !== '' && organizationId !== null,
    ...options,
  });
}

export function useLatestTelemetry(
  deviceId: string | undefined,
  options?: Omit<UseQueryOptions<LatestTelemetry>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.latestTelemetry(organizationId ?? '', deviceId ?? ''),
    queryFn: () => getLatestTelemetry(deviceId!, organizationId ?? undefined),
    enabled: deviceId !== undefined && deviceId !== '' && organizationId !== null,
    refetchInterval: 10_000,
    ...options,
  });
}

export function useTelemetryStats(
  deviceId: string | undefined,
  options?: Omit<UseQueryOptions<TelemetryStats>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.telemetryStats(organizationId ?? '', deviceId ?? ''),
    queryFn: () => getTelemetryStats(deviceId!, organizationId ?? undefined),
    enabled: deviceId !== undefined && deviceId !== '' && organizationId !== null,
    ...options,
  });
}

export function useExportTelemetry() {
  const organizationId = useCurrentOrganizationId();
  return {
    export: (params: Omit<TelemetryHistoryParams, 'format'> & { format?: 'json' | 'csv' }) =>
      exportTelemetry({ ...params, organizationId: organizationId ?? undefined }),
  };
}

export type {
  TelemetryHistoryParams,
  TelemetryHistoryResponse,
  LatestTelemetry,
  TelemetryStats,
};
