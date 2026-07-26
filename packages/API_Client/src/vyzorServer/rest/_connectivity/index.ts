/**
 * Connectivity Monitor Module
 * 
 * Feature 4: Connectivity & Status Monitor Logic
 */

export type {
  QueuedRequest,
  ConnectivityConfig,
  NetworkState,
  ConnectivityCallback,
  ConnectivityMonitor,
} from './connectivity.types';

export { DEFAULT_CONNECTIVITY_CONFIG } from './connectivity.types';

export {
  createConnectivityMonitor,
  getConnectivityMonitor,
  initConnectivityMonitor,
} from './connectivity-monitor';
