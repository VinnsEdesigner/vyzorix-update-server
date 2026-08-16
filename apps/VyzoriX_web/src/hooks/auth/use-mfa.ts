import { useMutation } from '@tanstack/react-query';
import { verifyMFA, type MFAVerifyResponse } from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';

export interface MfaVerifyInput {
  operatorId: string;
  code: string;
}

/**
 * MFA verification mutation for the login flow. Calls `verifyMFA`, then
 * ingests the tokens + operator into `authContext` via the store's
 * `setFromMfaVerify` (which also fetches `/me` to reconcile the full operator
 * profile, including organizations/`needs_organization`). On success the
 * staged `mfaChallenge` is cleared (status → `authenticated` or
 * `needs_organization`).
 */
export function useMfaVerify() {
  const setFromMfaVerify = useAuthStore((s) => s.setFromMfaVerify);

  return useMutation<MFAVerifyResponse, Error, MfaVerifyInput>({
    mutationFn: (input) => verifyMFA(input.operatorId, input.code),
    onSuccess: async (response) => {
      if (response.success) {
        await setFromMfaVerify(response);
      }
    },
  });
}
