/**
 * Shared test utilities for MSW-based integration tests.
 *
 * Usage in a test file:
 *
 *   import { setupIntegrationTest } from '../helpers/integration-test-setup';
 *
 *   const { server, resetApiState } = setupIntegrationTest();
 *
 *   beforeEach(() => resetApiState());
 *
 * This starts the MSW server (intercepts all HTTP), stubs VITE_API_URL to ''
 * so the real API client sends requests to /v1/... (matching MSW handlers),
 * and provides resetApiState() to clear cached axios instances + auth context
 * between tests.
 */
import { beforeAll, afterAll, afterEach, beforeEach, vi } from 'vitest';
import { createVyzorMswServer } from '@/test/msw/vyzor-msw-server';
import { authContext, resetClientState, graphqlClient } from '@vyzorix/api-client';
import { clearGraphQLResponses } from '@/test/msw/vyzor-msw-handlers-graphql';

export function setupIntegrationTest() {
  const server = createVyzorMswServer();

  beforeAll(() => server.listen({ onUnhandledRequest: 'error' }));
  afterAll(() => server.close());
  afterEach(() => server.resetHandlers());

  function resetApiState() {
    vi.stubEnv('VITE_API_URL', '');
    vi.stubEnv('VITE_REST_WITH_CREDENTIALS', 'false');
    resetClientState();
    authContext.clear();
    clearGraphQLResponses();
    graphqlClient.setOrganization('');
  }

  beforeEach(() => resetApiState());
  afterEach(() => resetApiState());

  return { server, resetApiState };
}
