/**
 * REST Client Configuration
 * 
 * Shared REST client setup for all features.
 * Uses session-based authentication (cookies).
 */

import { DomainError, errorFromHTTP } from "@/domain/_shared";

// ============================================================================
// Configuration
// ============================================================================

/**
 * REST API configuration
 */
export interface RESTConfig {
  baseUrl: string;
  credentials?: RequestCredentials;
  defaultHeaders?: Record<string, string>;
}

/**
 * Get REST API config from environment
 */
export function getRESTConfig(): RESTConfig {
  const baseUrl = import.meta.env.VITE_API_URL ?? "";
  
  return {
    baseUrl,
    credentials: "include", // Send cookies for session auth
    defaultHeaders: {
      "Content-Type": "application/json",
    },
  };
}

// ============================================================================
// Fetch Options
// ============================================================================

/**
 * Extended fetch options for REST requests
 */
export interface FetchOptions extends RequestInit {
  params?: Record<string, string | number | boolean | undefined>;
  timeout?: number;
}

/**
 * Build URL with query parameters
 */
export function buildUrl(
  baseUrl: string,
  path: string,
  params?: Record<string, string | number | boolean | undefined>
): string {
  const url = new URL(path, baseUrl);
  
  if (params) {
    Object.entries(params).forEach(([key, value]) => {
      if (value !== undefined) {
        url.searchParams.set(key, String(value));
      }
    });
  }
  
  return url.toString();
}

// ============================================================================
// REST Client
// ============================================================================

/**
 * Base fetch wrapper with error handling
 */
export async function apiFetch<T>(
  path: string,
  options: FetchOptions = {}
): Promise<T> {
  const config = getRESTConfig();
  const { params, timeout, headers, ...fetchInit } = options;
  
  // Build URL with params
  const url = buildUrl(config.baseUrl, path, params);
  
  // Setup abort controller for timeout
  const controller = new AbortController();
  const timeoutId = timeout ? setTimeout(() => controller.abort(), timeout) : undefined;
  
  try {
    const response = await fetch(url, {
      method: "GET",
      headers: {
        ...config.defaultHeaders,
        ...headers,
      },
      credentials: config.credentials,
      signal: controller.signal,
      ...fetchInit,
    });
    
    // Handle no content
    if (response.status === 204) {
      return undefined as T;
    }
    
    // Parse response
    const contentType = response.headers.get("content-type");
    const data = contentType?.includes("application/json")
      ? await response.json()
      : await response.text();
    
    // Handle errors
    if (!response.ok) {
      throw errorFromHTTP(response.status, data);
    }
    
    return data as T;
  } finally {
    if (timeoutId) {
      clearTimeout(timeoutId);
    }
  }
}

// ============================================================================
// HTTP Methods
// ============================================================================

/**
 * GET request
 */
export async function apiGet<T>(
  path: string,
  params?: Record<string, string | number | boolean | undefined>,
  options?: Omit<FetchOptions, "params">
): Promise<T> {
  return apiFetch<T>(path, { ...options, params, method: "GET" });
}

/**
 * POST request
 */
export async function apiPost<T>(
  path: string,
  body?: unknown,
  options?: Omit<FetchOptions, "body">
): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: "POST",
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * PUT request
 */
export async function apiPut<T>(
  path: string,
  body?: unknown,
  options?: Omit<FetchOptions, "body">
): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: "PUT",
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * PATCH request
 */
export async function apiPatch<T>(
  path: string,
  body?: unknown,
  options?: Omit<FetchOptions, "body">
): Promise<T> {
  return apiFetch<T>(path, {
    ...options,
    method: "PATCH",
    body: body ? JSON.stringify(body) : undefined,
  });
}

/**
 * DELETE request
 */
export async function apiDelete<T>(
  path: string,
  options?: Omit<FetchOptions, "params">
): Promise<T> {
  return apiFetch<T>(path, { ...options, method: "DELETE" });
}

// ============================================================================
// API Response Types
// ============================================================================

/**
 * Standard API error response
 */
export interface APIErrorResponse {
  error?: string;
  message?: string;
  code?: string;
}

/**
 * Check if response is an error
 */
export function isAPIError(response: unknown): response is APIErrorResponse {
  return (
    typeof response === "object" &&
    response !== null &&
    ("error" in response || "message" in response)
  );
}