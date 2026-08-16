import type { ApiKeyScope } from "./apikey-entity";
import { parseScope } from "./apikey-mappers";
import { paginationFromRaw } from "../_shared";
import type {
  AdminApiKey,
  AdminApiKeyListResult,
  GlobalApiKeyStats,
  OperatorApiKeyStats,
  TopOperatorStat,
  RawAdminApiKey,
  RawAdminApiKeyListResult,
  RawGlobalApiKeyStats,
  RawOperatorApiKeyStats,
  RawTopOperatorStat,
} from "./admin-types";

export function adminApiKeyFromRaw(raw: RawAdminApiKey): AdminApiKey {
  return {
    id: raw.id,
    operatorId: raw.operator_id ?? "",
    operatorName: raw.operator_name ?? "",
    name: raw.name,
    keyPrefix: raw.key_prefix,
    scope: parseScope(raw.scope) as ApiKeyScope,
    isActive: raw.is_active,
    requestCount: raw.request_count,
    createdAt: new Date(raw.created_at),
    updatedAt: new Date(raw.updated_at),
    expiresAt: raw.expires_at ? new Date(raw.expires_at) : null,
    lastRequestAt: raw.last_request_at ? new Date(raw.last_request_at) : null,
    revokedAt: raw.revoked_at ? new Date(raw.revoked_at) : null,
  };
}

export function adminApiKeyListFromRaw(raw: RawAdminApiKeyListResult): AdminApiKeyListResult {
  return {
    keys: raw.keys.map(adminApiKeyFromRaw),
    pagination: paginationFromRaw(raw.pagination),
  };
}

export function topOperatorStatFromRaw(raw: RawTopOperatorStat): TopOperatorStat {
  return {
    operatorId: raw.operator_id,
    operatorName: raw.operator_name,
    totalRequests: raw.total_requests,
    activeKeyCount: raw.active_key_count,
  };
}

export function globalApiKeyStatsFromRaw(raw: RawGlobalApiKeyStats): GlobalApiKeyStats {
  return {
    totalKeys: raw.total_keys,
    activeKeys: raw.active_keys,
    revokedKeys: raw.revoked_keys,
    totalRequests: raw.total_requests,
    requestsByScope: raw.requests_by_scope ?? {},
    topOperators: (raw.top_operators ?? []).map(topOperatorStatFromRaw),
    totalOperators: raw.total_operators,
    maxPerMonth: raw.max_per_month,
  };
}

export function operatorApiKeyStatsFromRaw(raw: RawOperatorApiKeyStats): OperatorApiKeyStats {
  return {
    operatorId: raw.operator_id,
    totalKeys: raw.total_keys,
    activeKeys: raw.active_keys,
    revokedKeys: raw.revoked_keys,
    keysCreatedThisMonth: raw.keys_created_this_month,
    monthlyLimit: raw.monthly_limit,
  };
}
