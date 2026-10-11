export { restClient, axios } from './rest-client';

export {
  VyzorixRequestQueue,
  DEFAULT_REQUEST_QUEUE_CONFIG,
  type RequestQueueConfig,
} from './request-queue';

export {
  setAuthToken,
  setCSRFToken,
  setRefreshToken,
  setSigningKey,
  getSigningKey,
  setOrganizationContext,
  clearAuthContext,
  fetchAndSetCSRFToken,
  getCSRFToken,
  resetClientState,
  resetCircuitBreaker,
  clearInFlightRequests,
} from './rest-client';

export {
  initConnectivityMonitor,
  getConnectivityMonitor,
  createConnectivityMonitor,
  DEFAULT_CONNECTIVITY_CONFIG,
  type ConnectivityMonitor,
  type NetworkState,
  type QueuedRequest,
  type ConnectivityConfig,
  type ConnectivityCallback,
} from '../_connectivity';
