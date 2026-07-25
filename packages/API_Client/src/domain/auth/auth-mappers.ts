


import type {
  LoginResponse,
  LoginWithTokensResponse,
  RegisterResponse,
  MeResponse,
  OperatorRole,
  AuthTokens,
  OrganizationInfo,
  OrganizationMembership,
} from "./auth-entity";
import type { Thresholds, ClientSettings } from "../settings/settings-entity";

export interface RawLoginResponse {
  operator_id: string;
  email: string;
  name: string;
  role: string;
  mfa_enabled: boolean;
  selected_organization?: OrganizationInfo;
  last_organization_id?: string;
  organizations?: OrganizationInfo[];
  needs_organization?: boolean;
}

export interface RawLoginMFARequiredResponse {
  mfa_required: true;
  operator_id: string;
}

export interface RawLoginWithTokensResponse {
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
  memberships?: OrganizationMembership[];
  selected_organization?: OrganizationInfo;
  last_organization_id?: string;
  needs_organization?: boolean;
}

export interface RawLoginWithTokensMFARequiredResponse {
  mfa_required: true;
  operator_id: string;
  email: string;
  name: string;
  mfa_enabled: boolean;
}

export interface RawRegisterResponse {
  operator_id: string;
  email: string;
  name: string;
}

export interface RawMeResponse {
  id: string;
  email: string;
  name: string;
  mfa_enabled: boolean;
  email_verified: boolean;
  thresholds?: Thresholds;
  client?: ClientSettings;
  needs_organization: boolean;
  organizations: OrganizationInfo[];
  memberships?: OrganizationMembership[];
  last_organization_id?: string;
  selected_organization?: OrganizationInfo;
}

export interface RawRefreshResponse {
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
    selectedOrganization: raw.selected_organization,
    lastOrganizationId: raw.last_organization_id,
    organizations: raw.organizations ?? [],
    needsOrganization: raw.needs_organization ?? false,
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
    memberships: raw.memberships ?? [],
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
    memberships: raw.memberships ?? [],
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

export function mfaEnrollResponseFromRaw(raw: { secret: string; uri?: string }): { secret: string; uri: string } {
  return {
    secret: raw.secret,
    uri: raw.uri ?? "",
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
    mfa_enabled: boolean;
  };
}

export function mfaVerifyResponseFromRaw(raw: RawMFAVerifyResponse): {
  success: boolean;
  sessionId?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: number;
  operator?: {
    id: string;
    email: string;
    name: string;
    mfaEnabled: boolean;
  };
} {
  return {
    success: raw.success,
    sessionId: raw.session_id,
    accessToken: raw.access_token,
    refreshToken: raw.refresh_token,
    expiresAt: raw.expires_at,
    operator: raw.operator ? {
      id: raw.operator.id,
      email: raw.operator.email,
      name: raw.operator.name,
      mfaEnabled: raw.operator.mfa_enabled,
    } : undefined,
  };
}

export function mfaStatusFromRaw(raw: { mfa_enabled: boolean }): { enabled: boolean } {
  return {
    enabled: raw.mfa_enabled,
  };
}

export function mfaEnrollFromRaw(raw: { secret: string; uri?: string }): { secret: string; uri: string } {
  return {
    secret: raw.secret,
    uri: raw.uri ?? "",
  };
}

export function mfaEnableFromRaw(raw: { success: boolean; backup_codes?: string[] }): { success: boolean; backupCodes?: string[] } {
  return {
    success: raw.success,
    backupCodes: raw.backup_codes,
  };
}

export function mfaDisableFromRaw(raw: { success: boolean }): { success: boolean } {
  return {
    success: raw.success,
  };
}

export function mfaVerifySetupFromRaw(raw: { verified: boolean }): { verified: boolean } {
  return {
    verified: raw.verified,
  };
}

export function mfaVerifyBackupFromRaw(raw: { valid: boolean }): { valid: boolean } {
  return {
    valid: raw.valid,
  };
}

export function mfaRegenerateCodesFromRaw(raw: { backup_codes: string[] }): { backupCodes: string[] } {
  return {
    backupCodes: raw.backup_codes,
  };
}

export function mfaVerifyFromRaw(raw: RawMFAVerifyResponse): {
  success: boolean;
  sessionId?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: number;
  operator?: {
    id: string;
    email: string;
    name: string;
    mfaEnabled: boolean;
  };
} {
  return {
    success: raw.success,
    sessionId: raw.session_id,
    accessToken: raw.access_token,
    refreshToken: raw.refresh_token,
    expiresAt: raw.expires_at,
    operator: raw.operator ? {
      id: raw.operator.id,
      email: raw.operator.email,
      name: raw.operator.name,
      mfaEnabled: raw.operator.mfa_enabled,
    } : undefined,
  };
}
