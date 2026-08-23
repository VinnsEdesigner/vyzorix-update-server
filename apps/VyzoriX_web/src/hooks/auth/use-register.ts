import { useMutation } from '@tanstack/react-query';
import { getAuth } from '@vyzorix/api-client';
import type { RegisterResult } from '@vyzorix/api-client';

export interface RegisterInput {
  email: string;
  password: string;
  name: string;
}

export function useRegister() {
  return useMutation<RegisterResult, Error, RegisterInput>({
    mutationFn: (input) => getAuth().postAuthRegister(input),
  });
}
