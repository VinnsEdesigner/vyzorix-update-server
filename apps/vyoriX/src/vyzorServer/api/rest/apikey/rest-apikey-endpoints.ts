import { apiGet, apiPost, apiPatch, apiDelete } from "../_shared/rest-client";
import type { ApiKey, ApiKeyWithSecret, ApiKeyListResponse } from "@/domain/apikey";
import type { ApiKeyScope } from "@/domain/apikey";
import { apiKeyFromRaw, apiKeyWithSecretFromRaw, paginationFromRaw, apiKeyStatsFromRaw } from "@/domain/apikey";

export const API_KEY_PATHS = {
  list: "/v1/auth/api-keys",
  get: (keyId: string) => `/v1/auth/api-keys/${keyId}`,
  create: "/v1/auth/api-keys",
  update: (keyId: string) => `/v1/auth/api-keys/${keyId}`,
  revoke: (keyId: string) => `/v1/auth/api-keys/${keyId}`,
  rotate: (keyId: string) => `/v1/auth/api-keys/${keyId}/rotate`,
} as const;

export interface CreateApiKeyRequest {
  name: string;
  scope: ApiKeyScope;
  expires_in_days?: number;
}

export interface UpdateApiKeyRequest {
  name?: string;
  scope?: ApiKeyScope;
}

export async function listApiKeys(params?: { page?: number; limit?: number }): Promise<ApiKeyListResponse> {
  const data = await apiGet<{
    keys: Parameters<typeof apiKeyFromRaw>[0][];
    pagination: Parameters<typeof paginationFromRaw>[0];
    monthly_limit: number;
    keys_created_this_month: number;
  }>(API_KEY_PATHS.list, {
    page: params?.page,
    limit: params?.limit,
  });

  return {
    keys: data.keys.map(apiKeyFromRaw),
    pagination: paginationFromRaw(data.pagination),
    ...apiKeyStatsFromRaw(data),
  };
}

export async function getApiKey(keyId: string): Promise<ApiKey> {
  const data = await apiGet<Parameters<typeof apiKeyFromRaw>[0]>(API_KEY_PATHS.get(keyId));
  return apiKeyFromRaw(data);
}

export async function createApiKey(input: CreateApiKeyRequest): Promise<{ success: boolean; key: ApiKeyWithSecret }> {
  const response = await apiPost<{
    success: boolean;
    key?: Parameters<typeof apiKeyWithSecretFromRaw>[0];
    error?: string;
  }>(API_KEY_PATHS.create, input);

  if (!response.success || !response.key) {
    throw new Error(response.error || "Failed to create API key");
  }

  return {
    success: true,
    key: apiKeyWithSecretFromRaw(response.key),
  };
}

export async function updateApiKey(keyId: string, input: UpdateApiKeyRequest): Promise<{ success: boolean; key: ApiKey }> {
  const response = await apiPatch<{
    success: boolean;
    key?: Parameters<typeof apiKeyFromRaw>[0];
    error?: string;
  }>(API_KEY_PATHS.update(keyId), input);

  if (!response.success || !response.key) {
    throw new Error(response.error || "Failed to update API key");
  }

  return {
    success: true,
    key: apiKeyFromRaw(response.key),
  };
}

export async function revokeApiKey(keyId: string): Promise<{ success: boolean }> {
  const response = await apiDelete<{ success: boolean; error?: string }>(API_KEY_PATHS.revoke(keyId));

  if (!response.success) {
    throw new Error(response.error || "Failed to revoke API key");
  }

  return { success: true };
}

export async function rotateApiKey(keyId: string): Promise<{ success: boolean; key: ApiKeyWithSecret }> {
  const response = await apiPost<{
    success: boolean;
    key?: Parameters<typeof apiKeyWithSecretFromRaw>[0];
    error?: string;
  }>(API_KEY_PATHS.rotate(keyId), {});

  if (!response.success || !response.key) {
    throw new Error(response.error || "Failed to rotate API key");
  }

  return {
    success: true,
    key: apiKeyWithSecretFromRaw(response.key),
  };
}
