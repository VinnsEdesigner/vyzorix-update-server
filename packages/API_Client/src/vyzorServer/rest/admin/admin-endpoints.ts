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
} from "@/domain/admin";
import type {
  ManagedOperatorListResponse,
  ManagedOperator,
  CreateOperatorRequest,
  CreateOperatorResponse,
  UpdateOperatorRequest,
  UpdateOperatorResponse,
  DeleteOperatorResponse,
} from "@/domain/admin";
import type { OperatorRole } from "@/domain/auth";

const PATHS = {
  operators: "/v1/admin/operators",
  operator: (id: string) => `/v1/admin/operators/${id}`,
  unlock: (operatorId: string) => `/v1/admin/lockout/unlock/${operatorId}`,
  keys: "/v1/admin/keys",
  operatorKeys: (operatorId: string) => `/v1/admin/keys/operator/${operatorId}`,
  key: (keyId: string) => `/v1/admin/keys/${keyId}`,
  keyStats: "/v1/admin/keys/stats",
  operatorKeyStats: (operatorId: string) => `/v1/admin/keys/stats/operator/${operatorId}`,
} as const;

export interface AdminKeyInfo {
  id: string;
  name: string;
  keyId: string;
  operatorId: string;
  operatorEmail: string;
  scope: string;
  createdAt: number;
  expiresAt?: number;
  lastUsedAt?: number;
  revokedAt?: number;
}

export interface AdminKeyStats {
  totalKeys: number;
  activeKeys: number;
  revokedKeys: number;
  keysByScope: Record<string, number>;
}

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

  async listAllKeys(): Promise<{ keys: AdminKeyInfo[]; total: number }> {
    return restClient.get<{ keys: AdminKeyInfo[]; total: number }>(PATHS.keys);
  },

  async getOperatorKeys(operatorId: string): Promise<{ keys: AdminKeyInfo[]; total: number }> {
    return restClient.get<{ keys: AdminKeyInfo[]; total: number }>(PATHS.operatorKeys(operatorId));
  },

  async forceRevokeKey(keyId: string): Promise<{ success: boolean; message: string }> {
    return restClient.delete<{ success: boolean; message: string }>(PATHS.key(keyId));
  },

  async getGlobalKeyStats(): Promise<AdminKeyStats> {
    return restClient.get<AdminKeyStats>(PATHS.keyStats);
  },

  async getOperatorKeyStats(operatorId: string): Promise<AdminKeyStats> {
    return restClient.get<AdminKeyStats>(PATHS.operatorKeyStats(operatorId));
  },
};
