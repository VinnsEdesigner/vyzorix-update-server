import { useCallback, useMemo } from 'react';
import { useTimelineStreamStore } from '@/stores';
import type {
  TimelineEventCategory,
  TimelineEventType,
} from '@vyzorix/api-client';
import type { TimelineParams } from './use-diagnostics';

export interface TimelineFilterState {
  category: TimelineEventCategory | 'all';
  rangeMs: number | null; // null = no time-window filter (server default 24h)
  autoScroll: boolean;
}

export interface TimelineFilterActions {
  setCategory: (category: TimelineEventCategory | 'all') => void;
  setRangeMs: (rangeMs: number | null) => void;
  toggleAutoScroll: () => void;
  clear: () => void;
  /** Derives the REST/GraphQL query params from the active filter state. */
  toQueryParams: () => TimelineParams;
}

const DEFAULT_STATE: TimelineFilterState = {
  category: 'all',
  rangeMs: null,
  autoScroll: true,
};

const CATEGORY_TO_EVENT_TYPE: Partial<
  Record<TimelineEventCategory, TimelineEventType>
> = {
  // The server accepts an exact TimelineEventType for `event_type`; a category filter
  // narrows the client-side stream, but the historical query still fetches all events
  // in the window (the UI filters the merged stream by category). Returning undefined
  // here means "fetch everything in the window".
  telemetry: undefined,
  command: undefined,
  connection: undefined,
  error: undefined,
};

export function useTimelineFilter(): TimelineFilterState & TimelineFilterActions {
  const category = useTimelineStreamStore((s) => s.filters.category);
  const rangeMs = useTimelineStreamStore((s) => s.filters.rangeMs);
  const autoScroll = useTimelineStreamStore((s) => s.autoScroll);
  const setFilter = useTimelineStreamStore((s) => s.setFilter);
  const toggleAutoScroll = useTimelineStreamStore((s) => s.toggleAutoScroll);
  const clearStream = useTimelineStreamStore((s) => s.clear);

  const state = useMemo<TimelineFilterState>(
    () => ({
      category: (category ?? 'all') as TimelineEventCategory | 'all',
      rangeMs: rangeMs ?? null,
      autoScroll,
    }),
    [category, rangeMs, autoScroll],
  );

  const setCategory = useCallback(
    (next: TimelineEventCategory | 'all') => {
      setFilter({ category: next === 'all' ? undefined : next });
    },
    [setFilter],
  );

  const setRangeMs = useCallback(
    (next: number | null) => {
      setFilter({ rangeMs: next ?? undefined });
    },
    [setFilter],
  );

  const clear = useCallback(() => {
    setFilter({ category: undefined, rangeMs: undefined });
    clearStream();
  }, [setFilter, clearStream]);

  const toQueryParams = useCallback((): TimelineParams => {
    const now = Date.now();
    const startTime =
      state.rangeMs != null ? now - state.rangeMs : now - 24 * 60 * 60 * 1000;
    const activeCategory =
      state.category === 'all' ? undefined : (state.category as TimelineEventCategory);
    return {
      eventType: activeCategory
        ? CATEGORY_TO_EVENT_TYPE[activeCategory]
        : undefined,
      startTime,
      endTime: now,
      limit: 50,
    };
  }, [state]);

  return {
    ...DEFAULT_STATE,
    ...state,
    setCategory,
    setRangeMs,
    toggleAutoScroll,
    clear,
    toQueryParams,
  };
}
