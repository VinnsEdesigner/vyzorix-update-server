import { useMutation } from '@tanstack/react-query';
import { register, type RegisterResponse } from '@vyzorix/api-client';

export interface RegisterInput {
  email: string;
  password: string;
  name: string;
}

/**
 * Registration mutation. The server creates the operator and sends a
 * verification email — registration does NOT auto-login (email must be
 * verified before `loginWithTokens` succeeds). The returned operator id/email
 * is exposed so the UI can route to a "check your email" state.
 */
export function useRegister() {
  return useMutation<RegisterResponse, Error, RegisterInput>({
    mutationFn: (input) => register(input),
  });
}
