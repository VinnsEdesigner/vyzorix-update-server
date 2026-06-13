/**
 * use-operator.ts - SSR-aware authentication hook
 *
 * This hook provides access to the authenticated operator from either:
 * 1. Server-injected state (SSR hydration)
 * 2. API fallback (for client-side navigation)
 *
 * Based on Library's getHydratedState pattern
 */

import { useState, useEffect } from "react";

import { getFullHydratedState } from "@/lib/server/state-injector";

export interface Operator {
  id: string;
  email: string;
  name: string;
  role: "viewer" | "operator" | "super_admin";
  createdAt?: number;
  emailVerified?: boolean;
  thresholds?: Record<string, number>;
}

export interface AuthContext {
  operator: Operator | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

/**
 * SSR-aware auth hook
 *
 * This replaces localStorage-based auth checking with server-provided state hydration.
 *
 * Library's pattern (src/lib/config.ts):
 *   export function getHydratedState<T>(key: string, defaultValue: T): T {
 *     if (typeof window !== 'undefined') {
 *       const globalState = (window as any).__VYZORIX_PREFETCHED_STATE__;
 *       if (globalState && globalState[key] !== undefined) {
 *         return globalState[key];
 *       }
 *     }
 *     return defaultValue;
 *   }
 */
export function useOperator(): AuthContext {
  const [operator, setOperator] = useState<Operator | null>(null);
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [isLoading, setIsLoading] = useState(true);

  useEffect(() => {
    // Try to get state from server injection first (SSR hydration)
    const globalState = getFullHydratedState();

    if (globalState) {
      // Server provided the auth state - use it
      setOperator(globalState.operator ?? null);
      setIsAuthenticated(globalState.isAuthenticated ?? false);
      setIsLoading(false);
      return;
    }

    // Fallback: Fetch from API if SSR state not available
    // This handles client-side navigation without SSR state
    // (e.g., direct page navigation, or SSR state not injected)
    fetch("/v1/auth/me", { credentials: "include" })
      .then((res) => {
        if (res.ok) {
          return res.json();
        }
        return null;
      })
      .then((data) => {
        setOperator(data);
        setIsAuthenticated(Boolean(data));
      })
      .catch(() => {
        setIsAuthenticated(false);
      })
      .finally(() => {
        setIsLoading(false);
      });
  }, []);

  return { operator, isAuthenticated, isLoading };
}

/**
 * Hook to check if user is authenticated (simpler version)
 */
export function useIsAuthenticated(): boolean {
  const { isAuthenticated, isLoading } = useOperator();
  return !isLoading && isAuthenticated;
}

/**
 * Hook to get the current operator (simpler version)
 */
export function useCurrentOperator(): Operator | null {
  const { operator } = useOperator();
  return operator;
}

/**
 * Fetch operator from API with automatic cookie forwarding
 * Useful for manual auth checks
 */
export async function fetchOperator(): Promise<Operator | null> {
  try {
    const res = await fetch("/v1/auth/me", { credentials: "include" });
    if (res.ok) {
      return res.json();
    }
    return null;
  } catch {
    return null;
  }
}

/**
 * Check if there's an authenticated session
 * First checks server state, falls back to API call
 */
export async function checkAuth(): Promise<boolean> {
  // Check server-injected state first
  const state = getFullHydratedState();
  if (state) {
    return state.isAuthenticated;
  }

  // Fallback to API call
  const operator = await fetchOperator();
  return operator !== null;
}
