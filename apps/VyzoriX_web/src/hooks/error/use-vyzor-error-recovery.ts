import { useVyzorErrorStore } from '@/stores/vyzor-error-store';

export function useVyzorErrorRecovery() {
  const error = useVyzorErrorStore((s) => s.error);
  const retry = useVyzorErrorStore((s) => s.retry);
  const reload = useVyzorErrorStore((s) => s.reload);
  const goHome = useVyzorErrorStore((s) => s.goHome);
  const dismiss = useVyzorErrorStore((s) => s.dismiss);

  return { error, retry, reload, goHome, dismiss };
}
