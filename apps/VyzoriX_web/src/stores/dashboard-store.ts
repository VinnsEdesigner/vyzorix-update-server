import { create } from 'zustand';
import type { DashboardStats } from '@vyzorix/api-client';

const DEFAULT_REFRESH_INTERVAL_MS = 30_000;
const MAX_ACTIVITY_ITEMS = 100;

export interface ActivityItem {
  id: string;
  type: 'command_sent' | 'log_alert' | 'device_online' | 'device_offline' | 'metric';
  message: string;
  timestamp: number;
  deviceId?: string;
}

export interface DashboardStoreState {
  stats: DashboardStats | null;
  lastRefreshedAt: number | null;
  isRefreshing: boolean;
  refreshIntervalMs: number;
  isPolling: boolean;
  recentActivity: ActivityItem[];
  activeOrganizationId: string | null;
  setStats: (stats: DashboardStats) => void;
  setRefreshing: (isRefreshing: boolean) => void;
  setRefreshInterval: (ms: number) => void;
  setPolling: (isPolling: boolean) => void;
  pushActivity: (item: Omit<ActivityItem, 'id' | 'timestamp'> & { id?: string; timestamp?: number }) => void;
  setActiveOrganization: (organizationId: string | null) => void;
  clear: () => void;
}

export const useDashboardStore = create<DashboardStoreState>((set) => ({
  stats: null,
  lastRefreshedAt: null,
  isRefreshing: false,
  refreshIntervalMs: DEFAULT_REFRESH_INTERVAL_MS,
  isPolling: false,
  recentActivity: [],
  activeOrganizationId: null,

  setStats: (stats) => set({ stats, lastRefreshedAt: Date.now() }),

  setRefreshing: (isRefreshing) => set({ isRefreshing }),

  setRefreshInterval: (ms) => {
    if (ms > 0) set({ refreshIntervalMs: ms });
  },

  setPolling: (isPolling) => set({ isPolling }),

  pushActivity: (item) =>
    set((state) => {
      const entry: ActivityItem = {
        id: item.id ?? `act-${Date.now()}-${Math.random().toString(36).slice(2, 9)}`,
        type: item.type,
        message: item.message,
        timestamp: item.timestamp ?? Date.now(),
        deviceId: item.deviceId,
      };
      const next = [entry, ...state.recentActivity];
      if (next.length > MAX_ACTIVITY_ITEMS) next.length = MAX_ACTIVITY_ITEMS;
      return { recentActivity: next };
    }),

  setActiveOrganization: (organizationId) =>
    set((state) => {
      if (state.activeOrganizationId === organizationId) return state;
      return {
        activeOrganizationId: organizationId,
        stats: null,
        lastRefreshedAt: null,
        recentActivity: [],
      };
    }),

  clear: () =>
    set({
      stats: null,
      lastRefreshedAt: null,
      isRefreshing: false,
      recentActivity: [],
    }),
}));
