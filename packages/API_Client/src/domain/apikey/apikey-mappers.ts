import type {
  ApiKey,
  ApiKeyWithSecret,
  ApiKeyStats,
  Pagination,
  ApiKeyScope,
} from "./apikey-entity";

export interface RawApiKey {
  id: string;
  operator_id: string;
  name: string;
  key_prefix: string;
  scope: ApiKeyScope;
  expires_at: number | null;
  last_request_at: number | null;
  is_active: boolean;
  request_count: number;
  created_at: number;
  updated_at: number;
  revoked_at: number | null;
}

export interface RawApiKeyWithSecret extends RawApiKey {
  api_key: string;
}

export interface RawPagination {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

export interface RawApiKeyListResult {
  keys: RawApiKey[];
  pagination: RawPagination;
  monthly_limit: number;
  keys_created_this_month: number;
}

export function apiKeyFromRaw(raw: RawApiKey): ApiKey {
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

export function apiKeyWithSecretFromRaw(raw: RawApiKeyWithSecret): ApiKeyWithSecret {
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
