import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import {
  getDevices,
  type DeviceEvent,
  type DeviceEventListResult,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export interface DeviceEventsParams {
  limit?: number;
  before?: string;
}

export function useDeviceEvents(
  imei: string | undefined,
  params?: DeviceEventsParams,
  options?: Omit<UseQueryOptions<DeviceEventListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceEvents(imei ?? '', { ...params, organizationId }),
    queryFn: () =>
      getDevices().getDashboardDeviceImeiEvents(imei!, {
        limit: params?.limit,
        before: params?.before,
      }),
    enabled: imei !== undefined && imei !== '',
    ...options,
  });
}

export function useRecentEvents(
  limit?: number,
  options?: Omit<UseQueryOptions<DeviceEventListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.recentEvents(limit),
    queryFn: () => getDevices().getDashboardEventsRecent(limit ? { limit } : undefined),
    enabled: organizationId !== null,
    ...options,
  });
}

export function useEventsByType(
  type: string | undefined,
  params?: { limit?: number; offset?: number },
  options?: Omit<UseQueryOptions<DeviceEventListResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.events({ type, ...params, organizationId }),
    queryFn: () =>
      getDevices().getDashboardEventsTypesType(type!, {
        limit: params?.limit,
        offset: params?.offset,
      }),
    enabled: type !== undefined && type !== '' && organizationId !== null,
    ...options,
  });
}

export function useEvent(
  id: string | undefined,
  options?: Omit<UseQueryOptions<DeviceEvent>, 'queryKey' | 'queryFn'>,
) {
  return useQuery({
    queryKey: ['events', 'entry', id ?? ''],
    queryFn: () => getDevices().getDashboardEventsId(id!),
    enabled: id !== undefined && id !== '',
    ...options,
  });
}

export type { DeviceEvent, DeviceEventListResult };
