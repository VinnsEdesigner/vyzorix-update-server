/**
 * passwordClient.ts - Password reset client.
 *
 * Handles forgot password requests and password reset with token.
 */

import { logger } from "@/lib/logger";

const API_BASE = "";

// Types
export interface ForgotPasswordResponse {
  success: boolean;
  message: string;
}

export interface ResetPasswordResponse {
  message: string;
}

interface ErrorResponse {
  error: string;
  message: string;
}

const jsonOrThrow = async <T>(res: Response): Promise<T> => {
  const contentType = res.headers.get("content-type") ?? "";
  if (!res.ok) {
    let msg = `HTTP ${res.status}`;
    if (contentType.includes("application/json")) {
      const body = (await res.json()) as ErrorResponse;
      msg = body.message ?? body.error ?? msg;
    }
    throw new Error(msg);
  }
  if (contentType.includes("application/json")) {
    return (await res.json()) as T;
  }
  throw new Error(`Expected JSON, got ${contentType}`);
};

/**
 * Request a password reset email for the given address.
 */
export async function requestPasswordReset(email: string): Promise<ForgotPasswordResponse> {
  logger.info("password", "-> POST /v1/auth/forgot-password", { email });
  const res = await fetch(`${API_BASE}/v1/auth/forgot-password`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email }),
  });
  const out = await jsonOrThrow<ForgotPasswordResponse>(res);
  logger.info("password", "<- forgot-password OK");
  return out;
}

/**
 * Reset password using a token from the reset email.
 */
export async function resetPasswordWithToken(
  token: string,
  newPassword: string,
): Promise<ResetPasswordResponse> {
  logger.info("password", "-> POST /v1/auth/reset-password");
  const res = await fetch(`${API_BASE}/v1/auth/reset-password`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ token, newPassword }),
  });
  const out = await jsonOrThrow<ResetPasswordResponse>(res);
  logger.info("password", "<- reset-password OK");
  return out;
}

export interface ResendPasswordResetResponse {
  success: boolean;
  message: string;
  retry_after?: number;
  locked_until?: number;
}

export interface ResendError {
  error: string;
  message: string;
  retry_after?: number;
  locked_until?: number;
}

/**
 * Resend password reset email with rate limiting.
 * Returns retry_after seconds if rate limited.
 */
export async function resendPasswordReset(email: string): Promise<ResendPasswordResetResponse> {
  logger.info("password", "-> POST /v1/auth/resend-password-reset", { email });
  const res = await fetch(`${API_BASE}/v1/auth/resend-password-reset`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email }),
  });

  const contentType = res.headers.get("content-type") ?? "";
  if (!contentType.includes("application/json")) {
    throw new Error(`Expected JSON, got ${contentType}`);
  }

  const data = await res.json();
  if (!res.ok) {
    const err = data as ResendError;
    const msg = err.message ?? `HTTP ${res.status}`;
    const error = new Error(msg) as Error & { retry_after?: number; locked_until?: number };
    error.retry_after = err.retry_after;
    error.locked_until = err.locked_until;
    throw error;
  }

  logger.info("password", "<- resend-password-reset OK");
  return data as ResendPasswordResetResponse;
}
