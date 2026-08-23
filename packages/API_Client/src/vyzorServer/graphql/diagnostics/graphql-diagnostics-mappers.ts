import type {
  DeviceInspection,
  TimelineResult,
  TimelineEvent,
  TimelineEventType,
  FCMStatus,
  WebSocketStatus,
} from '../../../domain/diagnostics';
import type { DeviceStatus } from '../../../domain/_shared';
import type {
  RawDeviceInspection,
  RawTimelineConnection,
} from './graphql-diagnostics-types';

/** Aliases kept for the web hooks' GraphQL fallback imports. */
export type RawGraphQLDeviceInspection = RawDeviceInspection;
export type RawGraphQLTimelineConnection = RawTimelineConnection;

function parseTimestamp(value?: string | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value);
}

export function graphqlDeviceInspectionFromRaw(raw: RawDeviceInspection): DeviceInspection {
  return {
    identity: {
      imei: raw.identity.imei,
      deviceName: raw.identity.deviceName,
      model: raw.identity.model,
      manufacturer: raw.identity.manufacturer,
    },
    software: {
      osVersion: raw.software.osVersion,
      appVersion: raw.software.appVersion,
      securityPatch: raw.software.securityPatch,
      buildId: raw.software.buildId,
    },
    registration: {
      status: raw.registration.status as DeviceStatus,
      registeredAt: parseTimestamp(raw.registration.registeredAt),
      fcmTokenValid: raw.registration.fcmTokenValid,
      fcmTokenRefreshedAt: parseTimestamp(raw.registration.fcmTokenRefreshedAt),
      commandSecretSet: raw.registration.commandSecretSet,
    },
    connection: {
      webSocketStatus: (raw.connection.webSocketStatus ?? 'disconnected') as WebSocketStatus,
      connectedAt: parseTimestamp(raw.connection.connectedAt),
      fcmStatus: (raw.connection.fcmStatus ?? 'not_set') as FCMStatus,
      lastSeen: parseTimestamp(raw.connection.lastSeen),
      clientIp: raw.connection.clientIp,
      protocol: raw.connection.protocol,
    },
    telemetry: {
      lastTimestamp: parseTimestamp(raw.telemetry.lastTimestamp),
      framesToday: raw.telemetry.framesToday,
      avgLatencyMs: raw.telemetry.avgLatencyMs,
      totalBytesToday: raw.telemetry.totalBytesToday,
      sessionsToday: raw.telemetry.sessionsToday,
    },
  };
}

export function graphqlTimelineResultFromRaw(raw: RawTimelineConnection): TimelineResult {
  const events: TimelineEvent[] = (raw.events ?? []).map((e) => ({
    id: e.id,
    deviceId: '',
    type: e.type as TimelineEventType,
    timestamp: parseTimestamp(e.timestamp) ?? new Date(),
    data: e.data ?? {},
  }));
  return {
    events,
    hasMore: raw.hasMore,
    nextCursor: raw.nextCursor,
  };
}
