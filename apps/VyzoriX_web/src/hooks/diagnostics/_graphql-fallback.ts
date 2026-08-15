import {
  queryDeviceInspection,
  queryDeviceTimeline,
  graphqlDeviceInspectionFromRaw,
  graphqlTimelineResultFromRaw,
  type RawGraphQLDeviceInspection,
  type RawGraphQLTimelineConnection,
  type DeviceInspection,
  type TimelineResult,
} from '@vyzorix/api-client';
import type { TimelineParams } from './use-diagnostics';

interface ApolloQueryResponse<T> {
  data?: { deviceInspection?: T };
}

interface ApolloTimelineResponse<T> {
  data?: { deviceTimeline?: T };
}

function extractInspection(response: unknown): RawGraphQLDeviceInspection {
  const r = response as ApolloQueryResponse<RawGraphQLDeviceInspection> | null;
  const raw = r?.data?.deviceInspection;
  if (!raw) throw new Error('GraphQL deviceInspection returned no data');
  return raw;
}

function extractTimeline(response: unknown): RawGraphQLTimelineConnection {
  const r = response as ApolloTimelineResponse<RawGraphQLTimelineConnection> | null;
  const raw = r?.data?.deviceTimeline;
  if (!raw) throw new Error('GraphQL deviceTimeline returned no data');
  return raw;
}

export async function fetchInspectionViaGraphQL(
  imei: string,
  organizationId: string,
): Promise<DeviceInspection> {
  const response = await queryDeviceInspection({ imei, organizationId });
  return graphqlDeviceInspectionFromRaw(extractInspection(response));
}

export async function fetchTimelineViaGraphQL(
  imei: string,
  organizationId: string,
  params?: TimelineParams,
): Promise<TimelineResult> {
  const response = await queryDeviceTimeline({
    imei,
    organizationId,
    eventType: params?.eventType,
    startTime: params?.startTime,
    endTime: params?.endTime,
    limit: params?.limit,
    cursor: params?.cursor,
  });
  const connection = extractTimeline(response);
  const result = graphqlTimelineResultFromRaw(connection);
  // GraphQL TimelineEvent does not expose deviceId; inject the known imei.
  result.events = result.events.map((e) => ({ ...e, deviceId: imei }));
  return result;
}
