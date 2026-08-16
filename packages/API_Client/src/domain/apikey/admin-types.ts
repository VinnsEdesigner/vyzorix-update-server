import type { ApiKeyScope } from "./apikey-entity";
import type { Pagination, RawPagination } from "../_shared";

/**
 * AdminApiKey — the domain (camelCase) form of a key in the super-admin
 * "all keys" view. Extends the operator-scoped key with operator identity,
 * which the operator-scoped ApiKey type omits (operator is implied by the
 * request context there).
 */
export interface AdminApiKey {
  id: string;
  operatorId: string;
  operatorName: string;
  name: string;
  keyPrefix: string;
  scope: ApiKeyScope;
  isActive: boolean;
  requestCount: number;
  createdAt: Date;
  updatedAt: Date;
  expiresAt: Date | null;
  lastRequestAt: Date | null;
  revokedAt: Date | null;
}

/**
 * TopOperatorStat — a per-operator cumulative request aggregate, returned in
 * GlobalStats.topOperators. request_count is lifetime cumulative per key
 * (the server schema has no per-request log, so there is no time-windowed
 * breakdown).
 */
export interface TopOperatorStat {
  operatorId: string;
  operatorName: string;
  totalRequests: number;
  activeKeyCount: number;
}

/**
 * GlobalApiKeyStats — global API key statistics (super admin).
 * Mirrors the server's GlobalAPIKeyStats struct (apps/api/internal/domain/
 * api_key_entity.go). totalRequests is the sum of every key's lifetime
 * request_count; requestsByScope is that sum grouped by scope.
 */
export interface GlobalApiKeyStats {
  totalKeys: number;
  activeKeys: number;
  revokedKeys: number;
  totalRequests: number;
  requestsByScope: Record<string, number>;
  topOperators: TopOperatorStat[];
  totalOperators: number;
  maxPerMonth: number;
}

/**
 * OperatorApiKeyStats — per-operator API key statistics (super admin).
 * Mirrors the server's GetOperatorStats handler response.
 */
export interface OperatorApiKeyStats {
  operatorId: string;
  totalKeys: number;
  activeKeys: number;
  revokedKeys: number;
  keysCreatedThisMonth: number;
  monthlyLimit: number;
}

export interface AdminApiKeyListResult {
  keys: AdminApiKey[];
  pagination: Pagination;
}

// ---- Raw (snake_case) shapes mirroring the server JSON ----

export interface RawAdminApiKey {
  id: string;
  operator_id?: string;
  operator_name?: string;
  name: string;
  key_prefix: string;
  scope: string;
  is_active: boolean;
  request_count: number;
  created_at: string;
  updated_at: string;
  expires_at: string | null;
  last_request_at: string | null;
  revoked_at: string | null;
}

export interface RawAdminApiKeyListResult {
  keys: RawAdminApiKey[];
  pagination: RawPagination;
}

export interface RawTopOperatorStat {
  operator_id: string;
  operator_name: string;
  total_requests: number;
  active_key_count: number;
}

export interface RawGlobalApiKeyStats {
  total_keys: number;
  active_keys: number;
  revoked_keys: number;
  total_requests: number;
  requests_by_scope: Record<string, number>;
  top_operators: RawTopOperatorStat[];
  total_operators: number;
  max_per_month: number;
}

export interface RawOperatorApiKeyStats {
  operator_id: string;
  total_keys: number;
  active_keys: number;
  revoked_keys: number;
  keys_created_this_month: number;
  monthly_limit: number;
}
