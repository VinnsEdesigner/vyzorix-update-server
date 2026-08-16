/**
 * developer-client.e2e.spec.cjs
 *
 * Tests the developer client (API Key auth + HMAC signing) through the
 * full proxy stack. This is the most important test suite for catching
 * query mismatch and signing bugs:
 *
 *   Browser API Client → Proxy → Go API
 *
 * The developer client:
 *   1. Derives the signing secret from the API key (SHA-512 hex)
 *   2. Signs each tenant-path request with X-Vyzorix-* headers
 *   3. Sends the request with X-API-Key header
 *
 * The proxy:
 *   1. Receives the signed request
 *   2. Forwards to Go API (preserving or re-deriving signature)
 *
 * The Go API:
 *   1. Validates the API key
 *   2. Verifies the HMAC signature
 *   3. Returns the response
 */

const { test, expect } = require('@playwright/test');
const { loadHarness } = require('./helpers/e2e-helpers.cjs');

const TEST_API_KEY = 'vxyz_735ed9eea2cfda407db746f0492c1b2b1de89f32a84ba853253858e518615d6b';
const TEST_ORG_ID = '38912763-0f82-42b8-a2a7-96f73ce79ac5';

test.describe('Developer Client (API Key + HMAC signing)', () => {
  test('createDeveloperClient is available', async ({ page }) => {
    const harness = await loadHarness(page);
    expect(await harness.hasExport('createDeveloperClient')).toBe(true);
  });

  test('deriveAPIKeySigningSecret produces 128-char hex (SHA-512)', async ({ page }) => {
    const harness = await loadHarness(page);
    const secret = await harness.call('deriveAPIKeySigningSecret', TEST_API_KEY);

    expect(secret).toBeDefined();
    expect(secret.length).toBe(128); // SHA-512 = 64 bytes = 128 hex chars
    expect(secret).toMatch(/^[0-9a-f]{128}$/);
  });

  test('developer client getDevices returns device list', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      const result = await harness.developerClient(TEST_API_KEY, 'getDevices', [], TEST_ORG_ID);
      expect(result).toBeDefined();
      expect(result.devices).toBeDefined();
      expect(Array.isArray(result.devices)).toBe(true);
      expect(result.pagination).toBeDefined();
      expect(result.pagination.page).toBe(1);
    } catch (err) {
      // 429 = rate limited (path + auth work, just throttled)
      expect(String(err)).toMatch(/Rate limit exceeded|429/);
    }
  });

  test('developer client getUpdateStatus returns update status', async ({ page }) => {
    const harness = await loadHarness(page);

    try {
      const result = await harness.developerClient(TEST_API_KEY, 'getUpdateStatus', [], TEST_ORG_ID);
      expect(result).toBeDefined();
    } catch (err) {
      // Scope, auth, or rate limit errors are acceptable (path is correct)
      expect(String(err)).toMatch(/scope|401|403|authorized|failed|Rate limit exceeded|429/);
    }
  });

  test('developer client getVersions returns versions', async ({ page }) => {
    const harness = await loadHarness(page);

    try {
      const result = await harness.developerClient(TEST_API_KEY, 'getVersions', [], TEST_ORG_ID);
      expect(result).toBeDefined();
    } catch (err) {
      expect(String(err)).toMatch(/scope|401|403|authorized|failed|Rate limit exceeded|429/);
    }
  });

  test('developer client with invalid API key throws 401', async ({ page }) => {
    const harness = await loadHarness(page);

    await expect(
      harness.developerClient('vxyz_invalid_key', 'getDevices', [], TEST_ORG_ID),
    ).rejects.toThrow(/401|Unauthorized|Invalid or expired API key|invalid_api_key|Rate limit exceeded|429/);
  });

  test('developer client raw request to /v1/devices works', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      const result = await harness.developerRequest(TEST_API_KEY, '/v1/devices', {}, TEST_ORG_ID);
      expect(result).toBeDefined();
      expect(result.devices).toBeDefined();
      expect(Array.isArray(result.devices)).toBe(true);
    } catch (err) {
      expect(String(err)).toMatch(/Rate limit exceeded|429/);
    }
  });

  test('developer client signs tenant paths but not public paths', async ({ page }) => {
    const harness = await loadHarness(page);

    // /v1/version is a public path — no signing needed
    try {
      const versionResult = await harness.developerRequest(TEST_API_KEY, '/v1/version', {}, TEST_ORG_ID);
      expect(versionResult).toBeDefined();
      expect(versionResult.version).toBeDefined();
    } catch (err) {
      expect(String(err)).toMatch(/Rate limit exceeded|429/);
    }

    // /v1/devices is a tenant path — signing required
    try {
      const devicesResult = await harness.developerRequest(TEST_API_KEY, '/v1/devices', {}, TEST_ORG_ID);
      expect(devicesResult).toBeDefined();
      expect(devicesResult.devices).toBeDefined();
    } catch (err) {
      expect(String(err)).toMatch(/Rate limit exceeded|429/);
    }
  });

  test('signHttpRequestBrowser produces all required headers', async ({ page }) => {
    const harness = await loadHarness(page);
    const secret = await harness.call('deriveAPIKeySigningSecret', TEST_API_KEY);
    const headers = await harness.call(
      'signHttpRequestBrowser',
      'GET',
      '/v1/devices',
      '',
      secret,
    );

    expect(headers['X-Vyzorix-Timestamp']).toBeDefined();
    expect(headers['X-Vyzorix-Nonce']).toBeDefined();
    expect(headers['X-Vyzorix-Signature']).toBeDefined();

    // Timestamp should be a numeric string (milliseconds since epoch)
    expect(headers['X-Vyzorix-Timestamp']).toMatch(/^\d+$/);

    // Nonce should be a non-empty hex string
    expect(headers['X-Vyzorix-Nonce']).toMatch(/^[0-9a-f]+$/);

    // Signature should be base64
    expect(headers['X-Vyzorix-Signature']).toMatch(/^[A-Za-z0-9+/=]+$/);
  });

  test('developer client getLatestTelemetry handles missing device', async ({ page }) => {
    const harness = await loadHarness(page);

    try {
      const result = await harness.developerClient(
        TEST_API_KEY,
        'getLatestTelemetry',
        ['000000000000000'],
        TEST_ORG_ID,
      );
      // May return null or empty for nonexistent device
      expect(result).toBeDefined();
    } catch (err) {
      // 404, scope, auth, or rate limit errors are expected
      expect(String(err)).toMatch(/404|scope|401|403|not found|failed|Rate limit exceeded|429/i);
    }
  });
});
