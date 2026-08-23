import { useMutation } from '@tanstack/react-query';
import { getMfa } from '@vyzorix/api-client';
import type { MFAVerifyResult } from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';

export interface MfaVerifyInput {
  operatorId: string;
  code: string;
}

export function useMfaVerify() {
  const setFromMfaVerify = useAuthStore((s) => s.setFromMfaVerify);

  return useMutation<MFAVerifyResult, Error, MfaVerifyInput>({
    mutationFn: (input) => getMfa().postAuthMfaVerify({ operator_id: input.operatorId, code: input.code }),
    onSuccess: async (response) => {
      if (response.success) {
        await setFromMfaVerify(response);
      }
    },
  });
}
