import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import {
  useGetDashboardDeviceImeiEvents,
  useGetDashboardEventsRecent,
  useGetDashboardEventsTypesType,
  useGetDashboardEventsId,
  getGetDashboardDeviceImeiEventsQueryKey,
  getGetDashboardEventsRecentQueryKey,
  getGetDashboardEventsTypesTypeQueryKey,
  getGetDashboardEventsIdQueryKey,
} from '@/generated-rq/devices/device-management';


export interface DeviceEventsParams {
  limit?: number;
  before?: string;
}

export function useDeviceEvents(imei: string | undefined, params?: DeviceEventsParams) {
  const keyParams = { limit: params?.limit, before: params?.before };
  return useGetDashboardDeviceImeiEvents(
    imei ?? '',
    keyParams,
    {
      query: {
        queryKey: getGetDashboardDeviceImeiEventsQueryKey(imei ?? '', keyParams),
        enabled: imei !== undefined && imei !== '',
      },
    },
  );
}

export function useRecentEvents(limit?: number) {
  const organizationId = useCurrentOrganizationId();
  return useGetDashboardEventsRecent(
    limit ? { limit } : undefined,
    {
      query: {
        queryKey: getGetDashboardEventsRecentQueryKey(limit ? { limit } : undefined),
        enabled: organizationId !== null,
      },
    },
  );
}

export function useEventsByType(type: string | undefined, params?: { limit?: number; offset?: number }) {
  const organizationId = useCurrentOrganizationId();
  const keyParams = { limit: params?.limit, offset: params?.offset };
  return useGetDashboardEventsTypesType(
    type ?? '',
    keyParams,
    {
      query: {
        queryKey: getGetDashboardEventsTypesTypeQueryKey(type ?? '', keyParams),
        enabled: type !== undefined && type !== '' && organizationId !== null,
      },
    },
  );
}

export function useEvent(id: string | undefined) {
  return useGetDashboardEventsId(
    id ?? '',
    {
      query: {
        queryKey: getGetDashboardEventsIdQueryKey(id ?? ''),
        enabled: id !== undefined && id !== '',
      },
    },
  );
}

export type { DeviceEvent, DeviceEventListResult } from '@vyzorix/api-client';
