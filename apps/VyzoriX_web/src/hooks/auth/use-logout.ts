import { useMutation } from '@tanstack/react-query';
import { postAuthLogout } from '@/generated-rq/auth/auth-session';
import { useAuthStore } from '@/stores/auth-store';

export function useLogout() {
  const clear = useAuthStore((s) => s.clear);
  return useMutation<void, Error, void>({
    mutationFn: async () => {
      try {
        await postAuthLogout();
      } catch {
        // Ignore — clear local state anyway.
      }
      clear();
    },
  });
}
