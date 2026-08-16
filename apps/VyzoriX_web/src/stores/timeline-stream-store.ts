import { createVyzorStore } from '@/lib/state';
import type {
  TimelineEvent,
  TimelineEventCategory,
  TimelineEventType,
} from '@vyzorix/api-client';
import { getEventCategory } from '@vyzorix/api-client';

const DEFAULT_MAX_EVENTS = 500;

export interface TimelineStreamFilters {
  category?: TimelineEventCategory;
  rangeMs?: number;
}

export interface TimelineStreamState {
  byDevice: Record<string, TimelineEvent[]>;
  filters: TimelineStreamFilters;
  autoScroll: boolean;
  activeOrganizationId: string | null;
  append: (imei: string, event: TimelineEvent) => void;
  appendBatch: (imei: string, events: TimelineEvent[]) => void;
  setFilter: (filter: Partial<TimelineStreamFilters>) => void;
  toggleAutoScroll: () => void;
  clear: (imei?: string) => void;
  trim: (imei: string, max?: number) => void;
  setActiveOrganization: (organizationId: string | null) => void;
  getEvents: (imei: string) => TimelineEvent[];
}

function applyFilters(
  events: TimelineEvent[],
  filters: TimelineStreamFilters,
): TimelineEvent[] {
  const now = Date.now();
  return events.filter((event) => {
    if (filters.category) {
      const cat = categoryOf(event.type);
      if (cat !== filters.category) return false;
    }
    if (filters.rangeMs != null) {
      if (event.timestamp.getTime() < now - filters.rangeMs) return false;
    }
    return true;
  });
}

function categoryOf(type: TimelineEventType): TimelineEventCategory {
  return getEventCategory(type);
}

export const useTimelineStreamStore = createVyzorStore<TimelineStreamState>('TimelineStreamStore', (set, get) => ({
  byDevice: {},
  filters: {},
  autoScroll: true,
  activeOrganizationId: null,

  // Newest-first (matches the server's desc timestamp+id ordering) at the head.
  append: (imei, event) => {
    set((state) => {
      const existing = state.byDevice[imei] ?? [];
      const next = [event, ...existing];
      if (next.length > DEFAULT_MAX_EVENTS) {
        next.length = DEFAULT_MAX_EVENTS;
      }
      return { byDevice: { ...state.byDevice, [imei]: next } };
    });
  },

  appendBatch: (imei, events) => {
    if (events.length === 0) return;
    set((state) => {
      const existing = state.byDevice[imei] ?? [];
      // Merge, dedupe by id, keep newest-first ordering.
      const seen = new Set(existing.map((e) => e.id));
      const incoming = events.filter((e) => !seen.has(e.id));
      const next = [...incoming, ...existing];
      if (next.length > DEFAULT_MAX_EVENTS) {
        next.length = DEFAULT_MAX_EVENTS;
      }
      return { byDevice: { ...state.byDevice, [imei]: next } };
    });
  },

  setFilter: (filter) =>
    set((state) => ({ filters: { ...state.filters, ...filter } })),

  toggleAutoScroll: () => set((state) => ({ autoScroll: !state.autoScroll })),

  clear: (imei) =>
    set((state) => {
      if (!imei) return { byDevice: {} };
      const byDevice = { ...state.byDevice };
      delete byDevice[imei];
      return { byDevice };
    }),

  trim: (imei, max = DEFAULT_MAX_EVENTS) =>
    set((state) => {
      const existing = state.byDevice[imei] ?? [];
      if (existing.length <= max) return state;
      return { byDevice: { ...state.byDevice, [imei]: existing.slice(0, max) } };
    }),

  setActiveOrganization: (organizationId) =>
    set((state) => {
      if (state.activeOrganizationId === organizationId) return state;
      return { activeOrganizationId: organizationId, byDevice: {} };
    }),

  getEvents: (imei) => {
    const state = get();
    return applyFilters(state.byDevice[imei] ?? [], state.filters);
  },
}));
