/**
 * verificationClient.ts - Email verification polling and lifecycle client.
 *
 * Handles verification status polling, token resend, and session cancellation.
 */

import { logger } from "@/lib/logger";

const API_BASE = "";

// Types
export type VerificationStatus = "waiting" | "success" | "expired" | "invalid";

export interface PollVerificationResponse {
  status: VerificationStatus;
  email?: string;
}

export interface ResendResponse {
  success: boolean;
  message: string;
}

export interface CancelResponse {
  success: boolean;
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
 * Poll for email verification status using the token.
 */
export async function pollVerificationStatus(token: string): Promise<PollVerificationResponse> {
  logger.info("verification", "-> GET /v1/auth/poll-verification", {
    token: token.substring(0, 8) + "...",
  });
  const res = await fetch(
    `${API_BASE}/v1/auth/poll-verification?token=${encodeURIComponent(token)}`,
    { credentials: "include" },
  );
  const out = await jsonOrThrow<PollVerificationResponse>(res);
  logger.info("verification", "<- poll status", { status: out.status });
  return out;
}

/**
 * Resend verification email for the given address.
 */
export async function triggerTokenResend(email: string): Promise<ResendResponse> {
  logger.info("verification", "-> POST /v1/auth/resend-token", { email });
  const res = await fetch(`${API_BASE}/v1/auth/resend-token`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email }),
  });
  const out = await jsonOrThrow<ResendResponse>(res);
  logger.info("verification", "<- resend OK");
  return out;
}

/**
 * Cancel pending verification session.
 */
export async function cancelVerificationSession(email: string): Promise<CancelResponse> {
  logger.info("verification", "-> POST /v1/auth/cancel-verification", {
    email,
  });
  const res = await fetch(`${API_BASE}/v1/auth/cancel-verification`, {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    credentials: "include",
    body: JSON.stringify({ email }),
  });
  const out = await jsonOrThrow<CancelResponse>(res);
  logger.info("verification", "<- cancel OK");
  return out;
}
