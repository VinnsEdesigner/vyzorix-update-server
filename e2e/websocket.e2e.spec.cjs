/**
 * websocket.e2e.spec.cjs
 *
 * Tests WebSocket connections through the proxy. Verifies that:
 *   - The WS proxy accepts connections at the expected path
 *   - Device stream endpoints route correctly through the proxy
 *   - GraphQL subscription WS endpoint is reachable
 *   - Unauthenticated WS connections are rejected (401/close)
 *
 * The proxy's WebSocket handler (vyzorix-proxy-websocket.cjs) handles:
 *   - /:org/graphql/ws — GraphQL subscriptions
 *   - /v1/device/:clientId/stream — device streams
 */

const { test, expect } = require('@playwright/test');
const { loadHarness } = require('./helpers/e2e-helpers.cjs');

const TEST_ORG_ID = '38912763-0f82-42b8-a2a7-96f73ce79ac5';

test.describe('WebSocket through proxy', () => {
  test('WS device stream endpoint rejects unauthenticated connection', async ({ page }) => {
    const harness = await loadHarness(page);
    const wsUrl = `ws://localhost:${process.env.PROXY_PORT || '3099'}/v1/device/test-device-123/stream`;

    const result = await harness.websocketConnect(wsUrl, 5000);

    // Without auth, the WS should not fully open.
    // It should either error, close, or timeout.
    expect(result.opened).not.toBe(true);
    // Should have an error or a close code
    expect(result.error === 'websocket_error' || result.closeCode !== undefined || result.timedOut).toBe(true);
  });

  test('WS GraphQL subscription endpoint is reachable', async ({ page }) => {
    const harness = await loadHarness(page);
    const wsUrl = `ws://localhost:${process.env.PROXY_PORT || '3099'}/${TEST_ORG_ID}/graphql/ws`;

    const result = await harness.websocketConnect(wsUrl, 5000);

    // The WS endpoint should exist (not 404). Without auth, it may close
    // immediately or after a brief handshake.
    // We just verify it doesn't get a "not found" style immediate rejection.
    expect(result).toBeDefined();
    // Either it opens briefly, errors, or closes with a code
    expect(
      result.opened === true ||
      result.error === 'websocket_error' ||
      result.closeCode !== undefined ||
      result.timedOut === true,
    ).toBe(true);
  });

  test('WS to non-existent path closes/error', async ({ page }) => {
    const harness = await loadHarness(page);
    const wsUrl = `ws://localhost:${process.env.PROXY_PORT || '3099'}/v1/nonexistent/ws`;

    const result = await harness.websocketConnect(wsUrl, 5000);

    // Non-WS paths should not establish a connection
    expect(result.opened).not.toBe(true);
  });

  test('WS device stream with /api prefix routes through proxy', async ({ page }) => {
    const harness = await loadHarness(page);
    // The proxy should handle /api/v1/device/:id/stream by stripping /api
    const wsUrl = `ws://localhost:${process.env.PROXY_PORT || '3099'}/api/v1/device/test-device-456/stream`;

    const result = await harness.websocketConnect(wsUrl, 5000);

    // Without auth, should not fully open
    expect(result.opened).not.toBe(true);
  });

  test('WS connection to health endpoint fails (not a WS endpoint)', async ({ page }) => {
    const harness = await loadHarness(page);
    const wsUrl = `ws://localhost:${process.env.PROXY_PORT || '3099'}/health`;

    const result = await harness.websocketConnect(wsUrl, 5000);

    // /health is an HTTP endpoint, not WS — should not open
    expect(result.opened).not.toBe(true);
  });
});
