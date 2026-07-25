import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  clientCredentialFromRaw,
  clientCredentialWithSecretFromRaw,
  clientCredentialListFromRaw,
  type RawClientCredentialResponse,
  type RawClientCredentialListResponse,
} from "@/domain/clientcredentials";
import type {
  ClientCredential,
  ClientCredentialWithSecret,
  CreateClientCredentialRequest,
  UpdateClientCredentialRequest,
} from "@/domain/clientcredentials";

const PATHS = {
  list: "/v1/auth/client-credentials",
  client: (clientId: string) => `/v1/auth/client-credentials/${clientId}`,
  rotateSecret: (clientId: string) => `/v1/auth/client-credentials/${clientId}/rotate-secret`,
} as const;

export const clientCredentials = {
  async create(request: CreateClientCredentialRequest, organizationId?: string): Promise<ClientCredentialWithSecret> {
    const response = await restClient.post<{
      clientId: string;
      clientSecret: string;
      platform: string;
      name: string;
      createdAt: number;
    }>(PATHS.list, {
      name: request.name,
      platform: request.platform,
      allowedOrigins: request.allowedOrigins,
      allowedPaths: request.allowedPaths,
      rateLimit: request.rateLimit,
    }, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return clientCredentialWithSecretFromRaw(response, request.rateLimit);
  },

  async list(organizationId?: string): Promise<{ clients: ClientCredential[] }> {
    const response = await restClient.get<RawClientCredentialListResponse>(PATHS.list, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return clientCredentialListFromRaw(response);
  },

  async get(clientId: string, organizationId?: string): Promise<ClientCredential> {
    const response = await restClient.get<RawClientCredentialResponse>(PATHS.client(clientId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return clientCredentialFromRaw(response.client);
  },

  async update(clientId: string, request: UpdateClientCredentialRequest, organizationId?: string): Promise<ClientCredential> {
    const response = await restClient.patch<RawClientCredentialResponse>(PATHS.client(clientId), {
      name: request.name,
      allowedOrigins: request.allowedOrigins,
      allowedPaths: request.allowedPaths,
      rateLimit: request.rateLimit,
      active: request.active,
    }, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
    return clientCredentialFromRaw(response.client);
  },

  async delete(clientId: string, organizationId?: string): Promise<{ success: boolean; clientId: string }> {
    return restClient.delete<{ success: boolean; clientId: string }>(PATHS.client(clientId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async rotateSecret(clientId: string, organizationId?: string): Promise<{ success: boolean; message: string }> {
    return restClient.post<{ success: boolean; message: string }>(PATHS.rotateSecret(clientId), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },
};
