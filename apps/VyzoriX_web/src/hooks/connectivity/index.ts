/**
 * Connectivity Hooks
 * 
 * Feature 4: Connectivity & Status Monitor Logic
 * 
 * Usage:
 * 
 * ```tsx
 * import { useConnectivity, useIsOnline, useInitConnectivity } from './hooks/connectivity';
 * 
 * function App() {
 *   // Initialize connectivity monitoring
 *   useInitConnectivity();
 *   
 *   // Simple online status
 *   const isOnline = useIsOnline();
 *   
 *   // Or full connectivity state
 *   const { isOnline, queueSize, flushQueue, clearQueue } = useConnectivity();
 * 
 *   return (
 *     <div>
 *       <p>Status: {isOnline ? '🟢 Online' : '🔴 Offline'}</p>
 *       <p>Pending requests: {queueSize}</p>
 *       {queueSize > 0 && (
 *         <>
 *           <button onClick={flushQueue}>Retry Now</button>
 *           <button onClick={clearQueue}>Cancel All</button>
 *         </>
 *       )}
 *     </div>
 *   );
 * }
 * ```
 */

export {
  useConnectivity,
  useConnectivityState,
  useIsOnline,
  useOfflineQueueSize,
  useOfflineQueue,
  useCheckConnectivity,
  useFlushOfflineQueue,
  useClearOfflineQueue,
  useInitConnectivity,
  type UseConnectivityReturn,
  type NetworkState,
} from './use-connectivity';
