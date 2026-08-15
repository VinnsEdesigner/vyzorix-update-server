import { restClient, getOrganizationContext } from "../_shared/rest-client";
import type {
  ApiKey,
  ApiKeyWithSecret,
  ApiKeyListResult,
  ApiKeyScope,
  RawApiKey,
  RawApiKeyWithSecret,
  RawApiKeyListResult,
} from "../../../domain/apikey";
import {
  apiKeyFromRaw,
  apiKeyWithSecretFromRaw,
  paginationFromRaw,
  apiKeyStatsFromRaw,
} from "../../../domain/apikey";

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
  async list(params?: { page?: number; limit?: number; organizationId?: string }): Promise<ApiKeyListResult> {
    const response = await restClient.get<RawApiKeyListResult>(PATHS.list, {
      params: {
        page: params?.page,
        limit: params?.limit,
        organization_id: params?.organizationId || getOrganizationContext(),
      },
    });
    return {
      keys: response.keys.map(apiKeyFromRaw),
      pagination: paginationFromRaw(response.pagination),
      stats: apiKeyStatsFromRaw(response),
    };
  },

  async get(keyId: string, organizationId?: string): Promise<ApiKey> {
    const response = await restClient.get<RawApiKey>(PATHS.key(keyId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return apiKeyFromRaw(response);
  },

  async create(input: CreateApiKeyInput, organizationId?: string): Promise<ApiKeyWithSecret> {
    const response = await restClient.post<RawApiKeyWithSecret>(PATHS.list, {
      name: input.name,
      scope: input.scope,
      expires_in_days: input.expiresInDays,
    }, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return apiKeyWithSecretFromRaw(response);
  },

  async update(keyId: string, input: UpdateApiKeyInput, organizationId?: string): Promise<ApiKey> {
    const response = await restClient.patch<RawApiKey>(PATHS.key(keyId), {
      name: input.name,
      scope: input.scope,
    }, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return apiKeyFromRaw(response);
  },

  async revoke(keyId: string, organizationId?: string): Promise<{ success: boolean }> {
    return restClient.delete<{ success: boolean }>(PATHS.key(keyId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async rotate(keyId: string, organizationId?: string): Promise<ApiKeyWithSecret> {
    const response = await restClient.post<RawApiKeyWithSecret>(PATHS.rotate(keyId), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return apiKeyWithSecretFromRaw(response);
  },
};
