import { describe, it, expect } from 'vitest';
import {
  graphqlDeviceInspectionFromRaw,
  graphqlTimelineResultFromRaw,
  type RawDeviceInspection,
  type RawTimelineConnection,
} from '@vyzorix/api-client';
import { getEventCategory, timelineEventTypeLabel } from '@vyzorix/api-client';

describe('getEventCategory', () => {
  it('maps REGISTERED and DEREGISTERED to "connection" (matches Go EventCategory)', () => {
    expect(getEventCategory('REGISTERED')).toBe('connection');
    expect(getEventCategory('DEREGISTERED')).toBe('connection');
  });

  it('maps the remaining types correctly', () => {
    expect(getEventCategory('TELEMETRY')).toBe('telemetry');
    expect(getEventCategory('COMMAND_SENT')).toBe('command');
    expect(getEventCategory('COMMAND_ACK')).toBe('command');
    expect(getEventCategory('COMMAND_FAILED')).toBe('command');
    expect(getEventCategory('CONNECTION_OPEN')).toBe('connection');
    expect(getEventCategory('CONNECTION_LOST')).toBe('connection');
    expect(getEventCategory('FCM_FALLBACK')).toBe('connection');
    expect(getEventCategory('RECONNECTED')).toBe('connection');
    expect(getEventCategory('THRESHOLD_BREACH')).toBe('error');
    expect(getEventCategory('ERROR')).toBe('error');
  });

  it('labels every event type', () => {
    const types = [
      'TELEMETRY', 'COMMAND_SENT', 'COMMAND_ACK', 'COMMAND_FAILED',
      'CONNECTION_OPEN', 'CONNECTION_LOST', 'FCM_FALLBACK', 'RECONNECTED',
      'THRESHOLD_BREACH', 'REGISTERED', 'DEREGISTERED', 'ERROR',
    ] as const;
    for (const t of types) {
      expect(timelineEventTypeLabel(t)).toBeTruthy();
    }
  });
});

describe('graphqlTimelineResultFromRaw', () => {
  it('maps events and pagination fields', () => {
    const raw: RawTimelineConnection = {
      events: [{ id: 'evt-1', type: 'TELEMETRY', timestamp: '2024-06-20T12:00:00Z', data: { riskScore: 45 } }],
      hasMore: true,
      nextCursor: 'eyJ0Ijoi...',
    };
    const result = graphqlTimelineResultFromRaw(raw);
    expect(result.events).toHaveLength(1);
    expect(result.events[0]?.id).toBe('evt-1');
    expect(result.events[0]?.type).toBe('TELEMETRY');
    expect(result.events[0]?.timestamp).toBeInstanceOf(Date);
    expect(result.hasMore).toBe(true);
    expect(result.nextCursor).toBe('eyJ0Ijoi...');
  });

  it('defaults hasMore to false when absent', () => {
    const result = graphqlTimelineResultFromRaw({ events: [], hasMore: false });
    expect(result.hasMore).toBe(false);
    expect(result.nextCursor).toBeUndefined();
  });
});

describe('graphqlDeviceInspectionFromRaw', () => {
  it('maps a full inspection (ISO timestamps -> Date)', () => {
    const raw: RawDeviceInspection = {
      identity: { imei: '861234567890123', deviceName: 'Pixel 8', model: 'Pixel 8', manufacturer: 'Google' },
      software: { osVersion: 'Android 14', appVersion: '2.1.0', securityPatch: '2024-03-01', buildId: 'UP1A' },
      registration: { status: 'registered', registeredAt: '2024-06-20T11:00:00Z', fcmTokenValid: true, commandSecretSet: true },
      connection: { webSocketStatus: 'connected', connectedAt: '2024-06-20T11:30:00Z', fcmStatus: 'valid', lastSeen: '2024-06-20T11:59:00Z', clientIp: '10.0.0.1', protocol: 'WSS' },
      telemetry: { lastTimestamp: '2024-06-20T12:00:00Z', framesToday: 4521, avgLatencyMs: 45, totalBytesToday: 15728640, sessionsToday: 3 },
    };
    const inspection = graphqlDeviceInspectionFromRaw(raw);
    expect(inspection.identity.imei).toBe('861234567890123');
    expect(inspection.registration.status).toBe('registered');
    expect(inspection.registration.registeredAt).toBeInstanceOf(Date);
    expect(inspection.connection.lastSeen).toBeInstanceOf(Date);
    expect(inspection.telemetry.framesToday).toBe(4521);
  });

  it('handles missing optional fields without throwing', () => {
    const raw: RawDeviceInspection = {
      identity: { imei: '123' },
      software: {},
      registration: { status: 'offline', fcmTokenValid: false, commandSecretSet: false },
      connection: { webSocketStatus: 'disconnected', fcmStatus: 'not_set' },
      telemetry: { framesToday: 0, sessionsToday: 0 },
    };
    const inspection = graphqlDeviceInspectionFromRaw(raw);
    expect(inspection.identity.imei).toBe('123');
    expect(inspection.registration.registeredAt).toBeUndefined();
    expect(inspection.connection.connectedAt).toBeUndefined();
  });
});
