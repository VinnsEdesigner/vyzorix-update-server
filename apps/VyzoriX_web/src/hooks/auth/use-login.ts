import { useMutation } from '@tanstack/react-query';
import { loginWithTokens, type LoginWithTokensResult } from '@vyzorix/api-client';
import { useAuthStore, type MfaChallenge } from '@/stores/auth-store';

export interface LoginInput {
  email: string;
  password: string;
}

export interface LoginMutationResult {
  result: LoginWithTokensResult;
  mfaChallenge: MfaChallenge | null;
}

/**
 * Login mutation. Uses `loginWithTokens` (API-client flow) so the web app
 * receives JWT + refresh token. On success the response is ingested into
 * `authContext` via the store; on MFA-required the operator is staged as an
 * `MfaChallenge` (status → `mfa_required`) for the MFA form to consume.
 *
 * The hook does NOT navigate — routing is a UI concern. Callers inspect the
 * returned `mfaChallenge` to decide whether to route to `/auth/mfa` or the
 * dashboard.
 */
export function useLogin() {
  const setFromLoginWithTokens = useAuthStore((s) => s.setFromLoginWithTokens);
  const setMfaChallenge = useAuthStore((s) => s.setMfaChallenge);

  return useMutation<LoginMutationResult, Error, LoginInput>({
    mutationFn: async (input) => {
      const result = await loginWithTokens(input);
      let mfaChallenge: MfaChallenge | null = null;
      if ('mfaRequired' in result) {
        mfaChallenge = {
          operatorId: result.operatorId,
          email: result.email,
          name: result.name,
          mfaEnabled: result.mfaEnabled,
        };
      } else {
        setFromLoginWithTokens(result.data);
      }
      return { result, mfaChallenge };
    },
    onSuccess: ({ mfaChallenge }) => {
      setMfaChallenge(mfaChallenge);
    },
  });
}
