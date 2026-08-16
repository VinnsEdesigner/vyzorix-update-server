

import { restClient, getCSRFToken, fetchAndSetCSRFToken } from "../_shared/rest-client";
import { forgotPasswordRequestToRaw } from "../../../domain/auth";
import type { ForgotPasswordResponse, ResetPasswordResponse } from "../../../domain/auth";

const PASSWORD_PATHS = {
  forgotPassword: "/v1/auth/forgot-password",
  resetPassword: "/v1/auth/reset-password",
  resendReset: "/v1/auth/resend-password-reset",
} as const;

export interface ResendResetResponse {
  success: boolean;
  message?: string;
  error?: string;
  retryAfter?: number;
  lockedUntil?: number;
}

async function ensureCSRFToken(): Promise<void> {
  if (!getCSRFToken()) {
    await fetchAndSetCSRFToken();
  }
}


export async function forgotPassword(email: string): Promise<ForgotPasswordResponse> {
  await ensureCSRFToken();
  return restClient.post<ForgotPasswordResponse>(PASSWORD_PATHS.forgotPassword, forgotPasswordRequestToRaw(email));
}


export async function resetPassword(token: string, newPassword: string): Promise<ResetPasswordResponse> {
  await ensureCSRFToken();
  return restClient.post<ResetPasswordResponse>(PASSWORD_PATHS.resetPassword, {
    token,
    newPassword,
  });
}


export async function resendPasswordReset(email: string): Promise<ResendResetResponse> {
  await ensureCSRFToken();
  return restClient.post<ResendResetResponse>(PASSWORD_PATHS.resendReset, forgotPasswordRequestToRaw(email));
}
