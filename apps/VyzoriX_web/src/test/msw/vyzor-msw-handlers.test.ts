import { describe, it, expect, beforeAll, afterAll, afterEach } from 'vitest';
import { createVyzorMswServer } from '@/test/msw/vyzor-msw-server';

const server = createVyzorMswServer();

beforeAll(() => server.listen({ onUnhandledRequest: 'bypass' }));
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('vyzor-msw-handlers', () => {
  it('GET /v1/updates/versions returns version list', async () => {
    const res = await fetch('/v1/updates/versions?page=1&limit=20');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.versions).toHaveLength(3);
    expect(data.versions[0].version).toBe('v1.2.0');
    expect(data.versions[0].isLatest).toBe(true);
    expect(data.pagination.total).toBe(3);
  });

  it('GET /v1/updates/status returns sync state + latest', async () => {
    const res = await fetch('/v1/updates/status');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.sync.status).toBe('synced');
    expect(data.latest.version).toBe('v1.2.0');
  });

  it('GET /v1/updates/history returns push history', async () => {
    const res = await fetch('/v1/updates/history');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.pushes).toHaveLength(2);
    expect(data.pushes[0].status).toBe('completed');
  });

  it('GET /v1/updates/history/:pushId returns single push', async () => {
    const res = await fetch('/v1/updates/history/push-test-1');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.id).toBe('push-test-1');
  });

  it('GET /v1/updates/history/:pushId returns 404 for unknown', async () => {
    const res = await fetch('/v1/updates/history/nonexistent');
    expect(res.status).toBe(404);
  });

  it('POST /v1/updates/push creates new push', async () => {
    const res = await fetch('/v1/updates/push', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({
        version: 'v1.2.0',
        deviceIds: ['device-1', 'device-2'],
        installType: 'immediate',
      }),
    });
    expect(res.status).toBe(201);
    const data = await res.json();
    expect(data.version).toBe('v1.2.0');
    expect(data.status).toBe('pending');
    expect(data.devices.total).toBe(2);
  });

  it('POST /v1/updates/history/:pushId/cancel cancels push', async () => {
    const res = await fetch('/v1/updates/history/push-test-2/cancel', {
      method: 'POST',
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.status).toBe('cancelled');
  });

  it('POST /v1/updates/sync triggers sync', async () => {
    const res = await fetch('/v1/updates/sync', { method: 'POST' });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.status).toBe('started');
  });

  it('GET /v1/devices returns device list', async () => {
    const res = await fetch('/v1/devices');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.devices).toHaveLength(3);
    expect(data.devices[0].imei).toHaveLength(15);
  });

  it('GET /v1/devices/stats returns stats', async () => {
    const res = await fetch('/v1/devices/stats');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.total).toBe(3);
    expect(data.online + data.offline).toBe(data.total);
  });

  it('GET /v1/devices/:imei returns device detail', async () => {
    const res = await fetch('/v1/devices/111111111111111');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.imei).toBe('111111111111111');
  });

  it('GET /v1/devices/:imei returns 404 for unknown', async () => {
    const res = await fetch('/v1/devices/999999999999999');
    expect(res.status).toBe(404);
  });

  it('POST /v1/auth/login returns operator (browser — no tokens in body)', async () => {
    const res = await fetch('/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: 'test@vyzorix.com', password: 'pass' }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.operator_id).toBe('operator-test-1');
    expect(data.email).toBe('test@vyzorix.com');
    expect(data.role).toBe('admin');
    expect(data.mfa_enabled).toBe(false);
    expect(data.selected_organization.id).toBe('org-test-1');
    // Browser login does NOT return tokens (session cookie is set instead)
    expect(data.access_token).toBeUndefined();
  });

  it('POST /v1/auth/login/tokens returns operator + tokens (API client)', async () => {
    const res = await fetch('/v1/auth/login/tokens', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email: 'test@vyzorix.com', password: 'pass' }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.operator_id).toBe('operator-test-1');
    expect(data.access_token).toBe('mock-access-token');
    expect(data.refresh_token).toBe('mock-refresh-token');
    expect(data.session_id).toBe('session-test-1');
    expect(typeof data.expires_at).toBe('number');
  });

  it('POST /v1/auth/login returns 401 for missing credentials', async () => {
    const res = await fetch('/v1/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    });
    expect(res.status).toBe(401);
  });

  it('POST /v1/auth/refresh returns new tokens', async () => {
    const res = await fetch('/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ refresh_token: 'mock-refresh-token' }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.access_token).toBe('mock-access-token-refreshed');
    expect(data.refresh_token).toBe('mock-refresh-token-refreshed');
    expect(data.session_id).toBe('session-test-1');
  });

  it('POST /v1/auth/refresh returns 400 without refresh_token', async () => {
    const res = await fetch('/v1/auth/refresh', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    });
    expect(res.status).toBe(400);
  });

  it('GET /v1/auth/me returns operator profile', async () => {
    const res = await fetch('/v1/auth/me');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.id).toBe('operator-test-1');
    expect(data.email).toBe('test@vyzorix.com');
    expect(data.name).toBe('Test Operator');
    expect(data.mfa_enabled).toBe(false);
    expect(data.email_verified).toBe(true);
    expect(data.needs_organization).toBe(false);
    expect(data.organizations).toHaveLength(1);
    expect(data.organizations[0].id).toBe('org-test-1');
    expect(data.selected_organization.id).toBe('org-test-1');
  });

  it('GET /v1/auth/sessions returns session list', async () => {
    const res = await fetch('/v1/auth/sessions');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.sessions).toHaveLength(1);
    expect(data.sessions[0].id).toBe('session-test-1');
    expect(data.sessions[0].ip_address).toBe('192.168.1.1');
    expect(data.sessions[0].is_current).toBe(true);
    expect(data.total).toBe(1);
  });

  it('GET /v1/auth/sessions/concurrent returns concurrent session info', async () => {
    const res = await fetch('/v1/auth/sessions/concurrent');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.has_concurrent).toBe(false);
    expect(data.count).toBe(1);
    expect(data.sessions[0].session_id).toBe('session-test-1');
  });
});

describe('settings handlers — operator tier', () => {
  it('GET /v1/auth/me/settings returns client + thresholds + notifications', async () => {
    const res = await fetch('/v1/auth/me/settings');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.client.requestTimeoutMs).toBe(8000);
    expect(data.thresholds.riskWarn).toBe(70);
    expect(data.notifications.enabled).toBe(true);
  });

  it('PATCH /v1/auth/me/settings updates client settings', async () => {
    const res = await fetch('/v1/auth/me/settings', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ client: { requestTimeoutMs: 10000 } }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.client.requestTimeoutMs).toBe(10000);
  });

  it('GET /v1/auth/me/thresholds returns thresholds', async () => {
    const res = await fetch('/v1/auth/me/thresholds');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.riskWarn).toBe(70);
    expect(data.bufferCrit).toBe(95);
  });

  it('PATCH /v1/auth/me/thresholds updates threshold fields', async () => {
    const res = await fetch('/v1/auth/me/thresholds', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ riskWarn: 75 }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.riskWarn).toBe(75);
    expect(data.riskCrit).toBe(90);
  });

  it('GET /v1/auth/me/notifications returns notification settings', async () => {
    const res = await fetch('/v1/auth/me/notifications');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.channels).toContain('email');
    expect(data.webhook.enabled).toBe(true);
  });

  it('PATCH /v1/auth/me/notifications updates enabled flag', async () => {
    const res = await fetch('/v1/auth/me/notifications', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ enabled: false }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.enabled).toBe(false);
  });

  it('POST /v1/auth/me/notifications/webhook/test returns success', async () => {
    const res = await fetch('/v1/auth/me/notifications/webhook/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ url: 'https://hooks.example.com/test' }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.success).toBe(true);
    expect(data.statusCode).toBe(200);
  });

  it('POST /v1/auth/me/notifications/webhook/test returns 400 for missing url', async () => {
    const res = await fetch('/v1/auth/me/notifications/webhook/test', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({}),
    });
    expect(res.status).toBe(400);
  });

  it('POST /v1/auth/me/notifications/webhook/rotate returns new secret', async () => {
    const res = await fetch('/v1/auth/me/notifications/webhook/rotate', {
      method: 'POST',
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.secret).toMatch(/^whsec_rotated_/);
  });
});

describe('settings handlers — organization tier', () => {
  it('GET /v1/organizations/:id/settings returns org settings', async () => {
    const res = await fetch('/v1/organizations/org-test-1/settings');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.timezone).toBe('UTC');
    expect(data.dateFormat).toBe('YYYY-MM-DD');
    expect(data.alertCooldownMinutes).toBe(15);
    expect(data.defaultThresholds.riskWarn).toBe(70);
  });

  it('PATCH /v1/organizations/:id/settings updates timezone', async () => {
    const res = await fetch('/v1/organizations/org-test-1/settings', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ timezone: 'Europe/Stockholm' }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.timezone).toBe('Europe/Stockholm');
  });

  it('GET /v1/organizations/:id/settings/thresholds returns thresholds', async () => {
    const res = await fetch('/v1/organizations/org-test-1/settings/thresholds');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.thresholds.riskWarn).toBe(70);
    expect(data.thresholds.thermalCrit).toBe(85);
  });

  it('PATCH /v1/organizations/:id/settings/thresholds updates fields', async () => {
    const res = await fetch('/v1/organizations/org-test-1/settings/thresholds', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ thermalWarn: 80 }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.thresholds.thermalWarn).toBe(80);
    expect(data.thresholds.thermalCrit).toBe(85);
  });
});

describe('settings handlers — device tier', () => {
  it('GET /v1/devices/:imei/settings returns device settings', async () => {
    const res = await fetch('/v1/devices/123456789012345/settings');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.deviceImei).toBe('123456789012345');
    expect(data.customName).toBe('Test Device');
    expect(data.thresholds.riskWarn).toBe(70);
  });

  it('PATCH /v1/devices/:imei/settings updates customName', async () => {
    const res = await fetch('/v1/devices/123456789012345/settings', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ customName: 'Renamed Device' }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.customName).toBe('Renamed Device');
  });

  it('GET /v1/devices/:imei/settings/thresholds returns thresholds', async () => {
    const res = await fetch('/v1/devices/123456789012345/settings/thresholds');
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.thresholds.riskWarn).toBe(70);
  });

  it('PATCH /v1/devices/:imei/settings/thresholds updates fields', async () => {
    const res = await fetch('/v1/devices/123456789012345/settings/thresholds', {
      method: 'PATCH',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ bufferWarn: 70 }),
    });
    expect(res.status).toBe(200);
    const data = await res.json();
    expect(data.thresholds.bufferWarn).toBe(70);
    expect(data.thresholds.bufferCrit).toBe(95);
  });
});
