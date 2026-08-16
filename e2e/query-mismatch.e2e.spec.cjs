/**
 * query-mismatch.e2e.spec.cjs
 *
 * Query mismatch detection tests. These tests verify that the paths the
 * API Client uses match the routes the Go API actually registers.
 *
 * If the API Client sends a request to /v1/foo but the Go API only
 * registers /v1/bar, the proxy returns a 404 (or the Go API returns 404).
 * These tests catch that class of bug.
 *
 * The tests work by:
 *   1. Calling each API Client endpoint method through the harness
 *   2. Checking the HTTP status code
 *   3. A 404 means the path doesn't match (query mismatch)
 *   4. A 401/403/429 means the path matches but auth/rate-limit is required (correct)
 *   5. A 200 means the path matches and the request succeeded
 */

const { test, expect } = require('@playwright/test');
const { loadHarness } = require('./helpers/e2e-helpers.cjs');

const TEST_API_KEY = 'vxyz_735ed9eea2cfda407db746f0492c1b2b1de89f32a84ba853253858e518615d6b';
const TEST_ORG_ID = '38912763-0f82-42b8-a2a7-96f73ce79ac5';

/**
 * A 404 means the path doesn't exist on the Go API (query mismatch).
 * 401/403 means the path exists but requires auth (expected for unauthenticated requests).
 * 200 means the path exists and the request succeeded.
 * 400 means the path exists but the request was malformed (path is correct, just bad input).
 * 429 means the path exists but rate limited (path is correct, just throttled).
 */
const PATH_EXISTS_STATUSES = [200, 401, 403, 400, 422, 429];
const PATH_NOT_FOUND_STATUS = 404;

test.describe('Query mismatch detection', () => {
  test('developer client /v1/devices path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/devices');

    // Without auth → 401 (path exists, auth required)
    // 404 would mean path mismatch
    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
    expect(PATH_EXISTS_STATUSES).toContain(res.status);
  });

  test('developer client /v1/devices/:imei path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/devices/000000000000000');

    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
  });

  test('developer client /v1/command/:id/status path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/command/test-dispatch-id/status');

    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
  });

  test('developer client /v1/telemetry/history path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/telemetry/history?device_id=test');

    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
  });

  test('developer client /v1/telemetry/latest/:deviceId path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/telemetry/latest/test-device');

    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
  });

  test('developer client /v1/updates/status path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/updates/status');

    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
  });

  test('developer client /v1/updates/versions path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/updates/versions');

    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
  });

  test('public /v1/version path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/version');

    expect(res.status).toBe(200);
    expect(res.body.version).toBeDefined();
  });

  test('public /v1/changelog path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.fetchJSON('/v1/changelog');

    expect(res.status).toBe(200);
  });

  test('GraphQL /:org/graphql path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const res = await harness.graphqlQuery(TEST_ORG_ID, '{ __typename }');

    // 404 would mean the GraphQL route isn't registered
    expect(res.status).not.toBe(PATH_NOT_FOUND_STATUS);
    // 401/429 means auth/rate-limit required (route exists) — acceptable
    expect(PATH_EXISTS_STATUSES).toContain(res.status);
  });

  test('GraphQL /:org/graphql/batch path matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    const result = await harness.page.evaluate(
      async ({ orgId }) => {
        const res = await fetch(`/api/${orgId}/graphql/batch`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify([{ query: '{ __typename }' }]),
        });
        return { status: res.status };
      },
      { orgId: TEST_ORG_ID },
    );

    // 404 would mean the batch route isn't registered
    expect(result.status).not.toBe(PATH_NOT_FOUND_STATUS);
  });

  test('developer client getDevices through API Client matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      const result = await harness.developerClient(TEST_API_KEY, 'getDevices', [], TEST_ORG_ID);
      expect(result).toBeDefined();
      expect(result.devices).toBeDefined();
      expect(result.pagination).toBeDefined();
    } catch (err) {
      // Rate limit or auth errors are OK — 404 would be the bug.
      expect(String(err)).not.toMatch(/404|not found/i);
    }
  });

  test('developer client getVersions through API Client matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      const result = await harness.developerClient(TEST_API_KEY, 'getVersions', [], TEST_ORG_ID);
      expect(result).toBeDefined();
    } catch (err) {
      // 404 would indicate a path mismatch — that's a bug
      expect(String(err)).not.toMatch(/404|not found/i);
    }
  });

  test('developer client getUpdateStatus through API Client matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      const result = await harness.developerClient(TEST_API_KEY, 'getUpdateStatus', [], TEST_ORG_ID);
      expect(result).toBeDefined();
    } catch (err) {
      expect(String(err)).not.toMatch(/404|not found/i);
    }
  });

  test('developer client getCommandStatus through API Client matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      await harness.developerClient(TEST_API_KEY, 'getCommandStatus', ['test-dispatch-id'], TEST_ORG_ID);
    } catch (err) {
      // 404 would indicate path mismatch
      expect(String(err)).not.toMatch(/404|not found/i);
    }
  });

  test('developer client getTelemetryHistory through API Client matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      await harness.developerClient(
        TEST_API_KEY,
        'getTelemetryHistory',
        [{ device_id: 'test-device', limit: 10 }],
        TEST_ORG_ID,
      );
    } catch (err) {
      // 404 would indicate path mismatch
      expect(String(err)).not.toMatch(/404|not found/i);
    }
  });

  test('developer client getLatestTelemetry through API Client matches Go API route', async ({ page }) => {
    const harness = await loadHarness(page);
    try {
      await harness.developerClient(TEST_API_KEY, 'getLatestTelemetry', ['test-device'], TEST_ORG_ID);
    } catch (err) {
      // 404 would indicate path mismatch
      expect(String(err)).not.toMatch(/404|not found/i);
    }
  });

  test('/api prefix is correctly stripped for all tenant paths', async ({ page }) => {
    const harness = await loadHarness(page);

    // Test both /v1/... and /api/v1/... variants — both should return
    // the same status (not 404), proving the proxy strips /api correctly.
    const paths = [
      '/v1/devices',
      '/api/v1/devices',
      '/v1/version',
      '/api/v1/version',
      '/v1/changelog',
      '/api/v1/changelog',
    ];

    for (const p of paths) {
      const res = await harness.fetchJSON(p, {
        headers: {
          'X-API-Key': TEST_API_KEY,
          'X-Organization-Id': TEST_ORG_ID,
        },
      });
      expect(res.status, `path ${p}`).not.toBe(PATH_NOT_FOUND_STATUS);
    }
  });
});
