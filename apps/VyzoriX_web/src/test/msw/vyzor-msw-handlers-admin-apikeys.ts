/**
 * MSW handlers for the super-admin API key endpoints.
 *
 * Mirrors the Go server contract in
 * apps/api/internal/api/handlers/admin/key_admin_handler.go:
 *   GET    /v1/admin/api-keys                       -> 200, { keys, pagination }
 *   GET    /v1/admin/api-keys/operator/:operatorId  -> 200, { keys, pagination }
 *   DELETE /v1/admin/api-keys/:keyId                -> 204 No Content
 *   GET    /v1/admin/api-keys/stats                 -> 200, GlobalAPIKeyStats
 *   GET    /v1/admin/api-keys/stats/operator/:operatorId -> 200, operator stats
 *
 * Query params on list-all: page, limit, operator_id, search.
 * Query params on operator-keys: page, limit.
 * Timestamps are RFC3339 ISO strings (Go time.Time serialization).
 */
import { http, HttpResponse, delay } from 'msw';

const API_BASE = '/v1/admin/api-keys';

const NOW = '2025-01-15T10:00:00.000Z';

interface MockAdminKey {
  id: string;
  operator_id: string;
  operator_name: string;
  name: string;
  key_prefix: string;
  scope: string;
  expires_at: string | null;
  is_active: boolean;
  request_count: number;
  created_at: string;
  updated_at: string;
  last_request_at: string | null;
  revoked_at: string | null;
}

function makeKey(
  id: string,
  operatorId: string,
  operatorName: string,
  name: string,
  scope: string,
  isActive = true,
  requestCount = 0,
): MockAdminKey {
  return {
    id,
    operator_id: operatorId,
    operator_name: operatorName,
    name,
    key_prefix: `vxyz_${id.slice(-4)}`,
    scope,
    expires_at: null,
    is_active: isActive,
    request_count: requestCount,
    created_at: NOW,
    updated_at: NOW,
    last_request_at: isActive ? NOW : null,
    revoked_at: isActive ? null : NOW,
  };
}

function seedKeys(): MockAdminKey[] {
  return [
    makeKey('admin-key-1', 'op-1', 'Acme Corp', 'Production Key', 'admin', true, 142),
    makeKey('admin-key-2', 'op-1', 'Acme Corp', 'Read Only Key', 'read', true, 30),
    makeKey('admin-key-3', 'op-2', 'Globex Inc', 'Write Key', 'write', true, 50),
    makeKey('admin-key-4', 'op-2', 'Globex Inc', 'Revoked Key', 'read', false, 5),
    makeKey('admin-key-5', 'op-3', 'Initech', 'CI Key', 'write', true, 88),
    makeKey('admin-key-6', 'op-3', 'Initech', 'Legacy Key', 'read', false, 12),
    makeKey('admin-key-7', 'op-1', 'Acme Corp', 'Staging Key', 'read', true, 7),
    makeKey('admin-key-8', 'op-2', 'Globex Inc', 'Mobile Key', 'admin', true, 210),
  ];
}

let keys: MockAdminKey[] = seedKeys();

export function resetAdminApiKeyFixtures() {
  keys = seedKeys();
}

export function createAdminApiKeysHandlers() {
  return [
    http.get(`${API_BASE}`, async ({ request }) => {
      await delay(30);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');
      const operatorId = url.searchParams.get('operator_id') ?? '';
      const search = url.searchParams.get('search') ?? '';

      let filtered = keys;
      if (operatorId) filtered = filtered.filter((k) => k.operator_id === operatorId);
      if (search) {
        const lower = search.toLowerCase();
        filtered = filtered.filter(
          (k) =>
            k.name.toLowerCase().includes(lower) ||
            k.operator_name.toLowerCase().includes(lower),
        );
      }

      const total = filtered.length;
      const totalPages = Math.max(1, Math.ceil(total / limit));
      const start = (page - 1) * limit;
      const paged = filtered.slice(start, start + limit);

      return HttpResponse.json({
        keys: paged,
        pagination: {
          page,
          limit,
          total,
          total_pages: totalPages,
        },
      });
    }),

    http.get(`${API_BASE}/operator/:operatorId`, async ({ request, params }) => {
      await delay(30);
      const url = new URL(request.url);
      const page = Number(url.searchParams.get('page') ?? '1');
      const limit = Number(url.searchParams.get('limit') ?? '20');
      const operatorId = String(params.operatorId);

      const filtered = keys.filter((k) => k.operator_id === operatorId);
      const total = filtered.length;
      const totalPages = Math.max(1, Math.ceil(total / limit));
      const start = (page - 1) * limit;
      const paged = filtered.slice(start, start + limit);

      return HttpResponse.json({
        keys: paged,
        pagination: {
          page,
          limit,
          total,
          total_pages: totalPages,
        },
      });
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

    http.get(`${API_BASE}/stats`, async () => {
      await delay(30);
      const active = keys.filter((k) => k.is_active);
      const revoked = keys.filter((k) => !k.is_active);
      const totalRequests = keys.reduce((sum, k) => sum + k.request_count, 0);

      const byScope: Record<string, number> = {};
      for (const k of keys) {
        byScope[k.scope] = (byScope[k.scope] ?? 0) + k.request_count;
      }

      const opMap = new Map<string, { name: string; requests: number; active: number }>();
      for (const k of keys) {
        const entry = opMap.get(k.operator_id) ?? { name: k.operator_name, requests: 0, active: 0 };
        entry.requests += k.request_count;
        if (k.is_active) entry.active += 1;
        opMap.set(k.operator_id, entry);
      }
      const topOperators = Array.from(opMap.entries())
        .map(([id, v]) => ({
          operator_id: id,
          operator_name: v.name,
          total_requests: v.requests,
          active_key_count: v.active,
        }))
        .sort((a, b) => b.total_requests - a.total_requests);

      return HttpResponse.json({
        total_keys: keys.length,
        active_keys: active.length,
        revoked_keys: revoked.length,
        total_requests: totalRequests,
        requests_by_scope: byScope,
        top_operators: topOperators,
        total_operators: opMap.size,
        max_per_month: 20,
      });
    }),

    http.get(`${API_BASE}/stats/operator/:operatorId`, async ({ params }) => {
      await delay(30);
      const operatorId = String(params.operatorId);
      const opKeys = keys.filter((k) => k.operator_id === operatorId);
      const active = opKeys.filter((k) => k.is_active);
      const revoked = opKeys.filter((k) => !k.is_active);

      return HttpResponse.json({
        operator_id: operatorId,
        total_keys: opKeys.length,
        active_keys: active.length,
        revoked_keys: revoked.length,
        keys_created_this_month: 1,
        monthly_limit: 20,
      });
    }),
  ];
}
