import { http, HttpResponse, delay } from 'msw';
import {
  buildDeviceListItem,
  resetFixtureCounter,
} from '../fixtures/vyzor-test-fixtures';
import type {
  DeviceListResult,
  DeviceDetailResult,
} from '@vyzorix/api-client';

const API_BASE = '/v1/devices';

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

      const start = (page - 1) * limit;
      const pageItems = devices.slice(start, start + limit);
      const result: DeviceListResult = {
        devices: pageItems,
        total: devices.length,
      };
      return HttpResponse.json(result);
    }),


    // GET /v1/device/:imei — device-management SDK path for a single device.
    http.get('/v1/device/count', async () => {
      await delay(20);
      return HttpResponse.json({ count: devices.length });
    }),
    http.get('/v1/device/:imei', async ({ params }) => {
      await delay(30);
      const imei = params.imei as string;
      // Don't shadow the inbox routes (/v1/device/inbox*) — return undefined so
      // MSW falls through to the registration handlers.
      if (imei === 'inbox') {
        return undefined;
      }
      const item = devices.find((d) => d.imei === imei);
      if (!item) {
        return HttpResponse.json({ error: 'device not found' }, { status: 404 });
      }
      const now = Date.now();
      const detail: DeviceDetailResult = {
        id: item.id,
        imei: item.imei,
        device_name: item.device_name,
        model: item.model,
        manufacturer: item.manufacturer,
        app_version: item.app_version,
        status: item.status,
        registered_at: now - 86_400_000,
        last_seen: now,
      };
      return HttpResponse.json(detail);
    }),

    // GET /v1/device/count — device count (SDK path).

    http.get(`${API_BASE}/:imei`, async ({ params }) => {
      await delay(30);
      const imei = params.imei as string;
      const item = devices.find((d) => d.imei === imei);
      if (!item) {
        return HttpResponse.json({ error: 'device not found' }, { status: 404 });
      }
      const now = Date.now();
      const detail: DeviceDetailResult = {
        id: item.id,
        imei: item.imei,
        device_name: item.device_name,
        model: item.model,
        manufacturer: item.manufacturer,
        app_version: item.app_version,
        status: item.status,
        registered_at: now - 86_400_000,
        last_seen: now,
      };
      return HttpResponse.json(detail);
    }),

    http.get(`${API_BASE}/:imei/connection-status`, async ({ params }) => {
      await delay(20);
      const imei = params.imei as string;
      return HttpResponse.json({
        device_id: imei,
        online: devices.find((d) => d.imei === imei)?.status === 'online',
        status: devices.find((d) => d.imei === imei)?.status ?? 'offline',
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
