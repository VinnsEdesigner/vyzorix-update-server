import { create } from 'zustand';
import type {
  DeviceInspection,
  ConnectionInfo,
  TelemetryInfo,
} from '@vyzorix/api-client';

const DEFAULT_REFRESH_INTERVAL_MS = 10_000;

export interface DiagnosticsState {
  snapshots: Record<string, DeviceInspection | undefined>;
  lastRefreshedAt: Record<string, number | null>;
  isRefreshing: Record<string, boolean>;
  refreshIntervalMs: number;
  isPolling: boolean;
  activeOrganizationId: string | null;

  getSnapshot: (organizationId: string, imei: string) => DeviceInspection | undefined;
  setSnapshot: (organizationId: string, imei: string, data: DeviceInspection) => void;
  patchConnection: (
    organizationId: string,
    imei: string,
    patch: Partial<ConnectionInfo>,
  ) => void;
  patchTelemetry: (
    organizationId: string,
    imei: string,
    patch: Partial<TelemetryInfo>,
  ) => void;
  setRefreshing: (organizationId: string, imei: string, v: boolean) => void;
  setRefreshInterval: (ms: number) => void;
  setPolling: (v: boolean) => void;
  setActiveOrganization: (organizationId: string | null) => void;
  clear: (organizationId: string, imei: string) => void;
  clearAll: () => void;
}

function key(organizationId: string, imei: string): string {
  return `${organizationId}:${imei}`;
}

export const useDiagnosticsStore = create<DiagnosticsState>((set, get) => ({
  snapshots: {},
  lastRefreshedAt: {},
  isRefreshing: {},
  refreshIntervalMs: DEFAULT_REFRESH_INTERVAL_MS,
  isPolling: false,
  activeOrganizationId: null,

  getSnapshot: (organizationId, imei) =>
    get().snapshots[key(organizationId, imei)],

  setSnapshot: (organizationId, imei, data) =>
    set((state) => {
      const k = key(organizationId, imei);
      return {
        snapshots: { ...state.snapshots, [k]: data },
        lastRefreshedAt: { ...state.lastRefreshedAt, [k]: Date.now() },
        isRefreshing: { ...state.isRefreshing, [k]: false },
      };
    }),

  patchConnection: (organizationId, imei, patch) =>
    set((state) => {
      const k = key(organizationId, imei);
      const current = state.snapshots[k];
      if (!current) return state;
      return {
        snapshots: {
          ...state.snapshots,
          [k]: {
            ...current,
            connection: { ...current.connection, ...patch },
          },
        },
      };
    }),

  patchTelemetry: (organizationId, imei, patch) =>
    set((state) => {
      const k = key(organizationId, imei);
      const current = state.snapshots[k];
      if (!current) return state;
      return {
        snapshots: {
          ...state.snapshots,
          [k]: {
            ...current,
            telemetry: { ...current.telemetry, ...patch },
          },
        },
      };
    }),

  setRefreshing: (organizationId, imei, v) =>
    set((state) => {
      const k = key(organizationId, imei);
      return { isRefreshing: { ...state.isRefreshing, [k]: v } };
    }),

  setRefreshInterval: (ms) => set({ refreshIntervalMs: ms }),

  setPolling: (v) => set({ isPolling: v }),

  setActiveOrganization: (organizationId) =>
    set((state) => {
      if (state.activeOrganizationId === organizationId) return state;
      // Clear all snapshots on org switch (defense-in-depth: never leak another org's data).
      return {
        activeOrganizationId: organizationId,
        snapshots: {},
        lastRefreshedAt: {},
        isRefreshing: {},
      };
    }),

  clear: (organizationId, imei) =>
    set((state) => {
      const k = key(organizationId, imei);
      const snapshots = { ...state.snapshots };
      const lastRefreshedAt = { ...state.lastRefreshedAt };
      const isRefreshing = { ...state.isRefreshing };
      delete snapshots[k];
      delete lastRefreshedAt[k];
      delete isRefreshing[k];
      return { snapshots, lastRefreshedAt, isRefreshing };
    }),

  clearAll: () =>
    set({ snapshots: {}, lastRefreshedAt: {}, isRefreshing: {} }),
}));
