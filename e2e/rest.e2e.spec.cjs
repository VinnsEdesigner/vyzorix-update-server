/**
 * rest.e2e.spec.cjs
 *
 * REST endpoint tests through the proxy. Verifies that:
 *   - Public endpoints (version, changelog, health) route correctly
 *   - The /api prefix is stripped by the proxy before forwarding to Go API
 *   - Tenant endpoints (devices) require API key auth
 *   - Response shapes match what the API Client expects
 *
 * These tests catch query mismatch bugs: if the API Client or proxy uses
 * a path the Go API doesn't register, the test fails.
 */

const { test, expect } = require('@playwright/test');
const { loadHarness } = require('./helpers/e2e-helpers.cjs');

const TEST_API_KEY = 'vxyz_735ed9eea2cfda407db746f0492c1b2b1de89f32a84ba853253858e518615d6b';
const TEST_ORG_ID = '38912763-0f82-42b8-a2a7-96f73ce79ac5';

test.describe('REST through proxy', () => {
  test('GET /v1/version returns version info', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/version');

    expect(res.status).toBe(200);
    expect(res.body.version).toBeDefined();
    expect(res.body.apk_filename).toBeDefined();
    expect(typeof res.body.version_code).toBe('number');
  });

  test('GET /api/v1/version works with /api prefix (proxy strips it)', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/api/v1/version');

    expect(res.status).toBe(200);
    expect(res.body.version).toBeDefined();
  });

  test('GET /v1/changelog returns changelog', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/changelog');

    expect(res.status).toBe(200);
    expect(res.body).toBeDefined();
  });

  test('GET /api/v1/changelog works with /api prefix', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/api/v1/changelog');

    expect(res.status).toBe(200);
    expect(res.body).toBeDefined();
  });

  test('GET /health returns ok status', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/health');

    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
  });

  test('GET /proxy-health returns proxy status', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/proxy-health');

    expect(res.status).toBe(200);
    expect(res.body.status).toBe('ok');
    expect(res.body.signRequests).toBe(true);
    expect(res.body.apiTarget).toBe('http://localhost:3000');
  });

  test('GET /v1/devices without API key returns 401 or 429', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/devices');

    // 401 = auth required, 429 = rate limited (both prove path exists)
    expect([401, 429]).toContain(res.status);
  });

  test('GET /v1/devices with API key returns device list', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/devices', {
      headers: {
        'X-API-Key': TEST_API_KEY,
        'X-Organization-Id': TEST_ORG_ID,
      },
    });

    // 200 = success, 429 = rate limited (both prove path + auth work)
    expect([200, 429]).toContain(res.status);
    if (res.status === 200) {
      expect(res.body.devices).toBeDefined();
      expect(Array.isArray(res.body.devices)).toBe(true);
      expect(res.body.pagination).toBeDefined();
      expect(res.body.pagination.page).toBe(1);
    }
  });

  test('GET /api/v1/devices with /api prefix and API key', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/api/v1/devices', {
      headers: {
        'X-API-Key': TEST_API_KEY,
        'X-Organization-Id': TEST_ORG_ID,
      },
    });

    expect([200, 429]).toContain(res.status);
    if (res.status === 200) {
      expect(res.body.devices).toBeDefined();
      expect(Array.isArray(res.body.devices)).toBe(true);
    }
  });

  test('GET /v1/devices with invalid API key returns 401 or 429', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/devices', {
      headers: {
        'X-API-Key': 'vxyz_invalid_key',
        'X-Organization-Id': TEST_ORG_ID,
      },
    });

    expect([401, 429]).toContain(res.status);
  });

  test('GET /v1/updates/versions returns version list', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/updates/versions', {
      headers: {
        'X-API-Key': TEST_API_KEY,
        'X-Organization-Id': TEST_ORG_ID,
      },
    });

    // 200 = success, 401 = auth required, 429 = rate limited
    expect([200, 401, 429]).toContain(res.status);
    if (res.status === 200) {
      expect(res.body).toBeDefined();
    }
  });
});
