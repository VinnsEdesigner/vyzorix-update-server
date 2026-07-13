import type {
  ManagedOperator,
  ManagedOperatorListResponse,
  CreateOperatorResponse,
  UpdateOperatorResponse,
} from "./admin-entity";
import type { OperatorRole } from "../auth/auth-entity";

export interface RawManagedOperator {
  id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
  email_verified: boolean;
  created_at: number;
  updated_at?: number;
}

export interface RawManagedOperatorListResponse {
  operators: RawManagedOperator[];
  total: number;
}

export interface RawCreateOperatorResponse {
  id: string;
  email: string;
  name: string;
  role: string;
  created_at: number;
}

export interface RawUpdateOperatorResponse {
  id: string;
  email: string;
  name: string;
  role: string;
  updated_at: number;
}

export function managedOperatorFromRaw(raw: RawManagedOperator): ManagedOperator {
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    role: raw.role as OperatorRole,
    mfaEnabled: raw.mfa_enabled,
    emailVerified: raw.email_verified,
    createdAt: new Date(raw.created_at),
    updatedAt: raw.updated_at ? new Date(raw.updated_at) : undefined,
  };
}

export function managedOperatorListFromRaw(raw: RawManagedOperatorListResponse): ManagedOperatorListResponse {
  return {
    operators: raw.operators.map(managedOperatorFromRaw),
    total: raw.total,
  };
}

export function createOperatorResponseFromRaw(raw: RawCreateOperatorResponse): CreateOperatorResponse {
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    role: raw.role as OperatorRole,
    createdAt: raw.created_at,
  };
}

export function updateOperatorResponseFromRaw(raw: RawUpdateOperatorResponse): UpdateOperatorResponse {
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    role: raw.role as OperatorRole,
    updatedAt: raw.updated_at,
  };
}

export interface RawCreateOperatorRequest {
  email: string;
  password: string;
  name: string;
  role?: string;
}

export interface RawUpdateOperatorRequest {
  name?: string;
  role?: string;
  email?: string;
}

export function createOperatorRequestToRaw(request: {
  email: string;
  password: string;
  name: string;
  role?: OperatorRole;
}): RawCreateOperatorRequest {
  return {
    email: request.email.toLowerCase().trim(),
    password: request.password,
    name: request.name.trim(),
    role: request.role,
  };
}

export function updateOperatorRequestToRaw(request: {
  name?: string;
  role?: OperatorRole;
  email?: string;
}): RawUpdateOperatorRequest {
  return {
    name: request.name?.trim(),
    role: request.role,
    email: request.email?.toLowerCase().trim(),
  };
}
