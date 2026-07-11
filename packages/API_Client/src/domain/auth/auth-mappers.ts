import type {
  Operator,
  LoginResponse,
  RegisterResponse,
  MeResponse,
  OperatorRole,
  AuthTokens,
} from "./auth-entity";
import type { Thresholds, ClientSettings } from "../settings/settings-entity";

interface RawOperator {
  id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
  email_verified: boolean;
}

interface RawLoginResponse {
  operator_id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
}

interface RawLoginMFARequiredResponse {
  mfa_required: true;
  operator_id: string;
}

interface RawRegisterResponse {
  operator_id: string;
  email: string;
  name: string;
}

interface RawMeResponse {
  id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
  email_verified: boolean;
  thresholds?: Thresholds;
  client?: ClientSettings;
}

interface RawRefreshResponse {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  session_id: string;
}

export function operatorFromRaw(raw: RawOperator): Operator {
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    role: raw.role as OperatorRole,
    mfaEnabled: raw.mfa_enabled,
    emailVerified: raw.email_verified,
  };
}

export function loginResponseFromRaw(raw: RawLoginResponse): LoginResponse {
  return {
    operatorId: raw.operator_id,
    email: raw.email,
    name: raw.name,
    role: raw.role as OperatorRole,
    mfaEnabled: raw.mfa_enabled,
  };
}

export function isMFAResponse(
  raw: RawLoginResponse | RawLoginMFARequiredResponse
): raw is RawLoginMFARequiredResponse {
  return "mfa_required" in raw && raw.mfa_required === true;
}

export function registerResponseFromRaw(raw: RawRegisterResponse): RegisterResponse {
  return {
    operatorId: raw.operator_id,
    email: raw.email,
    name: raw.name,
  };
}

export function meResponseFromRaw(raw: RawMeResponse): MeResponse {
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    role: raw.role as OperatorRole,
    mfaEnabled: raw.mfa_enabled,
    emailVerified: raw.email_verified,
    thresholds: raw.thresholds,
    client: raw.client,
  };
}

export function authTokensFromRaw(raw: RawRefreshResponse): AuthTokens {
  return {
    accessToken: raw.access_token,
    refreshToken: raw.refresh_token,
    expiresAt: raw.expires_at,
    sessionId: raw.session_id,
  };
}

export function loginRequestToRaw(request: {
  email: string;
  password: string;
}): { email: string; password: string } {
  return {
    email: request.email.toLowerCase().trim(),
    password: request.password,
  };
}

export function registerRequestToRaw(request: {
  email: string;
  password: string;
  name: string;
}): { email: string; password: string; name: string } {
  return {
    email: request.email.toLowerCase().trim(),
    password: request.password,
    name: request.name.trim(),
  };
}

export function refreshRequestToRaw(refreshToken: string): { refresh_token: string } {
  return { refresh_token: refreshToken };
}

export function updateNameRequestToRaw(name: string): { name: string } {
  return { name: name.trim() };
}
