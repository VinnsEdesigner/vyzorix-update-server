

import { restClient, getCSRFToken, fetchAndSetCSRFToken, clearAuthContext, setAuthToken, setRefreshToken } from "../_shared/rest-client";
import {
  loginRequestToRaw,
  registerRequestToRaw,
  updateNameRequestToRaw,
  forgotPasswordRequestToRaw,
  resetPasswordRequestToRaw,
  verifyEmailRequestToRaw,
  mfaVerifyRequestToRaw,
  mfaCodeRequestToRaw,
  mfaEnrollRequestToRaw,
  backupCodeVerifyRequestToRaw,
  mfaStatusResponseFromRaw,
  mfaEnrollResponseFromRaw,
  loginResponseFromRaw,
  isMFAResponse,
  registerResponseFromRaw,
  meResponseFromRaw,
  authTokensFromRaw,
  mfaVerifyResponseFromRaw,
  type RawMFAVerifyResponse,
} from "@/domain/auth";
import type {
  LoginResponse,
  LoginWithTokensResponse,
  RegisterResponse,
  MeResponse,
  OperatorRole,
  ForgotPasswordResponse,
  ResetPasswordResponse,
  MFAStatusResponse,
  MFAEnrollResponse,
  MFAVerifyResponse,
  MFAEnableResponse,
  AuthTokens,
} from "@/domain/auth";
import type { VerifyEmailResponse } from "@/domain/email";
import {
  type RawLoginResponse,
  type RawLoginMFARequiredResponse,
  type RawLoginWithTokensResponse,
  type RawLoginWithTokensMFARequiredResponse,
  type RawMeResponse,
  loginWithTokensResponseFromRaw,
  isLoginWithTokensMFARequired,
} from "@/domain/auth/auth-mappers";

const AUTH_PATHS = {
  login: "/v1/auth/login",
  loginTokens: "/v1/auth/login/tokens",
  register: "/v1/auth/register",
  logout: "/v1/auth/logout",
  refresh: "/v1/auth/refresh",
  me: "/v1/auth/me",
  updateName: "/v1/auth/me",
  forgotPassword: "/v1/auth/forgot-password",
  resetPassword: "/v1/auth/reset-password",
  resendPasswordReset: "/v1/auth/resend-password-reset",
  verifyEmail: "/v1/auth/verify-email",
  resendVerification: "/v1/auth/resend-verification",
  mfa: {
    status: "/v1/auth/mfa/status",
    enroll: "/v1/auth/mfa/enroll",
    verifySetup: "/v1/auth/mfa/verify-setup",
    enable: "/v1/auth/mfa/enable",
    disable: "/v1/auth/mfa/disable",
    verify: "/v1/auth/mfa/verify",
    verifyBackup: "/v1/auth/mfa/verify-backup",
    regenerateBackupCodes: "/v1/auth/mfa/regenerate-backup-codes",
  },
  sessions: "/v1/auth/sessions",
  sessionsConcurrent: "/v1/auth/sessions/concurrent",
  sessionsRevokeAll: "/v1/auth/sessions/revoke-all",
} as const;

async function ensureCSRFToken(): Promise<void> {
  if (!getCSRFToken()) {
    await fetchAndSetCSRFToken();
  }
}

interface LoginSuccess {
  success: true;
  data: LoginResponse;
}

interface LoginMFARequired {
  mfaRequired: true;
  operatorId: string;
}

export type LoginResult = LoginSuccess | LoginMFARequired;

interface LoginWithTokensSuccess {
  success: true;
  data: LoginWithTokensResponse;
}

interface LoginWithTokensMFARequired {
  mfaRequired: true;
  operatorId: string;
  email: string;
  name: string;
  mfaEnabled: boolean;
}

export type LoginWithTokensResult = LoginWithTokensSuccess | LoginWithTokensMFARequired;

export async function fetchCSRFToken(): Promise<string> {
  return fetchAndSetCSRFToken();
}

export async function login(credentials: {
  email: string;
  password: string;
}): Promise<LoginResult> {
  await ensureCSRFToken();
  const raw = await restClient.post<RawLoginResponse | RawLoginMFARequiredResponse>(
    AUTH_PATHS.login,
    loginRequestToRaw(credentials)
  );

  if (isMFAResponse(raw)) {
    return { mfaRequired: true, operatorId: raw.operator_id };
  }

  return { success: true, data: loginResponseFromRaw(raw) };
}


export async function loginWithTokens(credentials: {
  email: string;
  password: string;
}): Promise<LoginWithTokensResult> {
  await ensureCSRFToken();
  const raw = await restClient.post<RawLoginWithTokensResponse | RawLoginWithTokensMFARequiredResponse>(
    AUTH_PATHS.loginTokens,
    loginRequestToRaw(credentials)
  );

  if (isLoginWithTokensMFARequired(raw)) {
    return {
      mfaRequired: true,
      operatorId: raw.operator_id,
      email: raw.email,
      name: raw.name,
      mfaEnabled: raw.mfa_enabled,
    };
  }

  const result = loginWithTokensResponseFromRaw(raw);
  setAuthToken(result.access_token);
  setRefreshToken(result.refresh_token);
  return { success: true, data: result };
}


export async function register(credentials: {
  email: string;
  password: string;
  name: string;
}): Promise<RegisterResponse> {
  await ensureCSRFToken();
  const raw = await restClient.post<{
    operator_id: string;
    email: string;
    name: string;
  }>(AUTH_PATHS.register, registerRequestToRaw(credentials));
  return registerResponseFromRaw(raw);
}


export async function forgotPassword(
  email: string
): Promise<ForgotPasswordResponse> {
  await ensureCSRFToken();
  return restClient.post<ForgotPasswordResponse>(
    AUTH_PATHS.forgotPassword,
    forgotPasswordRequestToRaw(email)
  );
}


export async function resetPassword(
  token: string,
  newPassword: string
): Promise<ResetPasswordResponse> {
  await ensureCSRFToken();
  return restClient.post<ResetPasswordResponse>(
    AUTH_PATHS.resetPassword,
    resetPasswordRequestToRaw(token, newPassword)
  );
}


export async function resendPasswordReset(
  email: string
): Promise<{ success: boolean; error?: string; retryAfter?: number; lockedUntil?: number }> {
  await ensureCSRFToken();
  return restClient.post<{ success: boolean; error?: string; retryAfter?: number; lockedUntil?: number }>(
    AUTH_PATHS.resendPasswordReset,
    forgotPasswordRequestToRaw(email)
  );
}


export async function verifyEmail(token: string): Promise<VerifyEmailResponse> {
  await ensureCSRFToken();
  return restClient.post<VerifyEmailResponse>(
    AUTH_PATHS.verifyEmail,
    verifyEmailRequestToRaw(token)
  );
}


export async function resendVerification(
  email: string
): Promise<{ success: boolean }> {
  await ensureCSRFToken();
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.resendVerification,
    forgotPasswordRequestToRaw(email)
  );
}






export async function logout(): Promise<{ success: boolean }> {
  const result = await restClient.post<{ success: boolean }>(AUTH_PATHS.logout);
  clearAuthContext();
  return result;
}


export async function getMe(): Promise<MeResponse | null> {
  try {
    const raw = await restClient.get<RawMeResponse>(AUTH_PATHS.me);
    return meResponseFromRaw(raw);
  } catch {
    return null;
  }
}

export interface UpdateNameResponse {
  id: string;
  email: string;
  name: string;
  role?: string;
  mfa_enabled: boolean;
  email_verified: boolean;
}

export async function updateName(name: string): Promise<UpdateNameResponse> {
  const raw = await restClient.patch<{
    id: string;
    email: string;
    name: string;
    role: OperatorRole;
    mfa_enabled: boolean;
    email_verified: boolean;
  }>(AUTH_PATHS.updateName, updateNameRequestToRaw(name));
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    mfa_enabled: raw.mfa_enabled,
    email_verified: raw.email_verified,
  };
}






export async function getMFAStatus(): Promise<MFAStatusResponse> {
  const raw = await restClient.get<{ enabled: boolean; backup_codes?: string[] }>(
    AUTH_PATHS.mfa.status
  );
  return mfaStatusResponseFromRaw(raw);
}


export async function enrollMFA(): Promise<MFAEnrollResponse> {
  const raw = await restClient.post<{ secret: string; qr_code_url: string }>(
    AUTH_PATHS.mfa.enroll
  );
  return mfaEnrollResponseFromRaw(raw);
}


export async function verifyMFASetup(
  code: string
): Promise<MFAVerifyResponse> {
  return restClient.post<MFAVerifyResponse>(
    AUTH_PATHS.mfa.verifySetup,
    mfaEnrollRequestToRaw(code)
  );
}


export async function enableMFA(code: string): Promise<MFAEnableResponse> {
  return restClient.post<MFAEnableResponse>(
    AUTH_PATHS.mfa.enable,
    { code, token: code }
  );
}


export async function disableMFA(code: string): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.mfa.disable,
    mfaCodeRequestToRaw(code)
  );
}


export async function verifyMFA(
  operatorId: string,
  code: string
): Promise<MFAVerifyResponse> {
  const raw = await restClient.post<RawMFAVerifyResponse>(
    AUTH_PATHS.mfa.verify,
    mfaVerifyRequestToRaw(operatorId, code)
  );
  const result = mfaVerifyResponseFromRaw(raw);
  if (result.success && result.accessToken && result.refreshToken) {
    setAuthToken(result.accessToken);
    setRefreshToken(result.refreshToken);
  }
  return result;
}


export async function verifyBackupCode(
  code: string
): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.mfa.verifyBackup,
    backupCodeVerifyRequestToRaw(code)
  );
}


export async function regenerateBackupCodes(): Promise<{ backupCodes: string[] }> {
  const raw = await restClient.post<{ backup_codes: string[] }>(
    AUTH_PATHS.mfa.regenerateBackupCodes
  );
  return {
    backupCodes: raw.backup_codes,
  };
}


export async function refreshToken(refreshToken: string): Promise<AuthTokens> {
  const raw = await restClient.post<{
    access_token: string;
    refresh_token: string;
    expires_at: number;
    session_id: string;
  }>(AUTH_PATHS.refresh, { refresh_token: refreshToken });
  return authTokensFromRaw(raw);
}
