import { describe, it, expect, beforeEach, vi } from 'vitest';
import { createVyzorStore } from '@/lib/state/vyzor-store-factory';
import { isDevtoolsEnabled } from '@/lib/state/vyzor-store-devtools';

interface CounterState {
  count: number;
  inc: () => void;
  reset: () => void;
}

describe('isDevtoolsEnabled', () => {
  it('returns true when not in production', () => {
    expect(isDevtoolsEnabled()).toBe(true);
  });
});

describe('createVyzorStore', () => {
  beforeEach(() => {
    localStorage.clear();
  });

  it('creates a working store without persist', () => {
    const useCounter = createVyzorStore<CounterState>('CounterStore', (set) => ({
      count: 0,
      inc: () => set((s) => ({ count: s.count + 1 })),
      reset: () => set({ count: 0 }),
    }));

    expect(useCounter.getState().count).toBe(0);
    useCounter.getState().inc();
    expect(useCounter.getState().count).toBe(1);
    useCounter.getState().reset();
    expect(useCounter.getState().count).toBe(0);
  });

  it('passes get to the initializer', () => {
    const useCounter = createVyzorStore<CounterState>('CounterStore', (set, get) => ({
      count: 5,
      inc: () => set({ count: get().count + 1 }),
      reset: () => set({ count: 0 }),
    }));

    expect(useCounter.getState().count).toBe(5);
    useCounter.getState().inc();
    expect(useCounter.getState().count).toBe(6);
  });

  it('persists state when persist options are provided', () => {
    const useCounter = createVyzorStore<CounterState>(
      'PersistedCounterStore',
      (set) => ({
        count: 0,
        inc: () => set((s) => ({ count: s.count + 1 })),
        reset: () => set({ count: 0 }),
      }),
      {
        persist: {
          name: 'vyzor-test-counter',
          partialize: (state) => ({ count: state.count }),
        },
      },
    );

    useCounter.getState().inc();
    expect(useCounter.getState().count).toBe(1);

    const stored = JSON.parse(localStorage.getItem('vyzor-test-counter') ?? '{}');
    expect(stored.state.count).toBe(1);
  });

  it('hydrates persisted state on creation', async () => {
    localStorage.setItem(
      'vyzor-test-counter',
      JSON.stringify({ state: { count: 42 }, version: 0 }),
    );

    const useCounter = createVyzorStore<CounterState>(
      'PersistedCounterStore',
      (set) => ({
        count: 0,
        inc: () => set((s) => ({ count: s.count + 1 })),
        reset: () => set({ count: 0 }),
      }),
      {
        persist: {
          name: 'vyzor-test-counter',
          partialize: (state) => ({ count: state.count }),
        },
      },
    );

    await vi.waitFor(() => {
      expect(useCounter.getState().count).toBe(42);
    });
  });

  it('uses devtoolsName override when provided', () => {
    const useCounter = createVyzorStore<CounterState>(
      'InternalName',
      (set) => ({
        count: 0,
        inc: () => set((s) => ({ count: s.count + 1 })),
        reset: () => set({ count: 0 }),
      }),
      { devtoolsName: 'CustomDevtoolsName' },
    );

    expect(useCounter.getState().count).toBe(0);
  });
});
