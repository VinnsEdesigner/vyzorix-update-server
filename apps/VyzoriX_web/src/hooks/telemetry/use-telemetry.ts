import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  getTelemetry,
  type TelemetryHistoryQueryResult,
  type TelemetryEntry,
  type TelemetryStatsResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface TelemetryHistoryParams {
  startTime?: number;
  endTime?: number;
  limit?: number;
  format?: string;
}

export function useTelemetryHistory(
  deviceId: string | undefined,
  params?: TelemetryHistoryParams,
  options?: Omit<UseQueryOptions<TelemetryHistoryQueryResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.telemetryHistory(deviceId ?? '', { ...params, organizationId }),
    queryFn: () =>
      getTelemetry().getTelemetryHistory({
        deviceId: deviceId!,
        startTime: params?.startTime,
        endTime: params?.endTime,
        limit: params?.limit,
        format: params?.format,
      }),
    enabled: deviceId !== undefined && deviceId !== '' && organizationId !== null,
    ...options,
  });
}

export function useLatestTelemetry(
  deviceId: string | undefined,
  options?: Omit<UseQueryOptions<TelemetryEntry>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.latestTelemetry(organizationId ?? '', deviceId ?? ''),
    queryFn: () => getTelemetry().getTelemetryLatestDeviceId(deviceId!),
    enabled: deviceId !== undefined && deviceId !== '' && organizationId !== null,
    refetchInterval: 10_000,
    ...options,
  });
}

export function useTelemetryStats(
  deviceId: string | undefined,
  options?: Omit<UseQueryOptions<TelemetryStatsResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.telemetryStats(organizationId ?? '', deviceId ?? ''),
    queryFn: () => getTelemetry().getTelemetryStatsDeviceId(deviceId!),
    enabled: deviceId !== undefined && deviceId !== '' && organizationId !== null,
    ...options,
  });
}

export function useExportTelemetry() {
  return {
    export: (params: TelemetryHistoryParams & { deviceId: string; format?: 'json' | 'csv' }) =>
      getTelemetry().getTelemetryHistoryExport({
        deviceId: params.deviceId,
        startTime: params.startTime,
        endTime: params.endTime,
        limit: params.limit,
      }),
  };
}

export type {
  TelemetryHistoryParams as TelemetryHistoryQueryParams,
  TelemetryHistoryQueryResult,
  TelemetryEntry,
  TelemetryStatsResult,
};
