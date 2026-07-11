import type { Thresholds, ClientSettings } from "../settings/settings-entity";

export type OperatorRole = "super_admin" | "admin" | "operator" | "viewer";

export interface Operator {
  id: string;
  email: string;
  name: string;
  role: OperatorRole;
  mfaEnabled: boolean;
  emailVerified: boolean;
}

export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  sessionId: string;
}

export interface LoginResponse {
  operatorId: string;
  email: string;
  name: string;
  role: OperatorRole;
  mfaEnabled: boolean;
}

export interface LoginMFARequiredResponse {
  mfaRequired: true;
  operatorId: string;
}

export interface RegisterResponse {
  operatorId: string;
  email: string;
  name: string;
}

export interface MeResponse extends Operator {
  thresholds?: Thresholds;
  client?: ClientSettings;
}

export interface UpdateNameRequest {
  name: string;
}
