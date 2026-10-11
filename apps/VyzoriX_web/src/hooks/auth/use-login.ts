import { usePostAuthLoginTokens } from '@/generated-rq/auth/auth-session';
import { useAuthStore, type MfaChallenge } from '@/stores/auth-store';

export function useLogin() {
  const setFromLoginWithTokens = useAuthStore((s) => s.setFromLoginWithTokens);
  const setMfaChallenge = useAuthStore((s) => s.setMfaChallenge);

  return usePostAuthLoginTokens({
    mutation: {
      onSuccess: (result) => {
        let mfaChallenge: MfaChallenge | null = null;
        if (result.requires_mfa) {
          mfaChallenge = {
            operatorId: result.operator_id ?? '',
            email: result.email ?? '',
            name: result.name ?? '',
            mfaEnabled: result.mfa_enabled ?? false,
          };
        } else {
          setFromLoginWithTokens(result);
        }
        setMfaChallenge(mfaChallenge);
      },
    },
  });
}
