import { useMutation, useQueryClient } from '@tanstack/react-query';
import { logout } from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';

/**
 * Logout mutation. Calls the server logout endpoint (clears the session
 * server-side) then clears local auth state regardless of API outcome — a
 * failed logout request must not leave the user stuck authenticated. The
 * TanStack cache is purged so no stale org-scoped data leaks across users.
 */
export function useLogout() {
  const clear = useAuthStore((s) => s.clear);
  const queryClient = useQueryClient();

  return useMutation<void, Error, void>({
    mutationFn: async () => {
      try {
        await logout();
      } catch {
        // Ignore — clear local state anyway.
      }
      clear();
      queryClient.clear();
    },
  });
}
