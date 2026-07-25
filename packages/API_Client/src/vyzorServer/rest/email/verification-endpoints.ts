

import { restClient, getCSRFToken, fetchAndSetCSRFToken } from "../_shared/rest-client";
import {
  verifyEmailFromRaw,
  resendVerificationFromRaw,
  cancelVerificationFromRaw,
  pollVerificationFromRaw,
  type RawVerifyEmailResponse,
  type RawResendVerificationResponse,
  type RawCancelVerificationResponse,
  type RawPollVerificationResponse,
} from "@/domain/email";
import type {
  VerifyEmailResponse,
  ResendVerificationResponse,
  CancelVerificationResponse,
  PollVerificationResponse,
} from "@/domain/email";

const EMAIL_PATHS = {
  verifyEmail: "/v1/auth/verify-email",
  resendVerification: "/v1/auth/resend-verification",
  cancelVerification: "/v1/auth/cancel-verification",
  pollVerification: "/v1/auth/poll-verification",
} as const;

async function ensureCSRFToken(): Promise<void> {
  if (!getCSRFToken()) {
    await fetchAndSetCSRFToken();
  }
}


export async function verifyEmailGet(token: string): Promise<VerifyEmailResponse> {
  const response = await restClient.get<RawVerifyEmailResponse>(`${EMAIL_PATHS.verifyEmail}?token=${encodeURIComponent(token)}`);
  return verifyEmailFromRaw(response);
}


export async function verifyEmail(token: string): Promise<VerifyEmailResponse> {
  await ensureCSRFToken();
  const response = await restClient.post<RawVerifyEmailResponse>(EMAIL_PATHS.verifyEmail, { token });
  return verifyEmailFromRaw(response);
}


export async function resendVerification(email: string): Promise<ResendVerificationResponse> {
  await ensureCSRFToken();
  const response = await restClient.post<RawResendVerificationResponse>(EMAIL_PATHS.resendVerification, { email });
  return resendVerificationFromRaw(response);
}


export async function cancelVerification(email: string): Promise<CancelVerificationResponse> {
  await ensureCSRFToken();
  const response = await restClient.post<RawCancelVerificationResponse>(EMAIL_PATHS.cancelVerification, { email });
  return cancelVerificationFromRaw(response);
}


export async function pollVerification(token: string): Promise<PollVerificationResponse> {
  const response = await restClient.get<RawPollVerificationResponse>(`${EMAIL_PATHS.pollVerification}?token=${encodeURIComponent(token)}`);
  return pollVerificationFromRaw(response);
}
