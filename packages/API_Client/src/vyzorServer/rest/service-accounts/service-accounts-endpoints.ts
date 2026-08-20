import { restClient, getOrganizationContext } from "../_shared/rest-client";
import {
  serviceAccountFromRaw,
  serviceAccountsFromRaw,
  serviceAccountTokenFromRaw,
  serviceAccountTokensFromRaw,
  type ServiceAccount,
  type ServiceAccountToken,
  type CreateServiceAccountTokenRequest,
} from "../../../domain/service-accounts";

const PATHS = {
  accounts: "/v1/service-accounts",
  account: (id: string) => `/v1/service-accounts/${id}`,
  tokens: (id: string) => `/v1/service-accounts/${id}/tokens`,
  token: (id: string, tokenId: string) => `/v1/service-accounts/${id}/tokens/${tokenId}`,
  rotateToken: (id: string, tokenId: string) => `/v1/service-accounts/${id}/tokens/${tokenId}/rotate`,
} as const;

export const serviceAccounts = {
  async list(organizationId?: string): Promise<ServiceAccount[]> {
    const response = await restClient.get<{ service_accounts: Parameters<typeof serviceAccountsFromRaw>[0] }>(
      PATHS.accounts,
      { params: { organization_id: organizationId || getOrganizationContext() } },
    );
    return serviceAccountsFromRaw(response.service_accounts);
  },

  async create(name: string, organizationId?: string): Promise<ServiceAccount> {
    const response = await restClient.post<Parameters<typeof serviceAccountFromRaw>[0]>(
      PATHS.accounts,
      { name },
      { params: { organization_id: organizationId || getOrganizationContext() } },
    );
    return serviceAccountFromRaw(response);
  },

  async delete(id: string, organizationId?: string): Promise<void> {
    await restClient.delete(PATHS.account(id), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async listTokens(serviceId: string, organizationId?: string): Promise<ServiceAccountToken[]> {
    const response = await restClient.get<{ tokens: Parameters<typeof serviceAccountTokensFromRaw>[0] }>(
      PATHS.tokens(serviceId),
      { params: { organization_id: organizationId || getOrganizationContext() } },
    );
    return serviceAccountTokensFromRaw(response.tokens);
  },

  async createToken(
    serviceId: string,
    req: CreateServiceAccountTokenRequest,
    organizationId?: string,
  ): Promise<ServiceAccountToken> {
    const response = await restClient.post<Parameters<typeof serviceAccountTokenFromRaw>[0]>(
      PATHS.tokens(serviceId),
      req,
      { params: { organization_id: organizationId || getOrganizationContext() } },
    );
    return serviceAccountTokenFromRaw(response);
  },

  async rotateToken(serviceId: string, tokenId: string, organizationId?: string): Promise<ServiceAccountToken> {
    const response = await restClient.post<Parameters<typeof serviceAccountTokenFromRaw>[0]>(
      PATHS.rotateToken(serviceId, tokenId),
      {},
      { params: { organization_id: organizationId || getOrganizationContext() } },
    );
    return serviceAccountTokenFromRaw(response);
  },

  async revokeToken(serviceId: string, tokenId: string, organizationId?: string): Promise<void> {
    await restClient.delete(PATHS.token(serviceId, tokenId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },
};
