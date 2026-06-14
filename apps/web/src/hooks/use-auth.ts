/**
 * use-auth.ts — Unified authentication hook using cookie-based auth.
 *
 * This hook provides a single source of truth for authentication state,
 * using HttpOnly cookies instead of localStorage/JWT.
 *
 * Flow:
 *   - On mount: fetches /v1/auth/me with cookie credentials
 *   - Caches the operator in React state
 *   - Provides logout function that calls /v1/auth/logout
 */

import { useCallback, useEffect, useState } from "react";

import type { OperatorResponse } from "@/lib/clients/authClient";

export interface AuthState {
  operator: OperatorResponse | null;
  isAuthenticated: boolean;
  isLoading: boolean;
}

export interface AuthActions {
  refreshOperator: () => Promise<OperatorResponse | null>;
  signOut: () => Promise<void>;
}

/**
 * useAuth hook - provides authentication state and actions using cookie-based auth.
 */
export const useAuth = (): AuthState & AuthActions => {
  const [operator, setOperator] = useState<OperatorResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Fetch current session from server using cookie
  const fetchSession = useCallback(async (): Promise<OperatorResponse | null> => {
    try {
      const res = await fetch("/v1/auth/me", {
        method: "GET",
        credentials: "include",
      });

      if (!res.ok) {
        return null;
      }

      const data = (await res.json()) as OperatorResponse;
      return data;
    } catch {
      return null;
    }
  }, []);

  // Initial load - fetch session on mount
  useEffect(() => {
    const loadSession = async (): Promise<void> => {
      setIsLoading(true);

      // Fetch from server using cookie
      const session = await fetchSession();
      setOperator(session);
      setIsLoading(false);
    };

    loadSession().catch(() => {
      setIsLoading(false);
    });
  }, [fetchSession]);

  // Refresh - force re-fetch of session
  const refreshOperator = useCallback(async (): Promise<OperatorResponse | null> => {
    setIsLoading(true);
    const session = await fetchSession();
    setOperator(session);
    setIsLoading(false);
    return session;
  }, [fetchSession]);

  // Logout - call server to clear cookie, then clear local state
  const signOut = useCallback(async (): Promise<void> => {
    try {
      await fetch("/v1/auth/logout", {
        method: "POST",
        credentials: "include",
      });
    } catch {
      // ignore logout errors
    }

    // Clear local state
    setOperator(null);

    // Clear any localStorage references (for clean slate)
    try {
      localStorage.removeItem("vyz.auth.operator");
      localStorage.removeItem("vyz.auth.token");
    } catch {
      // ignore
    }
  }, []);

  return {
    operator,
    isAuthenticated: operator !== null,
    isLoading,
    refreshOperator,
    signOut,
  };
};

/**
 * useAuthGuard hook - protects routes that require authentication
 */
export const useAuthGuard = (): { isAuthenticated: boolean; isLoading: boolean } => {
  const { isAuthenticated, isLoading } = useAuth();
  return { isAuthenticated, isLoading };
};

/**
 * useRequireAuth hook - for components that need to wait for auth check
 */
export const useRequireAuth = (): { isReady: boolean; isAuthenticated: boolean } => {
  const [state, setState] = useState<{
    isReady: boolean;
    isAuthenticated: boolean;
  }>({
    isReady: false,
    isAuthenticated: false,
  });

  const { refreshOperator } = useAuth();

  useEffect(() => {
    const check = async (): Promise<void> => {
      const op = await refreshOperator();
      setState({ isReady: true, isAuthenticated: op !== null });
    };
    check();
  }, [refreshOperator]);

  return state;
};
