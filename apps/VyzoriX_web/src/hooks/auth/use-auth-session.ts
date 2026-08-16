import { useEffect, useCallback, useRef } from 'react';
import { authContext } from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';

const SESSION_TIMEOUT_MS = 30 * 60 * 1000; // 30 min inactivity → logout

const ACTIVITY_EVENTS = ['mousedown', 'keydown', 'scroll', 'touchstart'] as const;

export interface UseAuthSessionOptions {
  /** Inactivity duration before auto-logout (default 30 min). */
  sessionTimeoutMs?: number;
  /** Disable the activity watcher (e.g. for tests). */
  disableActivityWatch?: boolean;
}

export interface UseAuthSessionResult {
  refreshSession: () => Promise<void>;
}

/**
 * Session management for an authenticated operator.
 *
 - `authContext` already schedules proactive token refresh before expiry
   (see `scheduleTokenRefresh`), so this hook does NOT duplicate that. It adds:
   1. An activity-based session timeout — after `sessionTimeoutMs` of no user
      input, the session is cleared (auto-logout) as a defence-in-depth measure.
   2. A `refreshSession` action that delegates to `authContext.refreshTokens`
      and syncs the store.
 *
 * The watcher only runs while authenticated, so logged-out users incur no
 * timers. SSR-safe: no `window` access during server render.
 */
export function useAuthSession(
  options: UseAuthSessionOptions = {},
): UseAuthSessionResult {
  const { sessionTimeoutMs = SESSION_TIMEOUT_MS, disableActivityWatch = false } = options;
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated);
  const clear = useAuthStore((s) => s.clear);
  const refreshTokens = useAuthStore((s) => s.refreshTokens);
  const activityTimerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  const refreshSession = useCallback(async () => {
    await refreshTokens();
  }, [refreshTokens]);

  useEffect(() => {
    if (disableActivityWatch || !isAuthenticated || typeof window === 'undefined') {
      return;
    }

    const resetTimer = () => {
      if (activityTimerRef.current) {
        clearTimeout(activityTimerRef.current);
      }
      activityTimerRef.current = setTimeout(() => {
        clear();
      }, sessionTimeoutMs);
    };

    ACTIVITY_EVENTS.forEach((event) => window.addEventListener(event, resetTimer));
    resetTimer();

    return () => {
      ACTIVITY_EVENTS.forEach((event) => window.removeEventListener(event, resetTimer));
      if (activityTimerRef.current) {
        clearTimeout(activityTimerRef.current);
        activityTimerRef.current = null;
      }
    };
  }, [isAuthenticated, sessionTimeoutMs, disableActivityWatch, clear]);

  // Reconcile store on token-refresh events emitted by authContext.
  useEffect(() => {
    const unsubscribe = authContext.onTokenRefresh(() => {
      // authContext already mutates its own state; the store's onChange
      // subscription mirrors it. This listener is a no-op hook kept for
      // future extension (e.g. telemetry on refresh).
    });
    return unsubscribe;
  }, []);

  return { refreshSession };
}
