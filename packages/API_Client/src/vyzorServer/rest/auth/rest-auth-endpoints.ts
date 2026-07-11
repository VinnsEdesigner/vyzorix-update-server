import { restClient } from "../_shared/rest-client";
import {
  loginRequestToRaw,
  registerRequestToRaw,
  refreshRequestToRaw,
  updateNameRequestToRaw,
  forgotPasswordRequestToRaw,
  resetPasswordRequestToRaw,
  verifyEmailRequestToRaw,
  mfaVerifyRequestToRaw,
  mfaEnrollRequestToRaw,
  mfaStatusResponseFromRaw,
  mfaEnrollResponseFromRaw,
  loginResponseFromRaw,
  isMFAResponse,
  registerResponseFromRaw,
  meResponseFromRaw,
  authTokensFromRaw,
} from "@/domain/auth";
import type {
  LoginResponse,
  RegisterResponse,
  MeResponse,
  AuthTokens,
  Operator,
  ForgotPasswordResponse,
  ResetPasswordResponse,
  VerifyEmailResponse,
  MFAStatusResponse,
  MFAEnrollResponse,
  MFAVerifyResponse,
  MFAEnableResponse,
} from "@/domain/auth";
import type { RawLoginResponse, RawLoginMFARequiredResponse } from "@/domain/auth/auth-mappers";

const AUTH_PATHS = {
  login: "/v1/auth/login",
  register: "/v1/auth/register",
  logout: "/v1/auth/logout",
  refresh: "/v1/auth/refresh",
  me: "/v1/auth/me",
  updateName: "/v1/auth/me",
  forgotPassword: "/v1/auth/forgot-password",
  resetPassword: "/v1/auth/reset-password",
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
} as const;

interface LoginResult {
  success: true;
  data: LoginResponse;
}

interface MFAResult {
  mfaRequired: true;
  operatorId: string;
}

export type LoginSuccessResult = LoginResult | MFAResult;

export async function login(
  credentials: { email: string; password: string }
): Promise<LoginSuccessResult> {
  const raw = await restClient.post<RawLoginResponse | RawLoginMFARequiredResponse>(
    AUTH_PATHS.login,
    loginRequestToRaw(credentials)
  );

  if (isMFAResponse(raw)) {
    return { mfaRequired: true, operatorId: raw.operator_id };
  }

  return { success: true, data: loginResponseFromRaw(raw) };
}

export async function register(
  credentials: { email: string; password: string; name: string }
): Promise<RegisterResponse> {
  const raw = await restClient.post<{ operator_id: string; email: string; name: string }>(
    AUTH_PATHS.register,
    registerRequestToRaw(credentials)
  );
  return registerResponseFromRaw(raw);
}

export async function logout(): Promise<{ success: boolean }> {
  const raw = await restClient.post<{ success: boolean }>(AUTH_PATHS.logout);
  clearStoredAuth();
  return raw;
}

export async function refreshToken(refreshToken: string): Promise<AuthTokens> {
  const raw = await restClient.post<{
    access_token: string;
    refresh_token: string;
    expires_at: number;
    session_id: string;
  }>(AUTH_PATHS.refresh, refreshRequestToRaw(refreshToken));
  return authTokensFromRaw(raw);
}

export async function getMe(): Promise<MeResponse | null> {
  try {
    const raw = await restClient.get<{
      id: string;
      email: string;
      name: string;
      role: string;
      mfa_enabled: boolean;
      email_verified: boolean;
      thresholds?: unknown;
      client?: unknown;
    }>(AUTH_PATHS.me);
    return meResponseFromRaw(raw as any);
  } catch {
    return null;
  }
}

export async function updateName(name: string): Promise<Operator> {
  const raw = await restClient.patch<{
    id: string;
    email: string;
    name: string;
    role: string;
    mfa_enabled: boolean;
    email_verified: boolean;
  }>(AUTH_PATHS.updateName, updateNameRequestToRaw(name));
  return {
    id: raw.id,
    email: raw.email,
    name: raw.name,
    role: raw.role as any,
    mfaEnabled: raw.mfa_enabled,
    emailVerified: raw.email_verified,
  };
}

export async function forgotPassword(email: string): Promise<ForgotPasswordResponse> {
  return restClient.post<ForgotPasswordResponse>(
    AUTH_PATHS.forgotPassword,
    forgotPasswordRequestToRaw(email)
  );
}

export async function resetPassword(token: string, newPassword: string): Promise<ResetPasswordResponse> {
  return restClient.post<ResetPasswordResponse>(
    AUTH_PATHS.resetPassword,
    resetPasswordRequestToRaw(token, newPassword)
  );
}

export async function verifyEmail(token: string): Promise<VerifyEmailResponse> {
  return restClient.post<VerifyEmailResponse>(
    AUTH_PATHS.verifyEmail,
    verifyEmailRequestToRaw(token)
  );
}

export async function resendVerification(email: string): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.resendVerification,
    forgotPasswordRequestToRaw(email)
  );
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

export async function verifyMFASetup(code: string): Promise<MFAVerifyResponse> {
  return restClient.post<MFAVerifyResponse>(
    AUTH_PATHS.mfa.verifySetup,
    mfaEnrollRequestToRaw(code)
  );
}

export async function enableMFA(code: string): Promise<MFAEnableResponse> {
  return restClient.post<MFAEnableResponse>(
    AUTH_PATHS.mfa.enable,
    mfaVerifyRequestToRaw(code)
  );
}

export async function disableMFA(code: string): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.mfa.disable,
    mfaVerifyRequestToRaw(code)
  );
}

export async function verifyMFA(code: string): Promise<AuthTokens> {
  const raw = await restClient.post<{
    access_token: string;
    refresh_token: string;
    expires_at: number;
    session_id: string;
  }>(AUTH_PATHS.mfa.verify, mfaVerifyRequestToRaw(code));
  return authTokensFromRaw(raw);
}

export async function verifyBackupCode(code: string): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.mfa.verifyBackup,
    mfaVerifyRequestToRaw(code)
  );
}

export async function regenerateBackupCodes(): Promise<{ backupCodes: string[] }> {
  return restClient.post<{ backup_codes: string[] }>(
    AUTH_PATHS.mfa.regenerateBackupCodes
  );
}

const AUTH_TOKENS_KEY = "vyz_auth_tokens";

export interface StoredAuth {
  accessToken: string;
  refreshToken: string;
  expiresAt: number;
  sessionId: string;
  operatorId: string;
  email: string;
}

export function storeAuth(auth: StoredAuth): void {
  try {
    localStorage.setItem(AUTH_TOKENS_KEY, JSON.stringify(auth));
  } catch {
    // localStorage unavailable
  }
}

export function getStoredAuth(): StoredAuth | null {
  try {
    const stored = localStorage.getItem(AUTH_TOKENS_KEY);
    return stored ? JSON.parse(stored) : null;
  } catch {
    return null;
  }
}

export function clearStoredAuth(): void {
  try {
    localStorage.removeItem(AUTH_TOKENS_KEY);
  } catch {
    // localStorage unavailable
  }
}
