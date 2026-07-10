import type { ApiKey, ApiKeyWithSecret, ApiKeyStats, Pagination } from "./apikey-entity";

export type RawApiKeyResponse = {
  id: string;
  operator_id: string;
  name: string;
  key_prefix: string;
  scope: "read" | "write" | "admin";
  expires_at: string | null;
  last_request_at: string | null;
  is_active: boolean;
  request_count: number;
  created_at: string;
  updated_at: string;
  revoked_at: string | null;
};

export type RawApiKeyWithFullResponse = RawApiKeyResponse & {
  api_key: string;
};

export type RawPagination = {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
};

export type RawListResponse = {
  keys: RawApiKeyResponse[];
  pagination: RawPagination;
  monthly_limit: number;
  keys_created_this_month: number;
};

export function apiKeyFromRaw(raw: RawApiKeyResponse): ApiKey {
  return {
    id: raw.id,
    operatorId: raw.operator_id,
    name: raw.name,
    keyPrefix: raw.key_prefix,
    scope: raw.scope,
    expiresAt: raw.expires_at ? new Date(raw.expires_at) : null,
    isActive: raw.is_active,
    requestCount: raw.request_count,
    lastRequestAt: raw.last_request_at ? new Date(raw.last_request_at) : null,
    createdAt: new Date(raw.created_at),
    updatedAt: new Date(raw.updated_at),
    revokedAt: raw.revoked_at ? new Date(raw.revoked_at) : null,
  };
}

export function apiKeyWithSecretFromRaw(raw: RawApiKeyWithFullResponse): ApiKeyWithSecret {
  return {
    ...apiKeyFromRaw(raw),
    apiKey: raw.api_key,
  };
}

export function paginationFromRaw(raw: RawPagination): Pagination {
  return {
    page: raw.page,
    limit: raw.limit,
    total: raw.total,
    totalPages: raw.total_pages,
  };
}

export function apiKeyStatsFromRaw(raw: { monthly_limit: number; keys_created_this_month: number }): ApiKeyStats {
  return {
    monthlyLimit: raw.monthly_limit,
    keysCreatedThisMonth: raw.keys_created_this_month,
  };
}

export function apiKeyToResponse(key: ApiKey): RawApiKeyResponse {
  return {
    id: key.id,
    operator_id: key.operatorId,
    name: key.name,
    key_prefix: key.keyPrefix,
    scope: key.scope,
    expires_at: key.expiresAt?.toISOString() ?? null,
    is_active: key.isActive,
    request_count: key.requestCount,
    last_request_at: key.lastRequestAt?.toISOString() ?? null,
    created_at: key.createdAt.toISOString(),
    updated_at: key.updatedAt.toISOString(),
    revoked_at: key.revokedAt?.toISOString() ?? null,
  };
}
