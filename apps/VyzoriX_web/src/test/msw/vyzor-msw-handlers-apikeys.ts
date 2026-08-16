/**
 * MSW handlers for the operator-scoped API key endpoints.
 *
 * Mirrors the Go server contract in
 * apps/api/internal/api/handlers/auth/keys_handler.go:
 *   POST   /v1/auth/api-keys            -> 201, full key (api_key shown once)
 *   GET    /v1/auth/api-keys            -> 200, { keys, pagination, monthly_limit, keys_created_this_month }
 *   GET    /v1/auth/api-keys/:keyId     -> 200, single key (APIKeyResponse — no operator_id)
 *   PATCH  /v1/auth/api-keys/:keyId     -> 200, updated key
 *   DELETE /v1/auth/api-keys/:keyId     -> 204 No Content
 *   POST   /v1/auth/api-keys/:keyId/rotate -> 200, new full key
 *
 * Timestamps are RFC3339 ISO strings (Go time.Time serialization).
 */
import { http, HttpResponse, delay } from 'msw';

const API_BASE = '/v1/auth/api-keys';

const NOW = '2025-01-15T10:00:00.000Z';
const LATER = '2025-06-15T10:00:00.000Z';

interface MockKey {
  id: string;
  operator_id: string;
  name: string;
  key_prefix: string;
  api_key?: string;
  scope: string;
  expires_at: string | null;
  is_active: boolean;
  request_count: number;
  created_at: string;
  updated_at: string;
  last_request_at: string | null;
  revoked_at: string | null;
}

function seedKeys(): MockKey[] {
  return [
    {
      id: 'key-1',
      operator_id: 'op-1',
      name: 'Production Key',
      key_prefix: 'vxyz_abcd',
      scope: 'admin',
      expires_at: LATER,
      is_active: true,
      request_count: 142,
      created_at: NOW,
      updated_at: NOW,
      last_request_at: NOW,
      revoked_at: null,
    },
    {
      id: 'key-2',
      operator_id: 'op-1',
      name: 'Read Only Key',
      key_prefix: 'vxyz_efgh',
      scope: 'read',
      expires_at: null,
      is_active: true,
      request_count: 3,
      created_at: NOW,
      updated_at: NOW,
      last_request_at: null,
      revoked_at: null,
    },
    {
      id: 'key-3',
      operator_id: 'op-1',
      name: 'Old Key',
      key_prefix: 'vxyz_ijkl',
      scope: 'write',
      expires_at: null,
      is_active: false,
      request_count: 99,
      created_at: NOW,
      updated_at: NOW,
      last_request_at: null,
      revoked_at: NOW,
    },
  ];
}

// Module-level store so handlers share state within a test run. Tests reset
// MSW handlers via server.resetHandlers() between tests, so re-seeding happens
// on each createApiKeysHandlers() call (wired in the index).
let keys: MockKey[] = seedKeys();

export function resetApiKeyFixtures() {
  keys = seedKeys();
}

function toResponse(k: MockKey, includeSecret = false) {
  // APIKeyResponse omits the full secret; create/rotate include it once.
  if (includeSecret && k.api_key) {
    return { ...k, api_key: k.api_key };
  }
  const { api_key: _secret, ...rest } = k;
  void _secret;
  return rest;
}

export function createApiKeysHandlers() {
  return [
    http.get(`${API_BASE}`, async ({ request }) => {
      await delay(30);
      const url = new URL(request.url);
      const page = Math.max(1, Number(url.searchParams.get('page') ?? '1'));
      const limit = Math.max(1, Number(url.searchParams.get('limit') ?? '20'));

      const active = keys.filter((k) => k.is_active);
      const total = active.length;
      const totalPages = Math.max(1, Math.ceil(total / limit));
      const start = (page - 1) * limit;
      const paged = active.slice(start, start + limit);

      return HttpResponse.json({
        keys: paged.map((k) => toResponse(k)),
        pagination: {
          page,
          limit,
          total,
          total_pages: totalPages,
        },
        monthly_limit: 20,
        keys_created_this_month: 1,
      });
    }),

    http.get(`${API_BASE}/:keyId`, async ({ params }) => {
      await delay(30);
      const key = keys.find((k) => k.id === params.keyId);
      if (!key) {
        return HttpResponse.json(
          { error: 'api_key_not_found', message: 'api key not found' },
          { status: 404 },
        );
      }
      return HttpResponse.json(toResponse(key));
    }),

    http.post(`${API_BASE}`, async ({ request }) => {
      await delay(40);
      const body = (await request.json()) as {
        name: string;
        scope: string;
        expires_in_days?: number;
      };
      const id = `key-${Date.now()}`;
      const fullKey = `vxyz_secret_${id}`;
      const newKey: MockKey = {
        id,
        operator_id: 'op-1',
        name: body.name,
        key_prefix: fullKey.slice(0, 9),
        api_key: fullKey,
        scope: body.scope,
        expires_at: body.expires_in_days ? LATER : null,
        is_active: true,
        request_count: 0,
        created_at: NOW,
        updated_at: NOW,
        last_request_at: null,
        revoked_at: null,
      };
      keys.push(newKey);
      return HttpResponse.json(toResponse(newKey, true), { status: 201 });
    }),

    http.patch(`${API_BASE}/:keyId`, async ({ params, request }) => {
      await delay(30);
      const key = keys.find((k) => k.id === params.keyId);
      if (!key) {
        return HttpResponse.json(
          { error: 'api_key_not_found', message: 'api key not found' },
          { status: 404 },
        );
      }
      const body = (await request.json()) as { name?: string; scope?: string };
      if (body.name !== undefined) key.name = body.name;
      if (body.scope !== undefined) key.scope = body.scope;
      key.updated_at = NOW;
      return HttpResponse.json(toResponse(key));
    }),

    http.delete(`${API_BASE}/:keyId`, async ({ params }) => {
      await delay(20);
      const key = keys.find((k) => k.id === params.keyId);
      if (!key) {
        return HttpResponse.json(
          { error: 'api_key_not_found', message: 'api key not found' },
          { status: 404 },
        );
      }
      key.is_active = false;
      key.revoked_at = NOW;
      return new HttpResponse(null, { status: 204 });
    }),

    http.post(`${API_BASE}/:keyId/rotate`, async ({ params }) => {
      await delay(40);
      const key = keys.find((k) => k.id === params.keyId);
      if (!key) {
        return HttpResponse.json(
          { error: 'api_key_not_found', message: 'api key not found' },
          { status: 404 },
        );
      }
      const fullKey = `vxyz_secret_rotated_${Date.now()}`;
      key.key_prefix = fullKey.slice(0, 9);
      key.updated_at = NOW;
      return HttpResponse.json({ ...toResponse(key), api_key: fullKey });
    }),
  ];
}
