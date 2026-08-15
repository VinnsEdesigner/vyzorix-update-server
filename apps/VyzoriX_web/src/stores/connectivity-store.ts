import { create } from 'zustand';
import {
  initConnectivityMonitor,
  getConnectivityMonitor,
  type NetworkState,
  type QueuedRequest,
} from '@vyzorix/api-client';

export interface ConnectivityStoreState extends NetworkState {
  queueSize: number;
  queuedRequests: QueuedRequest[];
  isChecking: boolean;
  isFlushing: boolean;
  checkConnectivity: () => Promise<boolean>;
  flushQueue: () => Promise<void>;
  clearQueue: () => void;
}

const monitor = initConnectivityMonitor();

export const useConnectivityStore = create<ConnectivityStoreState>((set) => {
  const sync = (state: NetworkState) => {
    const m = getConnectivityMonitor();
    set({
      isOnline: state.isOnline,
      wasOnline: state.wasOnline,
      lastChecked: state.lastChecked,
      effectiveType: state.effectiveType,
      downlink: state.downlink,
      rtt: state.rtt,
      queueSize: m.getQueueSize(),
      queuedRequests: m.getQueuedRequests(),
    });
  };

  monitor.subscribe(sync);
  sync(monitor.getState());

  return {
    isOnline: monitor.getState().isOnline,
    wasOnline: monitor.getState().wasOnline,
    lastChecked: monitor.getState().lastChecked,
    effectiveType: monitor.getState().effectiveType,
    downlink: monitor.getState().downlink,
    rtt: monitor.getState().rtt,
    queueSize: monitor.getQueueSize(),
    queuedRequests: monitor.getQueuedRequests(),
    isChecking: false,
    isFlushing: false,
    checkConnectivity: async () => {
      set({ isChecking: true });
      try {
        return await monitor.checkConnectivity();
      } finally {
        set({ isChecking: false });
      }
    },
    flushQueue: async () => {
      set({ isFlushing: true });
      try {
        await monitor.flushQueue();
      } finally {
        set({ isFlushing: false });
      }
    },
    clearQueue: () => {
      monitor.clearQueue();
    },
  };
});
