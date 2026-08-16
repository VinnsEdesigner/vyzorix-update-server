import { useMutation } from '@tanstack/react-query';
import {
  forgotPassword,
  resetPassword,
  resendPasswordReset,
  type ForgotPasswordResponse,
  type ResetPasswordResponse,
  type ResendResetResponse,
} from '@vyzorix/api-client';

export interface ForgotPasswordInput {
  email: string;
}

export interface ResetPasswordInput {
  token: string;
  newPassword: string;
}

export interface ResendResetInput {
  email: string;
}

/**
 * Forgot-password mutation. The server always returns `{ success: true }`
 * (even for unknown emails) to avoid leaking which addresses are registered.
 */
export function useForgotPassword() {
  return useMutation<ForgotPasswordResponse, Error, ForgotPasswordInput>({
    mutationFn: (input) => forgotPassword(input.email),
  });
}

/**
 * Reset-password mutation. Consumes the token from the reset email link.
 */
export function useResetPassword() {
  return useMutation<ResetPasswordResponse, Error, ResetPasswordInput>({
    mutationFn: (input) => resetPassword(input.token, input.newPassword),
  });
}

/**
 * Resend reset email — used when the reset link expired or was never received.
 */
export function useResendPasswordReset() {
  return useMutation<ResendResetResponse, Error, ResendResetInput>({
    mutationFn: (input) => resendPasswordReset(input.email),
  });
}
