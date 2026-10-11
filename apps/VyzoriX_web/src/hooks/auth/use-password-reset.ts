import {
  usePostAuthForgotPassword,
  usePostAuthResetPassword,
  usePostAuthResendPasswordReset,
} from '@/generated-rq/auth/auth-session';

export function useForgotPassword() {
  return usePostAuthForgotPassword();
}

export function useResetPassword() {
  return usePostAuthResetPassword();
}

export function useResendPasswordReset() {
  return usePostAuthResendPasswordReset();
}
