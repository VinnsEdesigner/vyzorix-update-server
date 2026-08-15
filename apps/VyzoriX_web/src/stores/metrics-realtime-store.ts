import { create } from 'zustand';
import type { TelemetryFrame } from '@vyzorix/api-client';

const DEFAULT_WINDOW_MS = 6 * 60 * 60 * 1000;

export interface MetricPoint {
  t: number;
  riskScore: number;
  thermalTemp: number;
  bufferLevel: number;
  uptime: number;
}

export type MetricKey = 'riskScore' | 'thermalTemp' | 'bufferLevel' | 'uptime';

export interface MetricsRealtimeState {
  byDevice: Record<string, MetricPoint[]>;
  windowMs: number;
  lastFrame: Record<string, TelemetryFrame | null>;
  activeOrganizationId: string | null;
  push: (deviceId: string, frame: TelemetryFrame) => void;
  setWindow: (ms: number) => void;
  clear: (deviceId?: string) => void;
  setActiveOrganization: (organizationId: string | null) => void;
  getSeries: (deviceId: string, metric: MetricKey) => { t: number; value: number }[];
  getLastFrame: (deviceId: string) => TelemetryFrame | null;
}

export const useMetricsRealtimeStore = create<MetricsRealtimeState>((set, get) => ({
  byDevice: {},
  windowMs: DEFAULT_WINDOW_MS,
  lastFrame: {},
  activeOrganizationId: null,

  push: (deviceId, frame) => {
    const t = frame.timestamp instanceof Date ? frame.timestamp.getTime() : Date.now();
    const point: MetricPoint = {
      t,
      riskScore: frame.riskScore,
      thermalTemp: frame.thermalTemp,
      bufferLevel: frame.bufferLevel,
      uptime: 0,
    };
    set((state) => {
      const existing = state.byDevice[deviceId] ?? [];
      const cutoff = t - state.windowMs;
      const next = [...existing, point].filter((p) => p.t >= cutoff);
      return {
        byDevice: { ...state.byDevice, [deviceId]: next },
        lastFrame: { ...state.lastFrame, [deviceId]: frame },
      };
    });
  },

  setWindow: (ms) => {
    if (ms <= 0) return;
    set({ windowMs: ms });
  },

  clear: (deviceId) =>
    set((state) => {
      if (!deviceId) return { byDevice: {}, lastFrame: {} };
      const byDevice = { ...state.byDevice };
      const lastFrame = { ...state.lastFrame };
      delete byDevice[deviceId];
      delete lastFrame[deviceId];
      return { byDevice, lastFrame };
    }),

  setActiveOrganization: (organizationId) =>
    set((state) => {
      if (state.activeOrganizationId === organizationId) return state;
      return { activeOrganizationId: organizationId, byDevice: {}, lastFrame: {} };
    }),

  getSeries: (deviceId, metric) => {
    const state = get();
    return (state.byDevice[deviceId] ?? []).map((p) => ({ t: p.t, value: p[metric] }));
  },

  getLastFrame: (deviceId) => get().lastFrame[deviceId] ?? null,
}));
