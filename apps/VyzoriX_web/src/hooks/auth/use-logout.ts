import { useMutation, useQueryClient } from '@tanstack/react-query';
import { getAuth } from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';

export function useLogout() {
  const clear = useAuthStore((s) => s.clear);
  const queryClient = useQueryClient();

  return useMutation<void, Error, void>({
    mutationFn: async () => {
      try {
        await getAuth().postAuthLogout();
      } catch {
        // Ignore — clear local state anyway.
      }
      clear();
      queryClient.clear();
    },
  });
}
