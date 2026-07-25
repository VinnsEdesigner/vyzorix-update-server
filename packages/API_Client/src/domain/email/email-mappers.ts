

import type {
  VerifyEmailResponse,
  ResendVerificationResponse,
  CancelVerificationResponse,
  PollVerificationResponse,
} from "./email-entity";


export interface RawVerifyEmailResponse {
  verified: boolean;
  email?: string;
}


export interface RawResendVerificationResponse {
  message: string;
}


export interface RawCancelVerificationResponse {
  success: boolean;
}


export interface RawPollVerificationResponse {
  verified: boolean;
  email?: string;
}


export function verifyEmailFromRaw(raw: RawVerifyEmailResponse): VerifyEmailResponse {
  return {
    verified: raw.verified,
    email: raw.email,
  };
}


export function resendVerificationFromRaw(raw: RawResendVerificationResponse): ResendVerificationResponse {
  return {
    message: raw.message,
  };
}


export function cancelVerificationFromRaw(raw: RawCancelVerificationResponse): CancelVerificationResponse {
  return {
    success: raw.success,
  };
}


export function pollVerificationFromRaw(raw: RawPollVerificationResponse): PollVerificationResponse {
  return {
    verified: raw.verified,
    email: raw.email,
  };
}
