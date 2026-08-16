import { useCallback } from 'react';
import { useVyzorErrorStore } from '@/stores/vyzor-error-store';
import { classifyVyzorError } from '@/lib/error';

export function useVyzorErrorReporter() {
  const report = useVyzorErrorStore((s) => s.report);

  return useCallback(
    (error: unknown, context?: string) => {
      const classification = classifyVyzorError(error);
      report(error, classification, context);
      return classification;
    },
    [report],
  );
}
