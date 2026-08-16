import type {
  ApiKey,
  ApiKeyWithSecret,
  ApiKeyStats,
  ApiKeyScope,
} from "./apikey-entity";
import type { RawPagination } from "../_shared";

export interface RawApiKey {
  id: string;
  // Only present in create/rotate responses; list/get/patch omit it.
  operator_id?: string;
  name: string;
  key_prefix: string;
  scope: ApiKeyScope;
  // Server serializes time.Time/*time.Time as RFC3339 ISO strings.
  expires_at: string | null;
  last_request_at: string | null;
  is_active: boolean;
  request_count: number;
  created_at: string;
  updated_at: string;
  revoked_at: string | null;
}

export interface RawApiKeyWithSecret extends RawApiKey {
  api_key: string;
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

export function apiKeyStatsFromRaw(raw: { monthly_limit: number; keys_created_this_month: number }): ApiKeyStats {
  return {
    monthlyLimit: raw.monthly_limit,
    keysCreatedThisMonth: raw.keys_created_this_month,
  };
}

const VALID_SCOPES: readonly ApiKeyScope[] = ["read", "write", "admin"];

/**
 * Normalizes a raw scope string to a valid ApiKeyScope. Case-insensitive;
 * unknown/empty values default to "read" so a malformed server response can
 * never produce an invalid scope downstream.
 */
export function parseScope(raw: string | undefined | null): ApiKeyScope {
  if (!raw) return "read";
  const lower = raw.toLowerCase();
  return (VALID_SCOPES as readonly string[]).includes(lower) ? (lower as ApiKeyScope) : "read";
}
