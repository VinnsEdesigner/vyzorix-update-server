// @vyzorix/config/api - API Client Configuration
import { z } from "zod";

/**
 * API Client Configuration
 */
export interface ApiClientConfig {
  /**
   * Base URL for API requests
   */
  baseUrl: string;
  
  /**
   * Authentication configuration
   */
  auth?: {
    /**
     * Token storage method
     */
    tokenStorage?: "localStorage" | "sessionStorage" | "memory";
    
    /**
     * Token storage key
     */
    tokenKey?: string;
    
    /**
     * Auto-refresh token endpoint
     */
    refreshEndpoint?: string;
    
    /**
     * Refresh token before expiry (in seconds)
     */
    refreshBeforeExpiry?: number;
  };
  
  /**
   * Request timeout in milliseconds
   */
  timeout?: number;
  
  /**
   * Default headers
   */
  headers?: Record<string, string>;
  
  /**
   * Error handling
   */
  onError?: (error: Error, request: Request, response?: Response) => void;
}

/**
 * API Response structure
 */
export interface ApiResponse<T = any> {
  data: T;
  status: number;
  headers: Headers;
  ok: boolean;
}

/**
 * API Error
 */
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public response?: Response,
    public request?: Request
  ) {
    super(message);
    this.name = "ApiError";
  }
}

/**
 * Create API Client
 */
export function createApiClient(config: ApiClientConfig) {
  const { 
    baseUrl, 
    auth: authConfig = {}, 
    timeout = 10000, 
    headers: defaultHeaders = {},
    onError
  } = config;

  // Validate configuration
  const configSchema = z.object({
    baseUrl: z.string().url(),
    auth: z.object({
      tokenStorage: z.enum(["localStorage", "sessionStorage", "memory"]).default("localStorage"),
      tokenKey: z.string().default("vyzorix_token"),
      refreshEndpoint: z.string().url().optional(),
      refreshBeforeExpiry: z.number().positive().default(300), // 5 minutes
    }).default({}),
    timeout: z.number().positive().default(10000),
    headers: z.record(z.string()).default({}),
    onError: z.function().optional(),
  });

  const validatedConfig = configSchema.parse({
    baseUrl,
    auth: authConfig,
    timeout,
    headers: defaultHeaders,
    onError,
  });

  // Token storage adapter
  const tokenStorage = {
    get: () => {
      try {
        if (validatedConfig.auth.tokenStorage === "localStorage") {
          return localStorage.getItem(validatedConfig.auth.tokenKey);
        } else if (validatedConfig.auth.tokenStorage === "sessionStorage") {
          return sessionStorage.getItem(validatedConfig.auth.tokenKey);
        }
        return null;
      } catch {
        return null;
      }
    },
    set: (token: string) => {
      try {
        if (validatedConfig.auth.tokenStorage === "localStorage") {
          localStorage.setItem(validatedConfig.auth.tokenKey, token);
        } else if (validatedConfig.auth.tokenStorage === "sessionStorage") {
          sessionStorage.setItem(validatedConfig.auth.tokenKey, token);
        }
      } catch (e) {
        console.error("Failed to store token:", e);
      }
    },
    remove: () => {
      try {
        if (validatedConfig.auth.tokenStorage === "localStorage") {
          localStorage.removeItem(validatedConfig.auth.tokenKey);
        } else if (validatedConfig.auth.tokenStorage === "sessionStorage") {
          sessionStorage.removeItem(validatedConfig.auth.tokenKey);
        }
      } catch (e) {
        console.error("Failed to remove token:", e);
      }
    },
  };

  // Request helper
  const request = async <T = any>(
    method: string,
    path: string,
    data?: any,
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> => {
    const url = new URL(path, validatedConfig.baseUrl).toString();
    const token = tokenStorage.get();

    // Build headers
    const headers = new Headers({
      ...validatedConfig.headers,
      ...options.headers,
      "Content-Type": "application/json",
    });

    if (token) {
      headers.set("Authorization", `Bearer ${token}`);
    }

    // Build request options
    const requestOptions: RequestInit = {
      method,
      headers,
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    };

    // Set timeout
    const controller = new AbortController();
    const timeoutId = setTimeout(() => controller.abort(), validatedConfig.timeout);
    requestOptions.signal = controller.signal;

    try {
      const response = await fetch(url, requestOptions);
      clearTimeout(timeoutId);

      if (!response.ok) {
        let errorMessage = `HTTP error! status: ${response.status}`;
        try {
          const errorData = await response.json();
          errorMessage = errorData.message || errorMessage;
        } catch (_e) {
          // Couldn't parse error response, use default message
        }

        const error = new ApiError(errorMessage, response.status, response, new Request(url, requestOptions));
        if (onError) {
          onError(error, new Request(url, requestOptions), response);
        }
        throw error;
      }

      let responseData: T;
      try {
        responseData = await response.json();
      } catch (_e) {
        responseData = {} as T; // Empty object for non-JSON responses
      }

      return {
        data: responseData,
        status: response.status,
        headers: response.headers,
        ok: response.ok,
      };
    } catch (error) {
      clearTimeout(timeoutId);
      if (error instanceof ApiError) {
        throw error;
      }
      
      const errorMessage = error instanceof Error ? error.message : String(error);
      const apiError = new ApiError(errorMessage, 0, undefined, new Request(url, requestOptions));
      if (onError) {
        onError(apiError, new Request(url, requestOptions));
      }
      throw apiError;
    }
  };

  // Convenience methods
  const client = {
    get: <T = any>(path: string, options?: RequestInit) => 
      request<T>("GET", path, undefined, options),
    post: <T = any>(path: string, data?: any, options?: RequestInit) => 
      request<T>("POST", path, data, options),
    put: <T = any>(path: string, data?: any, options?: RequestInit) => 
      request<T>("PUT", path, data, options),
    patch: <T = any>(path: string, data?: any, options?: RequestInit) => 
      request<T>("PATCH", path, data, options),
    delete: <T = any>(path: string, options?: RequestInit) => 
      request<T>("DELETE", path, undefined, options),
    head: <T = any>(path: string, options?: RequestInit) => 
      request<T>("HEAD", path, undefined, options),
    
    // Token management
    getToken: () => tokenStorage.get(),
    setToken: (token: string) => tokenStorage.set(token),
    clearToken: () => tokenStorage.remove(),
    
    // Configuration
    config: validatedConfig,
  };

  return client;
}