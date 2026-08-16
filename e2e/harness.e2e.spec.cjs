/**
 * harness.e2e.spec.cjs
 *
 * Smoke tests for the E2E test harness itself. Verifies that:
 *   - The harness page loads at /__e2e__/ through the proxy
 *   - The API Client browser bundle loads and initializes
 *   - Key exports are available on window.__vyzorix
 *   - The config resolves correctly (VITE_API_URL → /api)
 */

const { test, expect } = require('@playwright/test');
const { loadHarness } = require('./helpers/e2e-helpers.cjs');

test.describe('E2E Harness', () => {
  test('harness page loads through proxy', async ({ page }) => {
    const response = await page.goto('/__e2e__/');
    expect(response?.status()).toBe(200);
    await expect(page.locator('h1')).toHaveText('Vyzorix E2E Test Harness');
  });

  test('API Client bundle loads and initializes', async ({ page }) => {
    const harness = await loadHarness(page);
    const status = await page.locator('#status').textContent();
    expect(status).toContain('ready');
    expect(await harness.hasExport('createDeveloperClient')).toBe(true);
  });

  test('key API Client exports are available', async ({ page }) => {
    const harness = await loadHarness(page);
    const exports = [
      'createDeveloperClient',
      'restClient',
      'graphqlClient',
      'devices',
      'deriveAPIKeySigningSecret',
      'signHttpRequestBrowser',
      'setOrganizationContext',
      'setAuthToken',
      'setCSRFToken',
      'setSigningKey',
      'clearAuthContext',
    ];
    for (const name of exports) {
      expect(await harness.hasExport(name), `export "${name}" should exist`).toBe(true);
    }
  });

  test('config resolves VITE_API_URL to /api', async ({ page }) => {
    const harness = await loadHarness(page);
    const config = await harness.getConfig();
    expect(config.rest.baseURL).toBe('/api');
    expect(config.rest.timeout).toBe(30000);
    expect(config.rest.withCredentials).toBe(true);
  });

  test('WebSocket config resolves to proxy host', async ({ page }) => {
    const harness = await loadHarness(page);
    const config = await harness.getConfig();
    expect(config.ws.url).toContain('ws');
    expect(config.ws.url).toContain('/v1/device');
    expect(config.ws.reconnectInterval).toBe(3000);
  });

  test('no console errors on harness load', async ({ page }) => {
    const errors = [];
    page.on('console', (msg) => {
      if (msg.type() === 'error') errors.push(msg.text());
    });
    page.on('pageerror', (err) => errors.push(err.message));

    await loadHarness(page);
    // Filter out non-critical errors (favicon, etc.)
    const criticalErrors = errors.filter(
      (e) => !e.includes('favicon') && !e.includes('Failed to load resource'),
    );
    expect(criticalErrors).toHaveLength(0);
  });
});
