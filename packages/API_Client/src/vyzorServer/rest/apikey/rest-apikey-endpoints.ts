import { restClient } from "../_shared/rest-client";
import {
  apiKeyFromRaw,
  apiKeyWithSecretFromRaw,
  paginationFromRaw,
  apiKeyStatsFromRaw,
  type RawApiKey,
  type RawApiKeyWithSecret,
  type RawApiKeyListResult,
} from "@/domain/apikey";
import type {
  ApiKey,
  ApiKeyWithSecret,
  ApiKeyListResult,
  ApiKeyScope,
} from "@/domain/apikey";

const PATHS = {
  list: "/v1/auth/api-keys",
  key: (keyId: string) => `/v1/auth/api-keys/${keyId}`,
  rotate: (keyId: string) => `/v1/auth/api-keys/${keyId}/rotate`,
} as const;

export interface CreateApiKeyInput {
  name: string;
  scope: ApiKeyScope;
  expiresInDays?: number;
}

export interface UpdateApiKeyInput {
  name?: string;
  scope?: ApiKeyScope;
}

export const apiKeys = {
  async list(params?: { page?: number; limit?: number }): Promise<ApiKeyListResult> {
    const response = await restClient.get<RawApiKeyListResult>(PATHS.list, {
      params: {
        page: params?.page,
        limit: params?.limit,
      },
    });
    return {
      keys: response.keys.map(apiKeyFromRaw),
      pagination: paginationFromRaw(response.pagination),
      stats: apiKeyStatsFromRaw(response),
    };
  },

  async get(keyId: string): Promise<ApiKey> {
    const response = await restClient.get<RawApiKey>(PATHS.key(keyId));
    return apiKeyFromRaw(response);
  },

  async create(input: CreateApiKeyInput): Promise<ApiKeyWithSecret> {
    const response = await restClient.post<RawApiKeyWithSecret>(PATHS.list, {
      name: input.name,
      scope: input.scope,
      expires_in_days: input.expiresInDays,
    });
    return apiKeyWithSecretFromRaw(response);
  },

  async update(keyId: string, input: UpdateApiKeyInput): Promise<ApiKey> {
    const response = await restClient.patch<RawApiKey>(PATHS.key(keyId), {
      name: input.name,
      scope: input.scope,
    });
    return apiKeyFromRaw(response);
  },

  async revoke(keyId: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(PATHS.key(keyId));
  },

  async rotate(keyId: string): Promise<ApiKeyWithSecret> {
    const response = await restClient.post<RawApiKeyWithSecret>(PATHS.rotate(keyId), {});
    return apiKeyWithSecretFromRaw(response);
  },
};
