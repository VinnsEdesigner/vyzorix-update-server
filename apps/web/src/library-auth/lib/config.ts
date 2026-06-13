/**
 * Vyzorix System Architecture Settings
 *
 * Toggle the entire frontend stack's routing, hydration behavior, and backend
 * connectivity with a single line of configuration.
 */
export const ARCHITECTURE_CONFIG = {
  /**
   * ARCHITECTURE MODE
   *
   * 'SPA' - Pure Single-Page Application (Vite client-side router, local state).
   * 'SSR' - React TanStack SSR & Nitro HTML Hydration mode. In this mode, initial
   *         values (like pre-authenticated operator profiles) are hydrated from the Go server.
   */
  MODE: "SSR" as "SPA" | "SSR",

  /**
   * CONNECTION SIMULATION
   *
   * true  - Runs low-latency high-fidelity mock handshakes.
   * false - Directly triggers API route requests targeting your Go API backend.
   */
  IS_SIMULATED: false,

  /**
   * ENDPOINTS & DIRECTORIES
   */
  API_BASE_URL: "/api/auth",
  GO_BACKEND_SERVER: "http://localhost:8080",
};

/**
 * SSR Hybrid Hydration helper
 * Retrieves prefetched server-side page state injected by Nitro/Go before JavaScript loads.
 */
export function getHydratedState<T>(key: string, defaultValue: T): T {
  if (typeof window !== "undefined") {
    const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
    if (globalState?.[key] !== undefined) {
      return globalState[key];
    }
  }
  return defaultValue;
}
