// React hook for using the signed API client.
// Provides authenticated API access with request signing and response encryption.
// Handles automatic re-authentication on SIGN_005 errors and key rotation.

import React, { useState, useCallback, useEffect, createContext, type ReactNode } from "react";

import { logger } from "../lib/logger";
import {
  SignedApiClient,
  getSignedApiClient,
  clearSignedApiClient,
  ClientCredentials,
  fetchClientCredentials,
  getActiveCredentials,
  listClientCredentials,
  deleteClientCredentials,
} from "../lib/signed-api-client";

// Error codes that require re-authentication
const ERROR_REQUIRES_REAUTH = ["SIGN_005", "SIGN_004", "SIGN_001"];
const ERROR_REQUIRES_KEY_ROTATION = ["INVALID_KEY", "KEY_EXPIRED"];

export interface UseSignedApiOptions {
  /** API base URL */
  apiUrl: string;
  /** Auto-fetch credentials on mount if session exists */
  autoFetchCredentials?: boolean;
  /** Client name for credentials */
  clientName?: string;
  /** Callback when re-authentication is needed */
  onReauthNeeded?: () => void;
  /** Callback when key rotation is needed */
  onKeyRotationNeeded?: (clientId: string) => void;
}

export interface UseSignedApiReturn {
  /** The signed API client instance */
  client: SignedApiClient;
  /** Whether we have valid credentials */
  hasCredentials: boolean;
  /** Whether credentials are being fetched */
  loading: boolean;
  /** Error if any */
  error: Error | null;
  /** Fetch new credentials (call after login) */
  fetchCredentials: () => Promise<ClientCredentials>;
  /** List stored client credentials (metadata only) */
  listCredentials: () => ReturnType<typeof listClientCredentials>;
  /** Delete a client's credentials */
  deleteCredentials: (clientId: string) => void;
  /** Clear all credentials (logout) */
  logout: () => void;
  /** Make a signed GET request with automatic error handling */
  get: <T = unknown>(path: string) => Promise<T>;
  /** Make a signed POST request with automatic error handling */
  post: <T = unknown>(path: string, body?: object) => Promise<T>;
  /** Make a signed PATCH request with automatic error handling */
  patch: <T = unknown>(path: string, body?: object) => Promise<T>;
  /** Make a signed DELETE request with automatic error handling */
  delete: <T = unknown>(path: string) => Promise<T>;
}

/**
 * Check if an error message contains a re-auth error code.
 */
function isReauthError(errorMessage: string): boolean {
  return ERROR_REQUIRES_REAUTH.some(
    (code) =>
      errorMessage.includes(code) || errorMessage.toLowerCase().includes(code.toLowerCase()),
  );
}

/**
 * Check if an error message indicates key rotation is needed.
 */
function isKeyRotationError(errorMessage: string): boolean {
  return ERROR_REQUIRES_KEY_ROTATION.some(
    (code) =>
      errorMessage.includes(code) || errorMessage.toLowerCase().includes(code.toLowerCase()),
  );
}

/**
 * Hook for using signed API client in React components.
 * Handles automatic re-authentication on signature failures.
 *
 * @example
 * ```tsx
 * function DeviceList() {
 *   const { get, loading, error } = useSignedApi({
 *     apiUrl: 'https://api.example.com',
 *     onReauthNeeded: () => navigate('/login'),
 *   });
 *
 *   useEffect(() => {
 *     get<Device[]>('/v1/device/count').then(console.log);
 *   }, []);
 *
 *   if (loading) return <Spinner />;
 *   if (error) return <Error error={error} />;
 *
 *   return <DeviceList />;
 * }
 * ```
 */
export function useSignedApi(options: UseSignedApiOptions): UseSignedApiReturn {
  const {
    apiUrl,
    autoFetchCredentials = true,
    clientName = "Web Dashboard",
    onReauthNeeded,
    onKeyRotationNeeded,
  } = options;

  const [client] = useState(() => getSignedApiClient(apiUrl));
  const [hasCredentials, setHasCredentials] = useState(() => client.hasCredentials());
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<Error | null>(null);
  const [reauthInProgress, setReauthInProgress] = useState(false);

  // Check for existing credentials on mount
  useEffect(() => {
    if (autoFetchCredentials && !hasCredentials && !reauthInProgress) {
      const existing = getActiveCredentials();
      if (existing) {
        client.setClient(existing);
        setHasCredentials(true);
        logger.info("useSignedApi", "Restored existing credentials");
      }
    }
  }, [autoFetchCredentials, hasCredentials, client, reauthInProgress]);

  /**
   * Handle signature/auth errors by attempting re-authentication.
   */
  const handleAuthError = useCallback(
    async (err: Error): Promise<boolean> => {
      const errorMessage = err.message || "";

      // Check if it's a key rotation error
      if (isKeyRotationError(errorMessage)) {
        logger.warn("useSignedApi", "Key rotation needed, clearing credentials");
        clearSignedApiClient();
        setHasCredentials(false);
        onKeyRotationNeeded?.("current");
        return false;
      }

      // Check if it's a re-auth error
      if (isReauthError(errorMessage)) {
        logger.warn("useSignedApi", "Signature auth failed, attempting re-auth");
        setReauthInProgress(true);
        try {
          // Clear old credentials
          clearSignedApiClient();

          // Try to fetch new credentials
          const apiUrlForFetch = apiUrl || window.location.origin;
          await fetchClientCredentials(apiUrlForFetch, clientName);

          // Update client with new credentials
          const newCreds = getActiveCredentials();
          if (newCreds) {
            client.setClient(newCreds);
            setHasCredentials(true);
            logger.info("useSignedApi", "Re-authentication successful");
            return true;
          }
        } catch (reauthErr) {
          logger.error("useSignedApi", `Re-auth failed: ${reauthErr}`);
          clearSignedApiClient();
          setHasCredentials(false);
          onReauthNeeded?.();
        } finally {
          setReauthInProgress(false);
        }
        return false;
      }

      return false;
    },
    [apiUrl, clientName, client, onReauthNeeded, onKeyRotationNeeded],
  );

  const fetchCredentials = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const apiUrlForFetch = apiUrl ?? window.location.origin;
      const creds = await client.fetchCredentials(apiUrlForFetch, clientName);
      setHasCredentials(true);
      logger.info("useSignedApi", `Fetched credentials for ${creds.clientId}`);
      return creds;
    } catch (e) {
      const err = e instanceof Error ? e : new Error(String(e));
      setError(err);
      logger.error("useSignedApi", `Failed to fetch credentials: ${err.message}`);
      throw err;
    } finally {
      setLoading(false);
    }
  }, [client, clientName, apiUrl]);

  const logout = useCallback(() => {
    clearSignedApiClient();
    setHasCredentials(false);
    setError(null);
    logger.info("useSignedApi", "Logged out");
  }, []);

  /**
   * Make a request with automatic re-authentication on failure.
   */
  const makeRequest = useCallback(
    async <T>(
      method: "GET" | "POST" | "PATCH" | "DELETE",
      path: string,
      body?: object,
    ): Promise<T> => {
      if (!hasCredentials && !reauthInProgress) {
        throw new Error("No credentials. Call fetchCredentials() first.");
      }

      try {
        setError(null);
        const result = await client.request<T>({ method, path, body });
        return result.data;
      } catch (e) {
        const err = e instanceof Error ? e : new Error(String(e));

        // Try to handle auth errors
        const reauthSucceeded = await handleAuthError(err);

        if (reauthSucceeded) {
          // Retry the request with new credentials
          const result = await client.request<T>({ method, path, body });
          return result.data;
        }

        // If re-auth failed or wasn't possible, throw the error
        setError(err);
        throw err;
      }
    },
    [client, hasCredentials, reauthInProgress, handleAuthError],
  );

  const get = useCallback(
    <T = unknown>(path: string) => makeRequest<T>("GET", path),
    [makeRequest],
  );

  const post = useCallback(
    <T = unknown>(path: string, body?: object) => makeRequest<T>("POST", path, body),
    [makeRequest],
  );

  const patch = useCallback(
    <T = unknown>(path: string, body?: object) => makeRequest<T>("PATCH", path, body),
    [makeRequest],
  );

  const delete_ = useCallback(
    <T = unknown>(path: string) => makeRequest<T>("DELETE", path),
    [makeRequest],
  );

  return {
    client,
    hasCredentials,
    loading,
    error,
    fetchCredentials,
    listCredentials: listClientCredentials,
    deleteCredentials: deleteClientCredentials,
    logout,
    get,
    post,
    patch,
    delete: delete_,
  };
}

/**
 * Provider component that makes the signed API client available globally.
 */
export const SignedApiContext = createContext<SignedApiClient | null>(null);

export function SignedApiProvider({
  children,
  apiUrl,
  clientName = "Web Dashboard",
}: {
  children: ReactNode;
  apiUrl: string;
  clientName?: string;
}) {
  const [client] = useState(() => getSignedApiClient(apiUrl));

  const _fetchCredentials = useCallback(async () => {
    const creds = await client.fetchCredentials(clientName);
    return creds;
  }, [client, clientName]);

  // Try to restore credentials on mount
  useEffect(() => {
    const existing = getActiveCredentials();
    if (existing) {
      client.setClient(existing);
    }
  }, [client]);

  return React.createElement(SignedApiContext.Provider, { value: client }, children);
}
