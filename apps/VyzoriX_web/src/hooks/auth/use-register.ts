import { useMutation } from '@tanstack/react-query';
import { postAuthRegister } from '@/generated-rq/auth/auth-session';
import type { RegisterResult } from '@vyzorix/api-client';

export interface RegisterInput {
  email: string;
  password: string;
  name: string;
}

export function useRegister() {
  return useMutation<RegisterResult, Error, RegisterInput>({
    mutationFn: (input) => postAuthRegister(input),
  });
}
