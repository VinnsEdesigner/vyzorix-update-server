import { createJSONStorage } from 'zustand/middleware';
import type { PresetCommandType, CommandParams } from '@vyzorix/api-client';
import { createVyzorStore } from '@/lib/state';

export type QueuedCommandStatus = 'queued' | 'dispatching' | 'dispatched' | 'failed';

export interface QueuedCommand {
  id: string;
  imei: string;
  commandType: PresetCommandType;
  params?: CommandParams;
  organizationId: string;
  queuedAt: number;
  attempts: number;
  nextRetryAt: number;
  status: QueuedCommandStatus;
  dispatchId?: string;
  lastError?: string;
}

export interface CommandQueueState {
  queue: QueuedCommand[];
  isFlushing: boolean;
  lastFlushError: string | null;
  enqueue: (cmd: Omit<QueuedCommand, 'queuedAt' | 'attempts' | 'nextRetryAt' | 'status'>) => string;
  dequeue: (id: string) => void;
  markDispatched: (id: string, dispatchId: string) => void;
  markFailed: (id: string, error: string) => void;
  setFlushing: (isFlushing: boolean) => void;
  setLastFlushError: (error: string | null) => void;
  getQueueForOrganization: (organizationId: string) => QueuedCommand[];
  clear: () => void;
  clearForOrganization: (organizationId: string) => void;
}

const MAX_ATTEMPTS = 5;
const BASE_BACKOFF_MS = 2_000;

function nextRetryAt(attempts: number): number {
  return Date.now() + BASE_BACKOFF_MS * 2 ** attempts;
}

export const useCommandQueueStore = createVyzorStore<CommandQueueState>(
  'CommandQueueStore',
  (set, get) => ({
    queue: [],
    isFlushing: false,
    lastFlushError: null,

    enqueue: (cmd) => {
      const id = cmd.id ?? `cq-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`;
      const entry: QueuedCommand = {
        ...cmd,
        id,
        queuedAt: Date.now(),
        attempts: 0,
        nextRetryAt: Date.now(),
        status: 'queued',
      };
      set((state) => ({ queue: [...state.queue, entry] }));
      return id;
    },

    dequeue: (id) =>
      set((state) => ({ queue: state.queue.filter((c) => c.id !== id) })),

    markDispatched: (id, dispatchId) =>
      set((state) => ({
        queue: state.queue.map((c) =>
          c.id === id ? { ...c, status: 'dispatched', dispatchId } : c,
        ),
      })),

    markFailed: (id, error) =>
      set((state) => ({
        queue: state.queue.map((c) => {
          if (c.id !== id) return c;
          const attempts = c.attempts + 1;
          if (attempts >= MAX_ATTEMPTS) {
            return { ...c, status: 'failed', attempts, lastError: error };
          }
          return {
            ...c,
            status: 'queued',
            attempts,
            nextRetryAt: nextRetryAt(attempts),
            lastError: error,
          };
        }),
      })),

    setFlushing: (isFlushing) => set({ isFlushing }),
    setLastFlushError: (lastFlushError) => set({ lastFlushError }),

    getQueueForOrganization: (organizationId) =>
      get().queue.filter((c) => c.organizationId === organizationId && c.status === 'queued'),

    clear: () => set({ queue: [], lastFlushError: null }),

    clearForOrganization: (organizationId) =>
      set((state) => ({
        queue: state.queue.filter((c) => c.organizationId !== organizationId),
      })),
  }),
  {
    persist: {
      name: 'vyzorix-command-queue',
      storage: createJSONStorage(() => localStorage),
      partialize: (state) => ({ queue: state.queue }),
    },
  },
);
