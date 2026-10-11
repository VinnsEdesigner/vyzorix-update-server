import { useCurrentOrganizationId } from '@/hooks/_shared/use-current-context';
import {
  getDeviceImeiInspect,
  useGetDeviceImeiInspect,
  getDeviceImeiTimeline,
  useGetDeviceImeiTimeline,
  getGetDeviceImeiInspectQueryKey,
  getGetDeviceImeiTimelineQueryKey,
} from '@/generated-rq/diagnostics/device-diagnostics';
import type { DeviceInspection, DomainTimelineResult as TimelineResult, TimelineEventType } from '@vyzorix/api-client';
import { fetchInspectionViaGraphQL, fetchTimelineViaGraphQL, normalizeWireInspection, normalizeWireTimeline } from './_graphql-fallback';

export interface TimelineParams {
  eventType?: TimelineEventType;
  startTime?: number;
  endTime?: number;
  cursor?: string;
  limit?: number;
}

// Aligns with the server's per-imei:orgID inspection cache (cfg.InspectionCacheTTLSeconds, 10s).
const INSPECTION_STALE_MS = 10_000;

export function useDeviceInspection(imei: string | undefined) {
  const organizationId = useCurrentOrganizationId();
  return useGetDeviceImeiInspect<DeviceInspection>(
    imei ?? '',
    {
      query: {
        queryKey: getGetDeviceImeiInspectQueryKey(imei ?? ''),
        enabled: organizationId !== null && imei !== undefined && imei !== '',
        staleTime: INSPECTION_STALE_MS,
        queryFn: async () => {
          try {
            return normalizeWireInspection(await getDeviceImeiInspect(imei!)) as unknown as Awaited<ReturnType<typeof getDeviceImeiInspect>>;
          } catch (restErr) {
            if (organizationId) {
              return fetchInspectionViaGraphQL(imei!, organizationId) as unknown as Awaited<ReturnType<typeof getDeviceImeiInspect>>;
            }
            throw restErr;
          }
        },
      },
    },
  );
}

export function useDeviceTimeline(imei: string | undefined, params?: TimelineParams) {
  const organizationId = useCurrentOrganizationId();
  const keyParams = {
    eventType: params?.eventType,
    startTime: params?.startTime,
    endTime: params?.endTime,
    cursor: params?.cursor,
    limit: params?.limit,
  };
  return useGetDeviceImeiTimeline<TimelineResult>(
    imei ?? '',
    keyParams,
    {
      query: {
        queryKey: getGetDeviceImeiTimelineQueryKey(imei ?? '', keyParams),
        enabled: organizationId !== null && imei !== undefined && imei !== '',
        queryFn: async () => {
          try {
            return normalizeWireTimeline(
              await getDeviceImeiTimeline(imei!, {
                eventType: params?.eventType,
                startTime: params?.startTime,
                endTime: params?.endTime,
                cursor: params?.cursor,
                limit: params?.limit,
              }),
              imei!,
            ) as unknown as Awaited<ReturnType<typeof getDeviceImeiTimeline>>;
          } catch (restErr) {
            if (organizationId) {
              return fetchTimelineViaGraphQL(imei!, organizationId, params) as unknown as Awaited<ReturnType<typeof getDeviceImeiTimeline>>;
            }
            throw restErr;
          }
        },
      },
    },
  );
}

export type { DeviceInspection, TimelineResult, TimelineEventType };
