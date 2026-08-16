import { http, HttpResponse, delay } from 'msw';
import {
  buildDeviceListItem,
  buildDeviceStats,
  resetFixtureCounter,
} from '../fixtures/vyzor-test-fixtures';
import type {
  RawDeviceListItem,
  RawDeviceListResult,
  RawDeviceStats,
} from '@vyzorix/api-client';

const API_BASE = '/v1/devices';

function toRawDeviceListItem(
  d: ReturnType<typeof buildDeviceListItem>,
): RawDeviceListItem {
  return { ...d } as unknown as RawDeviceListItem;
}

export function createDevicesHandlers() {
  resetFixtureCounter();
  const devices = [
    buildDeviceListItem({ imei: '111111111111111', status: 'online' }),
    buildDeviceListItem({ imei: '222222222222222', status: 'offline', id: 'device-test-2' }),
    buildDeviceListItem({ imei: '333333333333333', status: 'online', id: 'device-test-3' }),
  ];

  return [
    http.get(`${API_BASE}`, async ({ request }) => {
      await delay(50);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');

      const result: RawDeviceListResult = {
        devices: devices.map(toRawDeviceListItem),
        pagination: {
          page,
          limit,
          total: devices.length,
          totalPages: 1,
        },
      } as unknown as RawDeviceListResult;
      return HttpResponse.json(result);
    }),

    http.get(`${API_BASE}/stats`, async () => {
      await delay(30);
      const stats: RawDeviceStats = buildDeviceStats({
        total: devices.length,
        online: devices.filter((d) => d.status === 'online').length,
        offline: devices.filter((d) => d.status === 'offline').length,
      }) as unknown as RawDeviceStats;
      return HttpResponse.json(stats);
    }),

    http.get(`${API_BASE}/count`, async () => {
      await delay(20);
      return HttpResponse.json({ count: devices.length });
    }),

    http.get(`${API_BASE}/:imei`, async ({ params }) => {
      await delay(30);
      const imei = params.imei as string;
      const item = devices.find((d) => d.imei === imei);
      if (!item) {
        return HttpResponse.json({ error: 'device not found' }, { status: 404 });
      }
      // Return a proper RawDevice (camelCase, online boolean, numeric timestamps)
      // matching what the Go server actually returns and what deviceFromRaw expects.
      const now = Date.now();
      return HttpResponse.json({
        id: item.id,
        imei: item.imei,
        deviceName: item.device_name,
        model: item.model,
        manufacturer: item.manufacturer,
        osVersion: '14.0',
        appVersion: 'v1.1.0',
        securityPatch: '2025-01-01',
        online: item.status === 'online',
        fcmTokenValid: true,
        commandSecretSet: true,
        registeredAt: now - 86_400_000,
        lastSeen: now,
        createdAt: now - 86_400_000,
        updatedAt: now,
      });
    }),

    http.get(`${API_BASE}/:imei/connection-status`, async ({ params }) => {
      await delay(20);
      const imei = params.imei as string;
      return HttpResponse.json({
        imei,
        connected: devices.find((d) => d.imei === imei)?.status === 'online',
      });
    }),

    http.post(`${API_BASE}/:imei/disconnect`, async ({ params }) => {
      await delay(60);
      return HttpResponse.json({
        imei: params.imei,
        disconnected: true,
      });
    }),

    // GET /v1/devices/:imei/settings
    http.get(`${API_BASE}/:imei/settings`, async ({ params }) => {
      await delay(20);
      return HttpResponse.json({
        id: `settings-${params.imei}`,
        deviceImei: params.imei as string,
        customName: 'Test Device',
        location: 'Test Lab',
        metadata: {},
        thresholds: { riskWarn: 70, riskCrit: 90, thermalWarn: 60, thermalCrit: 85, bufferWarn: 70, bufferCrit: 95 },
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      });
    }),

    // PATCH /v1/devices/:imei/settings
    http.patch(`${API_BASE}/:imei/settings`, async ({ request, params }) => {
      const body = (await request.json()) as Record<string, unknown>;
      return HttpResponse.json({
        id: `settings-${params.imei}`,
        deviceImei: params.imei as string,
        customName: (body.customName as string) ?? 'Test Device',
        location: (body.location as string) ?? 'Test Lab',
        metadata: (body.metadata as Record<string, string>) ?? {},
        thresholds: body.thresholds ?? { riskWarn: 70, riskCrit: 90, thermalWarn: 60, thermalCrit: 85, bufferWarn: 70, bufferCrit: 95 },
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: new Date().toISOString(),
      });
    }),

    // GET /v1/devices/:imei/settings/thresholds
    http.get(`${API_BASE}/:imei/settings/thresholds`, async () => {
      await delay(20);
      return HttpResponse.json({
        thresholds: { riskWarn: 70, riskCrit: 90, thermalWarn: 60, thermalCrit: 85, bufferWarn: 70, bufferCrit: 95 },
      });
    }),

    // PATCH /v1/devices/:imei/settings/thresholds
    http.patch(`${API_BASE}/:imei/settings/thresholds`, async ({ request }) => {
      const body = (await request.json()) as Record<string, number>;
      return HttpResponse.json({
        thresholds: {
          riskWarn: body.riskWarn ?? 70,
          riskCrit: body.riskCrit ?? 90,
          thermalWarn: body.thermalWarn ?? 60,
          thermalCrit: body.thermalCrit ?? 85,
          bufferWarn: body.bufferWarn ?? 70,
          bufferCrit: body.bufferCrit ?? 95,
        },
      });
    }),

    // DELETE /v1/devices/:imei — deregister
    http.delete(`${API_BASE}/:imei`, async ({ params }) => {
      await delay(30);
      return HttpResponse.json({ success: true, imei: params.imei });
    }),
  ];
}
