import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  logs,
  type LogEntry,
  type LogListResult,
  type LogStats,
  type LogParams,
  type StatsParams,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchDeviceLogsViaGraphQL } from './_graphql-fallback';

export function useDeviceLogs(
  imei: string | undefined,
  params?: Omit<LogParams, 'organizationId'>,
  options?: Omit<UseQueryOptions<LogListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.logs(imei ?? '', { ...params, organizationId }),
    queryFn: async () => {
      try {
        return await logs.list(imei!, { ...params, organizationId: organizationId ?? undefined });
      } catch (restError) {
        if (!organizationId || !imei) throw restError;
        return fetchDeviceLogsViaGraphQL(organizationId, imei, params);
      }
    },
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export function useLog(
  id: string | undefined,
  options?: Omit<UseQueryOptions<LogEntry>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: ['logs', 'entry', id ?? ''],
    queryFn: () => logs.get(id!, organizationId ?? undefined),
    enabled: id !== undefined && id !== '' && organizationId !== null,
    ...options,
  });
}

export function useLogStats(
  imei: string | undefined,
  params?: Omit<StatsParams, 'organizationId'>,
  options?: Omit<UseQueryOptions<LogStats>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.logStats(imei ?? '', { ...params, organizationId }),
    queryFn: () => logs.stats(imei!, { ...params, organizationId: organizationId ?? undefined }),
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export type { LogListResult, LogEntry, LogStats, LogParams };
