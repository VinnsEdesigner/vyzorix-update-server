import { useMutation } from '@tanstack/react-query';
import { postAuthMfaVerify } from '@/generated-rq/mfa/mfa-management';
import type { MFAVerifyResult } from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';

export interface MfaVerifyInput {
  operatorId: string;
  code: string;
}

export function useMfaVerify() {
  const setFromMfaVerify = useAuthStore((s) => s.setFromMfaVerify);
  return useMutation<MFAVerifyResult, Error, MfaVerifyInput>({
    mutationFn: (input) => postAuthMfaVerify({ operator_id: input.operatorId, code: input.code }),
    onSuccess: async (response) => {
      if (response.success) {
        await setFromMfaVerify(response);
      }
    },
  });
}
