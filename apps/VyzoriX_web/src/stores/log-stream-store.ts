import { create } from 'zustand';
import type { LogEntry, LogEventType } from '@vyzorix/api-client';

const DEFAULT_MAX_ENTRIES = 500;

export interface LogStreamFilters {
  type?: LogEventType;
  search?: string;
}

export interface LogStreamState {
  byDevice: Record<string, LogEntry[]>;
  filters: LogStreamFilters;
  autoScroll: boolean;
  activeOrganizationId: string | null;
  append: (deviceId: string, entry: LogEntry) => void;
  appendBatch: (deviceId: string, entries: LogEntry[]) => void;
  setFilter: (filter: Partial<LogStreamFilters>) => void;
  toggleAutoScroll: () => void;
  clear: (deviceId?: string) => void;
  trim: (deviceId: string, max?: number) => void;
  setActiveOrganization: (organizationId: string | null) => void;
  getEntries: (deviceId: string) => LogEntry[];
}

function applyFilters(entries: LogEntry[], filters: LogStreamFilters): LogEntry[] {
  return entries.filter((entry) => {
    if (filters.type && entry.eventType !== filters.type) return false;
    if (filters.search) {
      const haystack = `${entry.id} ${entry.eventType} ${JSON.stringify(entry.data ?? {})}`.toLowerCase();
      if (!haystack.includes(filters.search.toLowerCase())) return false;
    }
    return true;
  });
}

export const useLogStreamStore = create<LogStreamState>((set, get) => ({
  byDevice: {},
  filters: {},
  autoScroll: true,
  activeOrganizationId: null,

  append: (deviceId, entry) => {
    set((state) => {
      const existing = state.byDevice[deviceId] ?? [];
      const next = [...existing, entry];
      if (next.length > DEFAULT_MAX_ENTRIES) {
        next.splice(0, next.length - DEFAULT_MAX_ENTRIES);
      }
      return { byDevice: { ...state.byDevice, [deviceId]: next } };
    });
  },

  appendBatch: (deviceId, entries) => {
    if (entries.length === 0) return;
    set((state) => {
      const existing = state.byDevice[deviceId] ?? [];
      const next = [...existing, ...entries];
      if (next.length > DEFAULT_MAX_ENTRIES) {
        next.splice(0, next.length - DEFAULT_MAX_ENTRIES);
      }
      return { byDevice: { ...state.byDevice, [deviceId]: next } };
    });
  },

  setFilter: (filter) => set((state) => ({ filters: { ...state.filters, ...filter } })),

  toggleAutoScroll: () => set((state) => ({ autoScroll: !state.autoScroll })),

  clear: (deviceId) =>
    set((state) => {
      if (!deviceId) return { byDevice: {} };
      const byDevice = { ...state.byDevice };
      delete byDevice[deviceId];
      return { byDevice };
    }),

  trim: (deviceId, max = DEFAULT_MAX_ENTRIES) =>
    set((state) => {
      const existing = state.byDevice[deviceId] ?? [];
      if (existing.length <= max) return state;
      const next = existing.slice(existing.length - max);
      return { byDevice: { ...state.byDevice, [deviceId]: next } };
    }),

  setActiveOrganization: (organizationId) =>
    set((state) => {
      if (state.activeOrganizationId === organizationId) return state;
      return { activeOrganizationId: organizationId, byDevice: {} };
    }),

  getEntries: (deviceId) => {
    const state = get();
    return applyFilters(state.byDevice[deviceId] ?? [], state.filters);
  },
}));
