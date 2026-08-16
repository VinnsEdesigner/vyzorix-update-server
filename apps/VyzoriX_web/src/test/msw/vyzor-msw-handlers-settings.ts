import { http, HttpResponse, delay } from 'msw';
import {
  buildThresholds,
  buildOperatorSettings,
  buildOrgSettings,
  buildDeviceSettings,
  resetFixtureCounter,
} from '../fixtures/vyzor-test-fixtures';

const AUTH_BASE = '/v1/auth';
const ORG_BASE = '/v1/organizations';
const DEVICE_BASE = '/v1/devices';

let operatorSettings = buildOperatorSettings();
let orgSettings = buildOrgSettings();
let deviceSettings = buildDeviceSettings();

export function createSettingsHandlers() {
  resetFixtureCounter();
  operatorSettings = buildOperatorSettings();
  orgSettings = buildOrgSettings();
  deviceSettings = buildDeviceSettings();

  return [
    http.get(`${AUTH_BASE}/me/settings`, async () => {
      await delay(30);
      return HttpResponse.json(operatorSettings);
    }),

    http.patch(`${AUTH_BASE}/me/settings`, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as { client?: Record<string, unknown> };
      if (body?.client) {
        operatorSettings = {
          ...operatorSettings,
          client: { ...operatorSettings.client, ...body.client },
        };
      }
      return HttpResponse.json(operatorSettings);
    }),

    http.get(`${AUTH_BASE}/me/thresholds`, async () => {
      await delay(20);
      return HttpResponse.json(operatorSettings.thresholds);
    }),

    http.patch(`${AUTH_BASE}/me/thresholds`, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as Record<string, number>;
      operatorSettings = {
        ...operatorSettings,
        thresholds: { ...operatorSettings.thresholds, ...body },
      };
      return HttpResponse.json(operatorSettings.thresholds);
    }),

    http.get(`${AUTH_BASE}/me/notifications`, async () => {
      await delay(30);
      return HttpResponse.json(operatorSettings.notifications);
    }),

    http.patch(`${AUTH_BASE}/me/notifications`, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as Record<string, unknown>;
      operatorSettings = {
        ...operatorSettings,
        notifications: {
          ...operatorSettings.notifications,
          ...body,
          webhook: { ...operatorSettings.notifications.webhook, ...(body.webhook as Record<string, unknown> ?? {}) },
        },
      };
      return HttpResponse.json(operatorSettings.notifications);
    }),

    http.post(`${AUTH_BASE}/me/notifications/webhook/test`, async ({ request }) => {
      await delay(100);
      const body = (await request.json()) as { url?: string };
      if (!body?.url) {
        return HttpResponse.json(
          { success: false, error: 'URL is required' },
          { status: 400 },
        );
      }
      return HttpResponse.json({
        success: true,
        statusCode: 200,
        responseTime: 42,
      });
    }),

    http.post(`${AUTH_BASE}/me/notifications/webhook/rotate`, async () => {
      await delay(50);
      const newSecret = 'whsec_rotated_' + Date.now();
      operatorSettings = {
        ...operatorSettings,
        notifications: {
          ...operatorSettings.notifications,
          webhook: { ...operatorSettings.notifications.webhook, secret: newSecret },
        },
      };
      return HttpResponse.json({ secret: newSecret });
    }),

    http.get(`${ORG_BASE}/:id/settings`, async () => {
      await delay(30);
      return HttpResponse.json(orgSettings);
    }),

    http.patch(`${ORG_BASE}/:id/settings`, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as Record<string, unknown>;
      orgSettings = {
        ...orgSettings,
        ...body,
        defaultThresholds: body.defaultThresholds
          ? (body.defaultThresholds as typeof orgSettings.defaultThresholds)
          : orgSettings.defaultThresholds,
      } as typeof orgSettings;
      return HttpResponse.json(orgSettings);
    }),

    http.get(`${ORG_BASE}/:id/settings/thresholds`, async () => {
      await delay(20);
      return HttpResponse.json({ thresholds: orgSettings.defaultThresholds });
    }),

    http.patch(`${ORG_BASE}/:id/settings/thresholds`, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as Record<string, number>;
      orgSettings = {
        ...orgSettings,
        defaultThresholds: { ...orgSettings.defaultThresholds!, ...body },
      };
      return HttpResponse.json({ thresholds: orgSettings.defaultThresholds });
    }),

    http.get(`${DEVICE_BASE}/:imei/settings`, async () => {
      await delay(30);
      return HttpResponse.json(deviceSettings);
    }),

    http.patch(`${DEVICE_BASE}/:imei/settings`, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as Record<string, unknown>;
      deviceSettings = {
        ...deviceSettings,
        ...body,
        thresholds: body.thresholds
          ? (body.thresholds as typeof deviceSettings.thresholds)
          : deviceSettings.thresholds,
      } as typeof deviceSettings;
      return HttpResponse.json(deviceSettings);
    }),

    http.get(`${DEVICE_BASE}/:imei/settings/thresholds`, async () => {
      await delay(20);
      return HttpResponse.json({ thresholds: deviceSettings.thresholds ?? buildThresholds() });
    }),

    http.patch(`${DEVICE_BASE}/:imei/settings/thresholds`, async ({ request }) => {
      await delay(30);
      const body = (await request.json()) as Record<string, number>;
      deviceSettings = {
        ...deviceSettings,
        thresholds: { ...deviceSettings.thresholds!, ...body },
      };
      return HttpResponse.json({ thresholds: deviceSettings.thresholds });
    }),
  ];
}
