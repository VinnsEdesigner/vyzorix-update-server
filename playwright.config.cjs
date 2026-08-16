/**
 * playwright.config.cjs
 *
 * Root Playwright configuration for Vyzorix E2E tests.
 *
 * Architecture:
 *   Headless Browser → Proxy (:3099) → Go API (:3000)
 *
 * The proxy serves the test harness at /__e2e__/ which loads the API Client
 * browser bundle. Tests navigate to the harness and exercise the real API
 * Client (REST, GraphQL, WebSocket, developer-client) through the full proxy
 * stack — catching query mismatches, routing bugs, and signing errors.
 *
 * No Vite. No separate dev server. The proxy serves everything.
 */

const { defineConfig, devices } = require('@playwright/test');

const PROXY_PORT = process.env.PROXY_PORT || '3099';
const API_PORT = process.env.API_PORT || '3000';
const PROXY_URL = `http://localhost:${PROXY_PORT}`;
const API_URL = `http://localhost:${API_PORT}`;

module.exports = defineConfig({
  testDir: './e2e',
  testMatch: /.*\.e2e\.spec\.cjs/,
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 0,
  workers: 1,
  reporter: process.env.CI ? [['github'], ['html', { open: 'never' }]] : 'list',
  timeout: 30_000,
  expect: { timeout: 10_000 },

  use: {
    baseURL: PROXY_URL,
    trace: 'on-first-retry',
    screenshot: 'only-on-failure',
    video: 'retain-on-failure',
  },

  projects: [
    {
      name: 'chromium',
      use: {
        ...devices['Desktop Chrome'],
        channel: undefined,
        launchOptions: {
          args: ['--no-sandbox', '--disable-setuid-sandbox'],
        },
      },
    },
  ],

  webServer: process.env.E2E_BASE_URL
    ? undefined
    : {
        command: `node apps/proxy-server/vyzorix-proxy-server.cjs`,
        url: `${PROXY_URL}/proxy-health`,
        reuseExistingServer: !process.env.CI,
        timeout: 30_000,
        env: {
          PROXY_PORT: PROXY_PORT,
          API_TARGET: API_URL,
          PROXY_SIGN: 'true',
          NODE_ENV: 'development',
        },
      },
});
