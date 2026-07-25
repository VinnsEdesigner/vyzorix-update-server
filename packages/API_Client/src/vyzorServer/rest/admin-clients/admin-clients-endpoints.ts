
import { restClient, getOrganizationContext } from "../_shared/rest-client";

export interface AdminClient {
  id: string;
  name: string;
  clientId: string;
  platform: string;
  allowedOrigins: string[];
  allowedPaths: string[];
  rateLimit: number;
  active: boolean;
  createdAt: number;
  lastUsedAt?: number;
  keyVersion: number;
}

export interface AdminClientListResponse {
  clients: AdminClient[];
  total: number;
}

export interface UpdateClientRequest {
  name?: string;
  allowedOrigins?: string[];
  allowedPaths?: string[];
  rateLimit?: number;
  active?: boolean;
}

const PATHS = {
  clients: "/v1/admin/clients",
  client: (clientId: string) => `/v1/admin/clients/${clientId}`,
  rotateKey: (clientId: string) => `/v1/admin/clients/${clientId}/rotate-key`,
} as const;

export const adminClients = {
  async list(organizationId?: string): Promise<AdminClientListResponse> {
    return restClient.get<AdminClientListResponse>(PATHS.clients, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async get(clientId: string, organizationId?: string): Promise<{ client: AdminClient }> {
    return restClient.get<{ client: AdminClient }>(PATHS.client(clientId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async update(clientId: string, request: UpdateClientRequest, organizationId?: string): Promise<{ client: AdminClient }> {
    return restClient.patch<{ client: AdminClient }>(PATHS.client(clientId), request, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async delete(clientId: string, organizationId?: string): Promise<{ success: boolean; clientId: string }> {
    return restClient.delete<{ success: boolean; clientId: string }>(PATHS.client(clientId), {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },

  async rotateKey(clientId: string, organizationId?: string): Promise<{ 
    success: boolean; 
    message: string; 
    clientId: string; 
    keyVersion: number 
  }> {
    return restClient.post<{
      success: boolean;
      message: string;
      clientId: string;
      keyVersion: number;
    }>(PATHS.rotateKey(clientId), {}, {
      params: { organization_id: organizationId || getOrganizationContext() },
    });
  },
};
