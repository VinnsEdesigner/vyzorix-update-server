import { useEffect, useState } from 'react';
import {
  oauth,
  type OAuthCallbackResult,
} from '@vyzorix/api-client';
import { useAuthStore } from '@/stores/auth-store';
import { useMe } from './use-me';

export interface UseAuthCallbackOptions {
  /** The callback URL to parse (defaults to `window.location.href`). */
  url?: string;
  /** Where to send the user after a successful callback (default `/dashboard`). */
  redirectOnSuccess?: string;
  /** Whether to fetch `/me` after a successful callback to hydrate state. */
  fetchMe?: boolean;
}

export interface AuthCallbackState {
  result: OAuthCallbackResult | null;
  isLoading: boolean;
  error: string | null;
}

/**
 * OAuth callback handler. Parses the provider-redirect URL the server sent the
 * browser back to, classifies success/error, and — on success — hydrates auth
 * state by fetching `/me` (the server has already established the session via
 * a cookie / tokens before redirecting).
 *
 * Pure parsing + side-effect orchestration; no UI. The component reads
 * `result` / `error` / `isLoading` to render the appropriate state and route.
 */
export function useAuthCallback(options: UseAuthCallbackOptions = {}): AuthCallbackState {
  const { url, fetchMe = true } = options;
  const [state, setState] = useState<AuthCallbackState>({
    result: null,
    isLoading: true,
    error: null,
  });

  const setFromMeResponse = useAuthStore((s) => s.setFromMeResponse);
  const meQuery = useMe({ enabled: false });
  // Extract the stable refetch fn so the effect doesn't re-run on every render
  // (the full `meQuery` result object changes identity each render, which would
  // cause an infinite effect → refetch → re-render loop).
  const meRefetch = meQuery.refetch;

  useEffect(() => {
    let cancelled = false;
    const targetUrl = url ?? (typeof window !== 'undefined' ? window.location.href : '');

    async function run() {
      if (!targetUrl) {
        setState({ result: null, isLoading: false, error: 'No callback URL' });
        return;
      }
      const parsed = oauth.parseCallback(targetUrl);
      if (cancelled) return;

      if (!parsed.success) {
        setState({
          result: parsed,
          isLoading: false,
          error: parsed.errorMessage ?? 'OAuth authentication failed',
        });
        return;
      }

      if (fetchMe) {
        try {
          const me = await meRefetch();
          if (cancelled) return;
          if (me.data) {
            setFromMeResponse(me.data);
          }
        } catch {
          // Session may still be valid via cookie; non-fatal.
        }
      }

      setState({ result: parsed, isLoading: false, error: null });
    }

    run();
    return () => {
      cancelled = true;
    };
  }, [url, fetchMe, meRefetch, setFromMeResponse]);

  return state;
}

/**
 * Initiate a Google OAuth login (redirects the browser to the server's
 * `/v1/auth/google` endpoint, which then bounces to Google).
 */
export function useInitiateGoogleLogin() {
  return (redirectAfterCallback?: string) =>
    oauth.initiateGoogleLogin({ redirectAfterCallback });
}

/**
 * Initiate a GitHub OAuth login.
 */
export function useInitiateGitHubLogin() {
  return (redirectAfterCallback?: string) =>
    oauth.initiateGitHubLogin({ redirectAfterCallback });
}
