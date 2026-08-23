import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  getDevices,
  type LogEntry,
  type LogListResult,
  type LogEventType,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { fetchDeviceLogsViaGraphQL, normalizeDeviceLogList } from './_graphql-fallback';

export interface LogParams {
  level?: string;
  limit?: number;
  before?: string;
}

export function useDeviceLogs(
  imei: string | undefined,
  params?: LogParams,
  options?: Omit<UseQueryOptions<LogListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.logs(imei ?? '', { ...params, organizationId }),
    queryFn: async (): Promise<LogListResult> => {
      try {
        const result = await getDevices().getDashboardDeviceImeiLogs(imei!, {
          limit: params?.limit,
          before: params?.before,
          level: params?.level,
        });
        return normalizeDeviceLogList(result, imei!);
      } catch (restError) {
        if (!organizationId || !imei) throw restError;
        return fetchDeviceLogsViaGraphQL(organizationId, imei, params);
      }
    },
    enabled: imei !== undefined && imei !== '' && organizationId !== null,
    ...options,
  });
}

export type { LogListResult, LogEntry, LogEventType };
