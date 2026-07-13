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
  operators: "/v1/auth/admin/operators",
  operator: (id: string) => `/v1/auth/admin/operators/${id}`,
} as const;

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
};
