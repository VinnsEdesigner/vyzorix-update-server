




import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';
import { getRESTConfig, type RESTConfig } from '../../config';
import { parseApiError, type RateLimitInfo } from '../../auth/api-error';

export { type RESTConfig } from '../../config';
export { parseApiError } from '../../auth/api-error';
export type { ApiError } from '../../auth/api-error';
export type { RateLimitInfo } from '../../auth/api-error';

let organizationId: string | null = null;
let authToken: string | null = null;
let csrfToken: string | null = null;
let refreshToken: string | null = null;

type RateLimitCallback = (info: RateLimitInfo) => void;
const rateLimitListeners = new Set<RateLimitCallback>();

let isRefreshing = false;
let refreshSubscribers: ((token: string | null) => void)[] = [];

export function setOrganizationContext(orgId: string | null): void {
  organizationId = orgId;
}

export function getOrganizationContext(): string | null {
  return organizationId;
}

export function setAuthToken(token: string | null): void {
  authToken = token;
}

export function getAuthToken(): string | null {
  return authToken;
}

export function setCSRFToken(token: string | null): void {
  csrfToken = token;
}

export function getCSRFToken(): string | null {
  return csrfToken;
}

export function setRefreshToken(token: string | null): void {
  refreshToken = token;
}

export function getRefreshToken(): string | null {
  return refreshToken;
}

export function clearAuthContext(): void {
  authToken = null;
  refreshToken = null;
  csrfToken = null;
}

export async function fetchAndSetCSRFToken(): Promise<string> {
  const response = await axios.get<{ csrf_token: string }>(
    `${getRESTConfig().baseURL}/v1/auth/csrf-token`,
    { withCredentials: true }
  );
  const token = response.data.csrf_token;
  setCSRFToken(token);
  return token;
}

export function onRateLimitChange(callback: RateLimitCallback): () => void {
  rateLimitListeners.add(callback);
  return () => rateLimitListeners.delete(callback);
}

function notifyRateLimit(info: RateLimitInfo): void {
  rateLimitListeners.forEach((cb) => cb(info));
}

async function handleTokenRefresh(): Promise<boolean> {
  if (!refreshToken) {
    return false;
  }

  if (isRefreshing) {
    return new Promise((resolve) => {
      refreshSubscribers.push((token) => {
        resolve(token !== null);
      });
    });
  }

  isRefreshing = true;

  try {
    const response = await axios.post(`${getRESTConfig().baseURL}/v1/auth/refresh`, {
      refresh_token: refreshToken,
    });

    const { access_token: newAccessToken, refresh_token: newRefreshToken } = response.data;
    if (newAccessToken) {
      setAuthToken(newAccessToken);
      if (newRefreshToken) {
        setRefreshToken(newRefreshToken);
      }
      refreshSubscribers.forEach((cb) => cb(newAccessToken));
      refreshSubscribers = [];
      return true;
    }
    return false;
  } catch {
    refreshSubscribers.forEach((cb) => cb(null));
    refreshSubscribers = [];
    return false;
  } finally {
    isRefreshing = false;
  }
}

function createAxiosInstance(config: RESTConfig): AxiosInstance {
  const instance = axios.create({
    baseURL: config.baseURL,
    timeout: config.timeout,
    withCredentials: config.withCredentials,
    headers: {
      'Content-Type': 'application/json',
    },
  });

  instance.interceptors.request.use(
    async (config: InternalAxiosRequestConfig) => {
      if (organizationId) {
        config.headers.set('X-Organization-ID', organizationId);
      }

      if (authToken) {
        config.headers.set('Authorization', `Bearer ${authToken}`);
      }

      if (csrfToken) {
        config.headers.set('X-CSRF-Token', csrfToken);
      }

      console.debug(`[REST] ${config.method?.toUpperCase()} ${config.url}`);
      return config;
    },
    (error) => {
      console.error('[REST] Request error:', error);
      return Promise.reject(error);
    }
  );

  instance.interceptors.response.use(
    (response: AxiosResponse) => {
      const rateLimit = response.headers['x-rate-limit'];
      if (rateLimit) {
        const rateLimitInfo: RateLimitInfo = {
          limit: parseInt(response.headers['x-ratelimit-limit'] as string) || 0,
          remaining: parseInt(response.headers['x-ratelimit-remaining'] as string) || 0,
          resetAt: parseInt(response.headers['x-ratelimit-reset'] as string) || 0,
          retryAfter: parseInt(response.headers['retry-after'] as string) || 0,
        };
        notifyRateLimit(rateLimitInfo);
      }
      return response;
    },
    async (error) => {
      if (axios.isAxiosError(error)) {
        const status = error.response?.status ?? 0;

        if (status === 401 && !error.config?.headers?.['X-Retry']) {
          const refreshed = await handleTokenRefresh();
          if (refreshed && error.config) {
            error.config.headers['X-Retry'] = 'true';
            const response = await getAxios().request(error.config);
            return response;
          }
        }

        const apiError = parseApiError(error);
        console.error(`[REST] Error ${status}:`, apiError.message);
        throw apiError;
      }
      console.error('[REST] Unexpected error:', error);
      throw parseApiError(error);
    }
  );

  return instance;
}

let axiosInstance: AxiosInstance | null = null;

function getAxios(): AxiosInstance {
  if (!axiosInstance) {
    axiosInstance = createAxiosInstance(getRESTConfig());
  }
  return axiosInstance;
}

export const restClient = {
  async get<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await getAxios().get<T>(url, config);
    return response.data;
  },

  async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await getAxios().post<T>(url, data, config);
    return response.data;
  },

  async put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await getAxios().put<T>(url, data, config);
    return response.data;
  },

  async patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    const response = await getAxios().patch<T>(url, data, config);
    return response.data;
  },

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    const response = await getAxios().delete<T>(url, config);
    return response.data;
  },
};

export { axios };
