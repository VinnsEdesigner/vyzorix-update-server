import { useMutation } from '@tanstack/react-query';
import { postAuthForgotPassword, postAuthResetPassword, postAuthResendPasswordReset } from '@/generated-rq/auth/auth-session';
import type { MessageResult, SuccessResult } from '@vyzorix/api-client';

export interface ForgotPasswordInput { email: string; }
export interface ResetPasswordInput { token: string; newPassword: string; }
export interface ResendResetInput { email: string; }

export function useForgotPassword() {
  return useMutation<MessageResult, Error, ForgotPasswordInput>({
    mutationFn: (input) => postAuthForgotPassword({ email: input.email }),
  });
}

export function useResetPassword() {
  return useMutation<SuccessResult, Error, ResetPasswordInput>({
    mutationFn: (input) => postAuthResetPassword({ token: input.token, newPassword: input.newPassword }),
  });
}

export function useResendPasswordReset() {
  return useMutation<SuccessResult, Error, ResendResetInput>({
    mutationFn: (input) => postAuthResendPasswordReset({ email: input.email }),
  });
}
