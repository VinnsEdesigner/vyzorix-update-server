/**
 * e2e-helpers.cjs
 *
 * Utilities for Playwright E2E tests. Provides a bridge between Node-side
 * test code and the browser-loaded API Client harness at /__e2e__/.
 *
 * Usage in tests:
 *   const harness = await loadHarness(page);
 *   const devices = await harness.developerClient('vxyz_...', 'getDevices');
 *   const version = await harness.fetchJSON('/v1/version');
 */

'use strict';

/**
 * Navigate to the harness page and wait for the API Client to be ready.
 * The harness sets window.__vyzorix when the ESM bundle finishes loading.
 */
async function loadHarness(page) {
  await page.goto('/__e2e__/');
  await page.waitForFunction(() => window.__vyzorix !== undefined, { timeout: 15_000 });
  await page.waitForSelector('#status:has-text("ready")', { timeout: 10_000 });
  return new HarnessClient(page);
}

/**
 * Wrapper that executes API Client calls inside the browser context.
 * Each method calls page.evaluate() with the real API Client functions.
 */
class HarnessClient {
  constructor(page) {
    this.page = page;
  }

  /**
   * Call a function on the loaded API Client namespace.
   * @param {string} fnName - function name on window.__vyzorix
   * @param {...any} args - serializable arguments
   */
  async call(fnName, ...args) {
    return this.page.evaluate(
      ({ fnName, args }) => {
        const fn = window.__vyzorix[fnName];
        if (typeof fn !== 'function') {
          throw new Error(`API Client function "${fnName}" not found or not a function`);
        }
        return Promise.resolve(fn(...args));
      },
      { fnName, args },
    );
  }

  // ── REST (direct fetch through proxy) ──────────────────────────

  /**
   * Fetch a REST endpoint through the proxy using the browser's fetch.
   * Tests query mismatch: the actual HTTP path the API Client uses vs.
   * what the Go API expects.
   */
  async fetchJSON(path, options = {}) {
    return this.page.evaluate(
      async ({ path, options }) => {
        const res = await fetch(path, {
          ...options,
          headers: {
            'Content-Type': 'application/json',
            ...(options.headers || {}),
          },
        });
        const text = await res.text();
        let body;
        try { body = JSON.parse(text); } catch { body = text; }
        return { status: res.status, ok: res.ok, body, headers: Object.fromEntries(res.headers.entries()) };
      },
      { path, options },
    );
  }

  // ── Developer Client (API Key auth + HMAC signing) ────────────

  /**
   * Create a developer client with an API key and call a method.
   * This exercises the full signing pipeline:
   *   client derives signing secret → signs request → proxy forwards → Go API verifies
   *
   * @param apiKey - full API key (vxyz_...)
   * @param method - developer client method name (e.g. 'getDevices')
   * @param args - method arguments
   * @param orgId - organization ID (injected as X-Organization-ID)
   */
  async developerClient(apiKey, method, args = [], orgId) {
    return this.page.evaluate(
      async ({ apiKey, method, args, orgId }) => {
        const { createDeveloperClient } = window.__vyzorix;
        const client = createDeveloperClient(apiKey, orgId ? { organizationId: orgId } : {});
        const fn = client[method];
        if (typeof fn !== 'function') {
          throw new Error(`Developer client method "${method}" not found`);
        }
        return Promise.resolve(fn(...args));
      },
      { apiKey, method, args, orgId },
    );
  }

  /**
   * Create a developer client and call the raw request function for
   * custom endpoint testing.
   */
  async developerRequest(apiKey, endpoint, init, orgId) {
    return this.page.evaluate(
      async ({ apiKey, endpoint, init, orgId }) => {
        const { createDeveloperClient } = window.__vyzorix;
        const client = createDeveloperClient(apiKey, orgId ? { organizationId: orgId } : {});
        return client.request(endpoint, init);
      },
      { apiKey, endpoint, init, orgId },
    );
  }

  // ── GraphQL ───────────────────────────────────────────────────

  /**
   * Execute a raw GraphQL query through the proxy using fetch.
   * Tests that the GraphQL endpoint path matches what the Go API expects.
   */
  async graphqlQuery(orgId, query, variables = {}) {
    return this.page.evaluate(
      async ({ orgId, query, variables }) => {
        const res = await fetch(`/api/${orgId}/graphql`, {
          method: 'POST',
          headers: { 'Content-Type': 'application/json' },
          body: JSON.stringify({ query, variables }),
        });
        const text = await res.text();
        let body;
        try { body = JSON.parse(text); } catch { body = text; }
        return { status: res.status, ok: res.ok, body };
      },
      { orgId, query, variables },
    );
  }

  // ── WebSocket ─────────────────────────────────────────────────

  /**
   * Open a WebSocket connection through the proxy and wait for
   * connection result (open/error/message/close).
   */
  async websocketConnect(url, timeoutMs = 5000) {
    return this.page.evaluate(
      async ({ url, timeoutMs }) => {
        return new Promise((resolve) => {
          const ws = new WebSocket(url);
          const result = { readyState: -1, error: null, message: null };
          const timer = setTimeout(() => {
            result.readyState = ws.readyState;
            result.timedOut = true;
            try { ws.close(); } catch {}
            resolve(result);
          }, timeoutMs);

          ws.onopen = () => {
            result.readyState = ws.readyState;
            result.opened = true;
            clearTimeout(timer);
            try { ws.close(); } catch {}
            resolve(result);
          };
          ws.onerror = () => {
            result.readyState = ws.readyState;
            result.error = 'websocket_error';
            clearTimeout(timer);
            resolve(result);
          };
          ws.onmessage = (e) => {
            result.message = typeof e.data === 'string' ? e.data : '[binary]';
          };
          ws.onclose = (e) => {
            result.readyState = ws.readyState;
            result.closeCode = e.code;
            result.closeReason = e.reason;
            clearTimeout(timer);
            resolve(result);
          };
        });
      },
      { url, timeoutMs },
    );
  }

  // ── Browser env inspection ────────────────────────────────────

  /** Get the resolved config the API Client sees. */
  async getConfig() {
    return this.page.evaluate(() => ({
      rest: window.__vyzorix.getRESTConfig(),
      ws: window.__vyzorix.getWebSocketConfig(),
    }));
  }

  /** Check that a specific export exists on the API Client namespace. */
  async hasExport(name) {
    return this.page.evaluate(
      (name) => typeof window.__vyzorix?.[name] !== 'undefined',
      name,
    );
  }
}

module.exports = { loadHarness, HarnessClient };
