import type { OperatorRole } from "../auth/auth-entity";

export interface ManagedOperator {
  id: string;
  email: string;
  name: string;
  role: OperatorRole;
  mfaEnabled: boolean;
  emailVerified: boolean;
  createdAt: Date;
  updatedAt?: Date;
}

export interface ManagedOperatorListResponse {
  operators: ManagedOperator[];
  total: number;
}

export interface CreateOperatorRequest {
  email: string;
  password: string;
  name: string;
  role?: OperatorRole;
}

export interface UpdateOperatorRequest {
  name?: string;
  role?: OperatorRole;
  email?: string;
}

export interface CreateOperatorResponse {
  id: string;
  email: string;
  name: string;
  role: OperatorRole;
  createdAt: number;
}

export interface UpdateOperatorResponse {
  id: string;
  email: string;
  name: string;
  role: OperatorRole;
  updatedAt: number;
}

export interface DeleteOperatorResponse {
  success: boolean;
  message: string;
}
