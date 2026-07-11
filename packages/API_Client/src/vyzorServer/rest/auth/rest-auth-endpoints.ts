/**
 * Authentication REST Endpoints
 * 
 * All endpoints use HttpOnly cookie-based authentication via the 'vyz_session' cookie.
 * The browser automatically handles cookie storage and transmission.
 * 
 * No token storage in localStorage/sessionStorage - this is handled securely by the server.
 */

import { restClient } from "../_shared/rest-client";
import {
  loginRequestToRaw,
  registerRequestToRaw,
  updateNameRequestToRaw,
  forgotPasswordRequestToRaw,
  resetPasswordRequestToRaw,
  verifyEmailRequestToRaw,
  mfaVerifyRequestToRaw,
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
  RegisterResponse,
  MeResponse,
  Operator,
  ForgotPasswordResponse,
  ResetPasswordResponse,
  VerifyEmailResponse,
  MFAStatusResponse,
  MFAEnrollResponse,
  MFAVerifyResponse,
  MFAEnableResponse,
  AuthTokens,
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

// ============================================================================
// Public Endpoints (No Auth Required)
// ============================================================================

interface LoginSuccess {
  success: true;
  data: LoginResponse;
}

interface LoginMFARequired {
  mfaRequired: true;
  operatorId: string;
}

export type LoginResult = LoginSuccess | LoginMFARequired;

/**
 * Login with email and password.
 * Server sets HttpOnly cookie on successful login.
 */
export async function login(credentials: {
  email: string;
  password: string;
}): Promise<LoginResult> {
  const raw = await restClient.post<RawLoginResponse | RawLoginMFARequiredResponse>(
    AUTH_PATHS.login,
    loginRequestToRaw(credentials)
  );

  if (isMFAResponse(raw)) {
    return { mfaRequired: true, operatorId: raw.operator_id };
  }

  return { success: true, data: loginResponseFromRaw(raw) };
}

/**
 * Register a new operator account.
 * Server sets HttpOnly cookie on successful registration.
 */
export async function register(credentials: {
  email: string;
  password: string;
  name: string;
}): Promise<RegisterResponse> {
  const raw = await restClient.post<{
    operator_id: string;
    email: string;
    name: string;
  }>(AUTH_PATHS.register, registerRequestToRaw(credentials));
  return registerResponseFromRaw(raw);
}

/**
 * Request a password reset email.
 */
export async function forgotPassword(
  email: string
): Promise<ForgotPasswordResponse> {
  return restClient.post<ForgotPasswordResponse>(
    AUTH_PATHS.forgotPassword,
    forgotPasswordRequestToRaw(email)
  );
}

/**
 * Reset password using a token from the reset email.
 */
export async function resetPassword(
  token: string,
  newPassword: string
): Promise<ResetPasswordResponse> {
  return restClient.post<ResetPasswordResponse>(
    AUTH_PATHS.resetPassword,
    resetPasswordRequestToRaw(token, newPassword)
  );
}

/**
 * Verify email using a token from the verification email.
 */
export async function verifyEmail(token: string): Promise<VerifyEmailResponse> {
  return restClient.post<VerifyEmailResponse>(
    AUTH_PATHS.verifyEmail,
    verifyEmailRequestToRaw(token)
  );
}

/**
 * Resend email verification email.
 */
export async function resendVerification(
  email: string
): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.resendVerification,
    forgotPasswordRequestToRaw(email)
  );
}

// ============================================================================
// Authenticated Endpoints (Require Session Cookie)
// ============================================================================

/**
 * Logout and clear the session cookie.
 */
export async function logout(): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(AUTH_PATHS.logout);
}

/**
 * Get current authenticated operator's profile.
 * Requires valid session cookie.
 */
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

/**
 * Update operator's name.
 * Requires valid session cookie.
 */
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

// ============================================================================
// MFA Endpoints (Require Session Cookie)
// ============================================================================

/**
 * Get MFA enrollment status.
 */
export async function getMFAStatus(): Promise<MFAStatusResponse> {
  const raw = await restClient.get<{ enabled: boolean; backup_codes?: string[] }>(
    AUTH_PATHS.mfa.status
  );
  return mfaStatusResponseFromRaw(raw);
}

/**
 * Start MFA enrollment - returns TOTP secret for QR code generation.
 */
export async function enrollMFA(): Promise<MFAEnrollResponse> {
  const raw = await restClient.post<{ secret: string; qr_code_url: string }>(
    AUTH_PATHS.mfa.enroll
  );
  return mfaEnrollResponseFromRaw(raw);
}

/**
 * Verify MFA setup with a TOTP code.
 */
export async function verifyMFASetup(
  code: string
): Promise<MFAVerifyResponse> {
  return restClient.post<MFAVerifyResponse>(
    AUTH_PATHS.mfa.verifySetup,
    mfaEnrollRequestToRaw(code)
  );
}

/**
 * Enable MFA after successful setup verification.
 */
export async function enableMFA(code: string): Promise<MFAEnableResponse> {
  return restClient.post<MFAEnableResponse>(
    AUTH_PATHS.mfa.enable,
    mfaVerifyRequestToRaw(code)
  );
}

/**
 * Disable MFA (requires current TOTP code).
 */
export async function disableMFA(code: string): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.mfa.disable,
    mfaVerifyRequestToRaw(code)
  );
}

/**
 * Verify TOTP code during login (when MFA is required).
 * Requires operatorId from the MFA-required login response.
 */
export async function verifyMFA(
  operatorId: string,
  code: string
): Promise<{
  success: boolean;
  sessionId?: string;
  accessToken?: string;
  refreshToken?: string;
  expiresAt?: number;
  operator?: Operator;
}> {
  const raw = await restClient.post<RawMFAVerifyResponse>(
    AUTH_PATHS.mfa.verify,
    mfaVerifyRequestToRaw(operatorId, code)
  );
  return mfaVerifyResponseFromRaw(raw);
}

/**
 * Verify a backup code during login.
 */
export async function verifyBackupCode(
  code: string
): Promise<{ success: boolean }> {
  return restClient.post<{ success: boolean }>(
    AUTH_PATHS.mfa.verifyBackup,
    backupCodeVerifyRequestToRaw(code)
  );
}

/**
 * Regenerate MFA backup codes.
 */
export async function regenerateBackupCodes(): Promise<{ backupCodes: string[] }> {
  return restClient.post<{ backup_codes: string[] }>(
    AUTH_PATHS.mfa.regenerateBackupCodes
  );
}

/**
 * Refresh access token using a refresh token.
 * Implements refresh token rotation for security.
 */
export async function refreshToken(refreshToken: string): Promise<AuthTokens> {
  const raw = await restClient.post<{
    access_token: string;
    refresh_token: string;
    expires_at: number;
    session_id: string;
  }>(AUTH_PATHS.refresh, { refresh_token: refreshToken });
  return authTokensFromRaw(raw);
}
