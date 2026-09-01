import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import {
  useGetDashboardDeviceImeiEvents,
  useGetDashboardEventsRecent,
  useGetDashboardEventsTypesType,
  useGetDashboardEventsId,
} from '@/generated-rq/devices/device-management';


export interface DeviceEventsParams {
  limit?: number;
  before?: string;
}

export function useDeviceEvents(imei: string | undefined, params?: DeviceEventsParams) {
  const organizationId = useCurrentOrganizationId();
  return useGetDashboardDeviceImeiEvents(
    imei ?? '',
    { limit: params?.limit, before: params?.before },
    {
      query: {
        queryKey: ['device-events', imei, { ...params, organizationId }] as const,
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
        queryKey: ['events', 'recent', limit] as const,
        enabled: organizationId !== null,
      },
    },
  );
}

export function useEventsByType(type: string | undefined, params?: { limit?: number; offset?: number }) {
  const organizationId = useCurrentOrganizationId();
  return useGetDashboardEventsTypesType(
    type ?? '',
    { limit: params?.limit, offset: params?.offset },
    {
      query: {
        queryKey: ['events', 'type', type, { ...params, organizationId }] as const,
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
        queryKey: ['events', 'entry', id ?? ''] as const,
        enabled: id !== undefined && id !== '',
      },
    },
  );
}

export type { DeviceEvent, DeviceEventListResult } from '@vyzorix/api-client';
