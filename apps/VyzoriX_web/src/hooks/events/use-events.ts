import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { events, type Event, type EventResult, type EventParams } from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';

export function useDeviceEvents(
  imei: string | undefined,
  params?: Omit<EventParams, 'organizationId'>,
  options?: Omit<UseQueryOptions<EventResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.deviceEvents(imei ?? '', { ...params, organizationId }),
    queryFn: () =>
      events.getDeviceEvents(imei!, { ...params, organizationId: organizationId ?? undefined }),
    enabled: imei !== undefined && imei !== '',
    ...options,
  });
}

export function useRecentEvents(
  limit?: number,
  options?: Omit<UseQueryOptions<Event[]>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.recentEvents(limit),
    queryFn: () => events.getRecentEvents(limit, organizationId ?? undefined),
    enabled: organizationId !== null,
    ...options,
  });
}

export function useEventsByType(
  type: string | undefined,
  params?: Omit<EventParams, 'types'>,
  options?: Omit<UseQueryOptions<EventResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.events({ type, ...params, organizationId }),
    queryFn: () =>
      events.getEventsByType(type!, {
        ...params,
        organizationId: organizationId ?? undefined,
      } as Omit<EventParams, 'types'> & { organizationId?: string }),
    enabled: type !== undefined && type !== '' && organizationId !== null,
    ...options,
  });
}

export function useEvent(
  id: string | undefined,
  options?: Omit<UseQueryOptions<Event>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: ['events', 'entry', id ?? ''],
    queryFn: () => events.getById(id!, organizationId ?? undefined),
    enabled: id !== undefined && id !== '',
    ...options,
  });
}

export type { Event, EventResult, EventParams };
