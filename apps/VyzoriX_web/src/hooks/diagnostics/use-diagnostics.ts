import { useQuery, type UseQueryOptions } from '@tanstack/react-query';
import { getDiagnostics,
  type DeviceInspection,
  type DomainTimelineResult as TimelineResult,
  type TimelineEventType,
} from '@vyzorix/api-client';
import { queryKeys } from '@/lib/query-keys';
import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import {
  fetchInspectionViaGraphQL,
  fetchTimelineViaGraphQL,
  normalizeWireInspection,
  normalizeWireTimeline,
} from './_graphql-fallback';

export interface TimelineParams {
  eventType?: TimelineEventType;
  startTime?: number;
  endTime?: number;
  cursor?: string;
  limit?: number;
}

// Aligns with the server's per-imei:orgID inspection cache (cfg.InspectionCacheTTLSeconds, 10s).
const INSPECTION_STALE_MS = 10_000;

export function useDeviceInspection(
  imei: string | undefined,
  options?: Omit<UseQueryOptions<DeviceInspection>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.inspection(organizationId ?? '', imei ?? ''),
    queryFn: async (): Promise<DeviceInspection> => {
      try {
        return normalizeWireInspection(await getDiagnostics().getDeviceImeiInspect(imei!));
      } catch (restErr) {
        if (organizationId) {
          return fetchInspectionViaGraphQL(imei!, organizationId);
        }
        throw restErr;
      }
    },
    enabled: organizationId !== null && imei !== undefined && imei !== '',
    staleTime: INSPECTION_STALE_MS,
    ...options,
  });
}

export function useDeviceTimeline(
  imei: string | undefined,
  params?: TimelineParams,
  options?: Omit<UseQueryOptions<TimelineResult>, 'queryKey' | 'queryFn'>,
) {
  const organizationId = useCurrentOrganizationId();
  return useQuery({
    queryKey: queryKeys.timeline(organizationId ?? '', imei ?? '', { ...params }),
    queryFn: async (): Promise<TimelineResult> => {
      try {
        return normalizeWireTimeline(
          await getDiagnostics().getDeviceImeiTimeline(imei!, {
            eventType: params?.eventType,
            startTime: params?.startTime,
            endTime: params?.endTime,
            cursor: params?.cursor,
            limit: params?.limit,
          }),
          imei!,
        );
      } catch (restErr) {
        if (organizationId) {
          return fetchTimelineViaGraphQL(imei!, organizationId, params);
        }
        throw restErr;
      }
    },
    enabled: organizationId !== null && imei !== undefined && imei !== '',
    ...options,
  });
}

export type { DeviceInspection, TimelineResult, TimelineEventType };
