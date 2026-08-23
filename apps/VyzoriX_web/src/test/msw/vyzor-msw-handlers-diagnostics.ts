import { http, HttpResponse, delay } from 'msw';
import type {
  DeviceInspectionResult,
  WireTimelineResult,
  WireTimelineEventResult,
} from '@vyzorix/api-client';

const API_BASE = '/v1/device';

const now = Date.now();

const inspection: DeviceInspectionResult = {
  identity: { imei: '123', deviceName: 'Test Device', model: 'Model X', manufacturer: 'Acme' },
  software: { osVersion: '14.0', appVersion: 'v1.1.0', securityPatch: '2025-01-01', buildId: 'build-1' },
  registration: {
    status: 'registered',
    registeredAt: now - 86_400_000,
    fcmTokenValid: true,
    fcmTokenRefreshedAt: now - 3_600_000,
    commandSecretSet: true,
  },
  connection: {
    webSocketStatus: 'connected',
    connectedAt: now - 3_600_000,
    fcmStatus: 'valid',
    lastSeen: now - 60_000,
    clientIp: '10.0.0.1',
    protocol: 'ws',
  },
  telemetry: {
    lastTimestamp: now - 60_000,
    framesToday: 5,
    avgLatencyMs: 42,
    totalBytesToday: 1024,
    sessionsToday: 1,
  },
};

const timelineEvents: WireTimelineEventResult[] = [
  {
    id: 'evt-1',
    deviceId: '123',
    type: 'TELEMETRY',
    timestamp: new Date(now - 60_000).toISOString(),
    data: { riskScore: 5, bufferLevel: 42 },
  },
  {
    id: 'evt-2',
    deviceId: '123',
    type: 'CONNECTION_OPEN',
    timestamp: new Date(now - 3_600_000).toISOString(),
    data: { ip: '10.0.0.1' },
  },
];

export function createDiagnosticsHandlers() {
  return [
    // GET /v1/device/:imei/inspect — device inspection
    http.get(`${API_BASE}/:imei/inspect`, async ({ params }) => {
      await delay(30);
      const imei = params.imei as string;
      return HttpResponse.json({
        ...inspection,
        identity: { ...inspection.identity, imei },
      });
    }),

    // GET /v1/device/:imei/timeline — timeline
    http.get(`${API_BASE}/:imei/timeline`, async ({ params }) => {
      await delay(30);
      const imei = params.imei as string;
      const events = timelineEvents.map((e) => ({ ...e, deviceId: imei }));
      const result: WireTimelineResult = {
        events,
        hasMore: false,
        nextCursor: undefined,
      };
      return HttpResponse.json(result);
    }),
  ];
}
