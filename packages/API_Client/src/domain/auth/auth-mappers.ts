import type {
  Operator,
  LoginResponse,
  LoginWithTokensResponse,
  LoginWithTokensMFARequiredResponse,
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

interface RawLoginWithTokensResponse {
  operator_id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
  access_token: string;
  refresh_token: string;
  expires_at: number;
  session_id: string;
}

interface RawLoginWithTokensMFARequiredResponse {
  mfa_required: true;
  operator_id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
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

export function loginWithTokensResponseFromRaw(raw: RawLoginWithTokensResponse): LoginWithTokensResponse {
  return {
    operatorId: raw.operator_id,
    email: raw.email,
    name: raw.name,
    role: raw.role as OperatorRole,
    mfaEnabled: raw.mfa_enabled,
    accessToken: raw.access_token,
    refreshToken: raw.refresh_token,
    expiresAt: raw.expires_at,
    sessionId: raw.session_id,
  };
}

export function isLoginWithTokensMFARequired(
  raw: RawLoginWithTokensResponse | RawLoginWithTokensMFARequiredResponse
): raw is RawLoginWithTokensMFARequiredResponse {
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

export function forgotPasswordRequestToRaw(email: string): { email: string } {
  return { email: email.toLowerCase().trim() };
}

export function resetPasswordRequestToRaw(token: string, newPassword: string): {
  token: string;
  new_password: string;
} {
  return { token, new_password: newPassword };
}

export function verifyEmailRequestToRaw(token: string): { token: string } {
  return { token };
}

export function mfaCodeRequestToRaw(code: string): { code: string } {
  return { code };
}

export function mfaEnrollRequestToRaw(code: string): { code: string } {
  return { code };
}

export function backupCodeVerifyRequestToRaw(code: string): { code: string } {
  return { code };
}

export function mfaStatusResponseFromRaw(raw: { enabled: boolean; backup_codes?: string[] }): {
  enabled: boolean;
  backupCodes?: string[];
} {
  return {
    enabled: raw.enabled,
    backupCodes: raw.backup_codes,
  };
}

export function mfaEnrollResponseFromRaw(raw: { secret: string; qr_code_url: string }): {
  secret: string;
  qrCodeUrl: string;
} {
  return {
    secret: raw.secret,
    qrCodeUrl: raw.qr_code_url,
  };
}

export function mfaVerifyRequestToRaw(operatorId: string, code: string): { operator_id: string; code: string } {
  return { operator_id: operatorId, code };
}

export interface RawMFAVerifyResponse {
  success: boolean;
  session_id?: string;
  access_token?: string;
  refresh_token?: string;
  expires_at?: number;
  operator?: {
    id: string;
    email: string;
    name: string;
    role: string;
    mfa_enabled: boolean;
  };
}

export function mfaVerifyResponseFromRaw(raw: RawMFAVerifyResponse): {
  success: boolean;
  sessionId?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: number;
  operator?: Operator;
} {
  return {
    success: raw.success,
    sessionId: raw.session_id,
    accessToken: raw.access_token,
    refreshToken: raw.refresh_token,
    expiresAt: raw.expires_at,
    operator: raw.operator ? operatorFromRaw(raw.operator as RawOperator) : undefined,
  };
}
