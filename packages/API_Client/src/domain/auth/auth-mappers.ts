import type {
  Operator,
  LoginResponse,
  LoginWithTokensResponse,
  LoginWithTokensMFARequiredResponse,
  RegisterResponse,
  MeResponse,
  OperatorRole,
  AuthTokens,
  OrganizationInfo,
} from "./auth-entity";
import type { Thresholds, ClientSettings } from "../settings/settings-entity";

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
  organizations?: OrganizationInfo[];
  selected_organization?: OrganizationInfo;
  last_organization_id?: string;
  needs_organization?: boolean;
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
  mfa_enabled: boolean;
  email_verified: boolean;
  thresholds?: Thresholds;
  client?: ClientSettings;
  needs_organization: boolean;
  organizations: OrganizationInfo[];
  last_organization_id?: string;
  selected_organization?: OrganizationInfo;
}

interface RawRefreshResponse {
  access_token: string;
  refresh_token: string;
  expires_at: number;
  session_id: string;
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
    operator_id: raw.operator_id,
    email: raw.email,
    name: raw.name,
    role: raw.role,
    mfa_enabled: raw.mfa_enabled,
    access_token: raw.access_token,
    refresh_token: raw.refresh_token,
    expires_at: raw.expires_at,
    session_id: raw.session_id,
    needs_organization: raw.needs_organization ?? false,
    organizations: raw.organizations ?? [],
    last_organization_id: raw.last_organization_id,
    selected_organization: raw.selected_organization,
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
    mfa_enabled: raw.mfa_enabled,
    email_verified: raw.email_verified,
    thresholds: raw.thresholds,
    client: raw.client,
    needs_organization: raw.needs_organization,
    organizations: raw.organizations,
    last_organization_id: raw.last_organization_id,
    selected_organization: raw.selected_organization,
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

export function loginRequestToRaw(request: { email: string; password: string }): { email: string; password: string } {
  return {
    email: request.email.toLowerCase().trim(),
    password: request.password,
  };
}

export function registerRequestToRaw(request: { email: string; password: string; name: string }): { email: string; password: string; name: string } {
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

export function resetPasswordRequestToRaw(token: string, newPassword: string): { token: string; new_password: string } {
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

export function mfaStatusResponseFromRaw(raw: { enabled: boolean; backup_codes?: string[] }): { enabled: boolean; backupCodes?: string[] } {
  return {
    enabled: raw.enabled,
    backupCodes: raw.backup_codes,
  };
}

export function mfaEnrollResponseFromRaw(raw: { secret: string; qr_code_url: string }): { secret: string; qrCodeUrl: string } {
  return {
    secret: raw.secret,
    qrCodeUrl: raw.qr_code_url,
  };
}

export function mfaVerifyRequestToRaw(operatorId: string, code: string): { operator_id: string; code: string } {
  return { operator_id: operatorId, code };
}

interface RawMFAVerifyResponse {
  success: boolean;
  session_id?: string;
  access_token?: string;
  refresh_token?: string;
  expires_at?: number;
}

export function mfaVerifyResponseFromRaw(raw: RawMFAVerifyResponse): {
  success: boolean;
  sessionId?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: number;
} {
  return {
    success: raw.success,
    sessionId: raw.session_id,
    accessToken: raw.access_token,
    refreshToken: raw.refresh_token,
    expiresAt: raw.expires_at,
  };
}
