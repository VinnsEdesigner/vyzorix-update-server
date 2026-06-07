/**
 * use-auth.ts — React hooks for authentication
 *
 * Provides authentication state and actions for the React app.
 */

import { useCallback, useEffect, useState } from "react";

import { logger } from "@/lib/logger";
import { getToken, getStoredOperator, me, logout, type Operator } from "@/lib/vyzorix-auth";
import { useVyzorixConfig } from "@/lib/vyzorix-config";

export interface AuthState {
  isAuthenticated: boolean;
  isLoading: boolean;
  operator: Operator | null;
  token: string | null;
}

export interface AuthActions {
  checkAuth: () => Promise<boolean>;
  refreshOperator: () => Promise<Operator | null>;
  signOut: () => Promise<void>;
}

/**
 * useAuth hook - provides authentication state
 */
// eslint-disable-next-line func-style
export function useAuth(): AuthState {
  const [state, setState] = useState<AuthState>(() => {
    const token = getToken();
    const operator = getStoredOperator();
    return {
      isAuthenticated: Boolean(token) && Boolean(operator),
      isLoading: false,
      operator,
      token,
    };
  });

  // Update state when localStorage changes (for cross-tab sync)
  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const handleStorage = () => {
      const token = getToken();
      const operator = getStoredOperator();
      setState({
        isAuthenticated: Boolean(token) && Boolean(operator),
        isLoading: false,
        operator,
        token,
      });
    };

    window.addEventListener("storage", handleStorage);
    return () => window.removeEventListener("storage", handleStorage);
  }, []);

  return state;
}

/**
 * useAuthActions hook - provides authentication actions
 */
// eslint-disable-next-line func-style
export function useAuthActions(): AuthActions {
  const { operator: _operator } = useAuth();
  const { serverUrl } = useVyzorixConfig();

  const checkAuth = useCallback(async (): Promise<boolean> => {
    const token = getToken();
    if (!token) return false;

    try {
      const op = await me(serverUrl);
      logger.info("auth", "Session validated", { email: op.email });
      return true;
    } catch (err) {
      logger.warn("auth", "Session validation failed", {
        error: err instanceof Error ? err.message : String(err),
      });
      return false;
    }
  }, [serverUrl]);

  const refreshOperator = useCallback(async (): Promise<Operator | null> => {
    try {
      const op = await me(serverUrl);
      return op;
    } catch (err) {
      logger.warn("auth", "Failed to refresh operator", {
        error: err instanceof Error ? err.message : String(err),
      });
      return null;
    }
  }, [serverUrl]);

  const signOut = useCallback(async (): Promise<void> => {
    try {
      await logout(serverUrl);
      logger.info("auth", "Signed out");
    } catch (err) {
      logger.warn("auth", "Sign out error", {
        error: err instanceof Error ? err.message : String(err),
      });
    }
  }, [serverUrl]);

  return { checkAuth, refreshOperator, signOut };
}

/**
 * useAuthGuard hook - protects routes that require authentication
 *
 * Returns loading state while checking auth, and a redirect function.
 */

// eslint-disable-next-line func-style
export function useAuthGuard(): object {
  const { isAuthenticated, isLoading } = useAuth();
  const { checkAuth } = useAuthActions();

  const validate = useCallback(async (): Promise<boolean> => {
    // First check local state
    if (!getToken()) return false;

    // Then validate with server
    // eslint-disable-next-line no-return-await
    return await checkAuth();
  }, [checkAuth]);

  return {
    isAuthenticated,
    isLoading,
    validate,
  };
}

/**
 * useRequireAuth hook - for components that need to wait for auth check
 */

// eslint-disable-next-line func-style
export function useRequireAuth(): { isReady: boolean; isAuthenticated: boolean } {
  const [state, setState] = useState<{
    isReady: boolean;
    isAuthenticated: boolean;
  }>({
    isReady: false,
    isAuthenticated: false,
  });
  const { checkAuth } = useAuthActions();

  useEffect(() => {
    // eslint-disable-next-line @typescript-eslint/explicit-function-return-type
    const check = async () => {
      const isAuth = await checkAuth();
      setState({ isReady: true, isAuthenticated: isAuth });
    };
    check();
  }, [checkAuth]);

  return state;
}
