import {
  queryDeviceInspection,
  queryDeviceTimeline,
  graphqlDeviceInspectionFromRaw,
  graphqlTimelineResultFromRaw,
  type RawGraphQLDeviceInspection,
  type RawGraphQLTimelineConnection,
  type DeviceInspection,
  type TimelineResult,
  type TimelineEvent,
  type TimelineEventType,
  type DeviceInspectionResult,
  type DeviceStatus,
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

function toDate(value?: number | string | null): Date | undefined {
  if (value === undefined || value === null) return undefined;
  return new Date(value);
}

/** Maps the REST wire DTO (DeviceInspectionResult, epoch-millis numbers) onto
 * the domain DeviceInspection shape so hook consumers get one type. */
export function normalizeWireInspection(raw: DeviceInspectionResult): DeviceInspection {
  return {
    identity: {
      imei: raw.identity?.imei ?? '',
      deviceName: raw.identity?.deviceName,
      model: raw.identity?.model,
      manufacturer: raw.identity?.manufacturer,
    },
    software: {
      osVersion: raw.software?.osVersion,
      appVersion: raw.software?.appVersion,
      securityPatch: raw.software?.securityPatch,
      buildId: raw.software?.buildId,
    },
    registration: {
      status: (raw.registration?.status ?? 'offline') as DeviceStatus,
      registeredAt: toDate(raw.registration?.registeredAt),
      fcmTokenValid: raw.registration?.fcmTokenValid ?? false,
      fcmTokenRefreshedAt: toDate(raw.registration?.fcmTokenRefreshedAt),
      commandSecretSet: raw.registration?.commandSecretSet ?? false,
    },
    connection: {
      webSocketStatus: raw.connection?.webSocketStatus === 'connected' ? 'connected' : 'disconnected',
      connectedAt: toDate(raw.connection?.connectedAt),
      fcmStatus: (raw.connection?.fcmStatus ?? 'not_set') as DeviceInspection['connection']['fcmStatus'],
      lastSeen: toDate(raw.connection?.lastSeen),
      clientIp: raw.connection?.clientIp,
      protocol: raw.connection?.protocol,
    },
    telemetry: {
      lastTimestamp: toDate(raw.telemetry?.lastTimestamp),
      framesToday: raw.telemetry?.framesToday ?? 0,
      avgLatencyMs: raw.telemetry?.avgLatencyMs,
      totalBytesToday: raw.telemetry?.totalBytesToday,
      sessionsToday: raw.telemetry?.sessionsToday ?? 0,
    },
  };
}

/** Maps the REST wire timeline (TimelineEventResult[], ISO timestamps) onto
 * the domain TimelineResult (Date timestamps). */
export function normalizeWireTimeline(
  raw: { events?: { id?: string; deviceId?: string; type?: string; timestamp?: string; data?: Record<string, unknown> }[]; hasMore?: boolean; nextCursor?: string },
  imei: string,
): TimelineResult {
  const events: TimelineEvent[] = (raw.events ?? []).map((e) => ({
    id: e.id ?? '',
    deviceId: e.deviceId ?? imei,
    type: (e.type ?? 'ERROR') as TimelineEventType,
    timestamp: toDate(e.timestamp) ?? new Date(),
    data: e.data ?? {},
  }));
  return { events, hasMore: raw.hasMore ?? false, nextCursor: raw.nextCursor };
}
