import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import { getDashboardDeviceImeiLogs, useGetDashboardDeviceImeiLogs } from '@/generated-rq/devices/device-management';
import type { LogListResult } from '@vyzorix/api-client';
import { fetchDeviceLogsViaGraphQL, normalizeDeviceLogList } from './_graphql-fallback';

export interface LogParams {
  level?: string;
  limit?: number;
  before?: string;
}

export function useDeviceLogs(imei: string | undefined, params?: LogParams) {
  const organizationId = useCurrentOrganizationId();
  return useGetDashboardDeviceImeiLogs<LogListResult>(
    imei ?? '',
    { limit: params?.limit, before: params?.before, level: params?.level },
    {
      query: {
        queryKey: ['logs', imei, { ...params, organizationId }] as const,
        enabled: imei !== undefined && imei !== '' && organizationId !== null,
        queryFn: async () => {
          try {
            const result = await getDashboardDeviceImeiLogs(
              imei!,
              { limit: params?.limit, before: params?.before, level: params?.level },
            );
            return normalizeDeviceLogList(result, imei!) as unknown as Awaited<ReturnType<typeof getDashboardDeviceImeiLogs>>;
          } catch (restError) {
            if (!organizationId || !imei) throw restError;
            return fetchDeviceLogsViaGraphQL(organizationId, imei!, params) as unknown as Awaited<ReturnType<typeof getDashboardDeviceImeiLogs>>;
          }
        },
      },
    },
  );
}

export type { LogListResult, LogEntry, LogEventType } from '@vyzorix/api-client';
