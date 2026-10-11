import { usePostAuthRegister } from '@/generated-rq/auth/auth-session';

export function useRegister() {
  return usePostAuthRegister();
}
