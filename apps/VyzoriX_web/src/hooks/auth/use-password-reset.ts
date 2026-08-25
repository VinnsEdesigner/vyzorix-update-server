import { useMutation } from '@tanstack/react-query';
import { getAuth } from '@vyzorix/api-client';
import type { MessageResult, SuccessResult } from '@vyzorix/api-client';

export interface ForgotPasswordInput { email: string; }
export interface ResetPasswordInput { token: string; newPassword: string; }
export interface ResendResetInput { email: string; }

export function useForgotPassword() {
  return useMutation<MessageResult, Error, ForgotPasswordInput>({
    mutationFn: (input) => getAuth().postAuthForgotPassword({ email: input.email }),
  });
}

export function useResetPassword() {
  return useMutation<SuccessResult, Error, ResetPasswordInput>({
    mutationFn: (input) => getAuth().postAuthResetPassword({ token: input.token, newPassword: input.newPassword }),
  });
}

export function useResendPasswordReset() {
  return useMutation<SuccessResult, Error, ResendResetInput>({
    mutationFn: (input) => getAuth().postAuthResendPasswordReset({ email: input.email }),
  });
}
