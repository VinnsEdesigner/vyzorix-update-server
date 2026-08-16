import { restClient } from "../_shared/rest-client";
import {
  managedOperatorListFromRaw,
  managedOperatorFromRaw,
  createOperatorResponseFromRaw,
  updateOperatorResponseFromRaw,
  createOperatorRequestToRaw,
  updateOperatorRequestToRaw,
  type RawManagedOperator,
  type RawCreateOperatorResponse,
  type RawUpdateOperatorResponse,
  type ManagedOperatorListResponse,
  type ManagedOperator,
  type CreateOperatorRequest,
  type CreateOperatorResponse,
  type UpdateOperatorRequest,
  type UpdateOperatorResponse,
  type DeleteOperatorResponse,
} from "../../../domain/admin";
import {
  adminApiKeyListFromRaw,
  globalApiKeyStatsFromRaw,
  operatorApiKeyStatsFromRaw,
  type AdminApiKeyListResult,
  type GlobalApiKeyStats,
  type OperatorApiKeyStats,
  type RawAdminApiKeyListResult,
  type RawGlobalApiKeyStats,
  type RawOperatorApiKeyStats,
} from "../../../domain/apikey";

const PATHS = {
  operators: "/v1/admin/operators",
  operator: (id: string) => `/v1/admin/operators/${id}`,
  unlock: (operatorId: string) => `/v1/admin/lockout/unlock/${operatorId}`,
  // Server registers these under /v1/admin/api-keys (see server_routes.go setupAdminRoutes).
  keys: "/v1/admin/api-keys",
  operatorKeys: (operatorId: string) => `/v1/admin/api-keys/operator/${operatorId}`,
  key: (keyId: string) => `/v1/admin/api-keys/${keyId}`,
  keyStats: "/v1/admin/api-keys/stats",
  operatorKeyStats: (operatorId: string) => `/v1/admin/api-keys/stats/operator/${operatorId}`,
} as const;

// API key admin types (AdminApiKey, GlobalApiKeyStats, OperatorApiKeyStats,
// AdminApiKeyListResult + raw/mapper functions) live in domain/apikey and are
// re-exported via the domain barrel. See rest/admin/index.ts for re-exports.

export const admin = {
  async listOperators(): Promise<ManagedOperatorListResponse> {
    const response = await restClient.get<{
      operators: RawManagedOperator[];
      total: number;
    }>(PATHS.operators);
    return managedOperatorListFromRaw(response);
  },

  async getOperator(operatorId: string): Promise<ManagedOperator> {
    const response = await restClient.get<RawManagedOperator>(PATHS.operator(operatorId));
    return managedOperatorFromRaw(response);
  },

  async createOperator(request: CreateOperatorRequest): Promise<CreateOperatorResponse> {
    const response = await restClient.post<RawCreateOperatorResponse>(
      PATHS.operators,
      createOperatorRequestToRaw(request)
    );
    return createOperatorResponseFromRaw(response);
  },

  async updateOperator(operatorId: string, request: UpdateOperatorRequest): Promise<UpdateOperatorResponse> {
    const response = await restClient.patch<RawUpdateOperatorResponse>(
      PATHS.operator(operatorId),
      updateOperatorRequestToRaw(request)
    );
    return updateOperatorResponseFromRaw(response);
  },

  async deleteOperator(operatorId: string): Promise<DeleteOperatorResponse> {
    return restClient.delete<DeleteOperatorResponse>(PATHS.operator(operatorId));
  },

  async unlockAccount(operatorId: string): Promise<{ success: boolean; message: string }> {
    return restClient.post<{ success: boolean; message: string }>(PATHS.unlock(operatorId), {});
  },

  async listAllKeys(params?: {
    page?: number;
    limit?: number;
    operatorId?: string;
    search?: string;
  }): Promise<AdminApiKeyListResult> {
    const raw = await restClient.get<RawAdminApiKeyListResult>(PATHS.keys, {
      params: {
        page: params?.page,
        limit: params?.limit,
        operator_id: params?.operatorId,
        search: params?.search,
      },
    });
    return adminApiKeyListFromRaw(raw);
  },

  async getOperatorKeys(operatorId: string, page?: number, limit?: number): Promise<AdminApiKeyListResult> {
    const raw = await restClient.get<RawAdminApiKeyListResult>(PATHS.operatorKeys(operatorId), {
      params: { page, limit },
    });
    return adminApiKeyListFromRaw(raw);
  },

  // Server returns 204 No Content on successful force revocation.
  async forceRevokeKey(keyId: string): Promise<void> {
    await restClient.delete<void>(PATHS.key(keyId));
  },

  async getGlobalKeyStats(): Promise<GlobalApiKeyStats> {
    const raw = await restClient.get<RawGlobalApiKeyStats>(PATHS.keyStats);
    return globalApiKeyStatsFromRaw(raw);
  },

  async getOperatorKeyStats(operatorId: string): Promise<OperatorApiKeyStats> {
    const raw = await restClient.get<RawOperatorApiKeyStats>(PATHS.operatorKeyStats(operatorId));
    return operatorApiKeyStatsFromRaw(raw);
  },
};
