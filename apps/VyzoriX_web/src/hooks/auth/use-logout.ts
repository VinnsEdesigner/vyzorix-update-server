import { usePostAuthLogout } from '@/generated-rq/auth/auth-session';
import { useAuthStore } from '@/stores/auth-store';

export function useLogout() {
  const clear = useAuthStore((s) => s.clear);
  return usePostAuthLogout({
    mutation: {
      onSettled: () => {
        // Clear local state whether or not the server call succeeded.
        clear();
      },
    },
  });
}
