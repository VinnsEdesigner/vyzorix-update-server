import { usePostAuthMfaVerify } from '@/generated-rq/mfa/mfa-management';
import { useAuthStore } from '@/stores/auth-store';

export function useMfaVerify() {
  const setFromMfaVerify = useAuthStore((s) => s.setFromMfaVerify);
  return usePostAuthMfaVerify({
    mutation: {
      onSuccess: async (response) => {
        if (response.success) {
          await setFromMfaVerify(response);
        }
      },
    },
  });
}
