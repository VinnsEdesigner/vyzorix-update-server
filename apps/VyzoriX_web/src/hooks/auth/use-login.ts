import { useMutation } from '@tanstack/react-query';
import { getAuth, type LoginWithTokensResult } from '@vyzorix/api-client';
import { useAuthStore, type MfaChallenge } from '@/stores/auth-store';

export interface LoginInput {
  email: string;
  password: string;
}

export interface LoginMutationResult {
  result: LoginWithTokensResult;
  mfaChallenge: MfaChallenge | null;
}

export function useLogin() {
  const setFromLoginWithTokens = useAuthStore((s) => s.setFromLoginWithTokens);
  const setMfaChallenge = useAuthStore((s) => s.setMfaChallenge);

  return useMutation<LoginMutationResult, Error, LoginInput>({
    mutationFn: async (input) => {
      const result = await getAuth().postAuthLoginTokens(input);
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
      return { result, mfaChallenge };
    },
    onSuccess: ({ mfaChallenge }) => {
      setMfaChallenge(mfaChallenge);
    },
  });
}
