import { createVyzorStore } from '@/lib/state';
import type { VyzorReportedError, VyzorErrorClassification } from '@/lib/error';

export interface VyzorErrorStoreState {
  error: VyzorReportedError | null;
  report: (
    error: unknown,
    classification: VyzorErrorClassification,
    context?: string,
  ) => void;
  dismiss: () => void;
  incrementRetry: () => void;
  retry: () => void;
  reload: () => void;
  goHome: () => void;
}

function buildReportedError(
  error: unknown,
  classification: VyzorErrorClassification,
  context?: string,
): VyzorReportedError {
  const message =
    error instanceof Error
      ? error.message
      : typeof error === 'string'
        ? error
        : 'Unknown error';
  return {
    id: `err-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
    error: error instanceof Error ? error : null,
    message,
    classification,
    context,
    timestamp: Date.now(),
    retryCount: 0,
  };
}

export const useVyzorErrorStore = createVyzorStore<VyzorErrorStoreState>(
  'VyzorErrorStore',
  (set, get) => ({
    error: null,

    report: (error, classification, context) =>
      set({
        error: buildReportedError(error, classification, context),
      }),

    dismiss: () => set({ error: null }),

    incrementRetry: () =>
      set((state) => {
        if (!state.error) return state;
        return {
          error: { ...state.error, retryCount: state.error.retryCount + 1 },
        };
      }),

    retry: () => {
      get().incrementRetry();
      window.location.reload();
    },

    reload: () => {
      window.location.reload();
    },

    goHome: () => {
      window.location.href = '/';
    },
  }),
);
