import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';
import { getRESTConfig, type RESTConfig } from '../../config';
import { parseApiError, type RateLimitInfo } from '../../auth/api-error';
import { getConnectivityMonitor } from '../_connectivity';
import { getRESTBatcher } from '../_batching';

export { type RESTConfig } from '../../config';
export { parseApiError } from '../../auth/api-error';
export type { ApiError } from '../../auth/api-error';
export type { RateLimitInfo } from '../../auth/api-error';

export { initConnectivityMonitor, getConnectivityMonitor } from '../_connectivity';
export { getRESTBatcher, resetBatchers } from '../_batching';

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

const CIRCUIT_BREAKER_CONFIG = {
  failureThreshold: 5,
  successThreshold: 2,
  windowMs: 30000,
  timeout: 60000,
  halfOpenMaxAttempts: 3,
};

enum CircuitState {
  CLOSED = 'CLOSED',
  OPEN = 'OPEN',
  HALF_OPEN = 'HALF_OPEN',
}

interface FailureEvent {
  timestamp: number;
}

interface CircuitBreaker {
  state: CircuitState;
  failures: FailureEvent[];
  successes: number;
  halfOpenAttempts: number;
  lastFailureTime: number;
  nextAttempt: number;
  cachedException: Error | null;
}

const circuitBreaker: CircuitBreaker = {
  state: CircuitState.CLOSED,
  failures: [],
  successes: 0,
  halfOpenAttempts: 0,
  lastFailureTime: 0,
  nextAttempt: 0,
  cachedException: null,
};

const RETRY_CONFIG = {
  maxRetries: 3,
  baseDelay: 1000,
  maxDelay: 10000,
  retryableStatuses: [502, 503, 504],
};

function getRecentFailures(): FailureEvent[] {
  const now = Date.now();
  const cutoff = now - CIRCUIT_BREAKER_CONFIG.windowMs;
  circuitBreaker.failures = circuitBreaker.failures.filter((f) => f.timestamp > cutoff);
  return circuitBreaker.failures;
}

function shouldRetry(error: unknown): boolean {
  if (axios.isAxiosError(error)) {
    const status = error.response?.status ?? 0;
    if (RETRY_CONFIG.retryableStatuses.includes(status)) {
      return true;
    }
    if (error.code === 'ECONNRESET' || error.code === 'ETIMEDOUT' || error.code === 'NETWORK_ERROR') {
      return true;
    }
  }
  if (!axios.isAxiosError(error) && error instanceof Error) {
    if (error.message.includes('Network Error') || error.message.includes('timeout')) {
      return true;
    }
  }
  return false;
}

function calculateBackoff(retryCount: number): number {
  const exponentialDelay = RETRY_CONFIG.baseDelay * Math.pow(2, retryCount);
  const jitter = (Math.random() * 0.4 - 0.2) * exponentialDelay;
  return Math.min(Math.max(exponentialDelay + jitter, RETRY_CONFIG.baseDelay), RETRY_CONFIG.maxDelay);
}

function generateUUIDv4(): string {
  return 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, (c) => {
    const r = (Math.random() * 16) | 0;
    const v = c === 'x' ? r : (r & 0x3) | 0x8;
    return v.toString(16);
  });
}

function recordSuccess(): void {
  circuitBreaker.lastFailureTime = 0;
  circuitBreaker.cachedException = null;

  if (circuitBreaker.state === CircuitState.HALF_OPEN) {
    circuitBreaker.successes++;

    if (circuitBreaker.successes >= CIRCUIT_BREAKER_CONFIG.successThreshold) {
      console.log('[CircuitBreaker] Recovery successful, closing circuit');
      circuitBreaker.state = CircuitState.CLOSED;
      circuitBreaker.failures = [];
      circuitBreaker.successes = 0;
      circuitBreaker.halfOpenAttempts = 0;
      circuitBreaker.cachedException = null;
    }
  } else if (circuitBreaker.state === CircuitState.CLOSED) {
    circuitBreaker.failures = [];
  }
}

function recordFailure(error: Error): void {
  circuitBreaker.lastFailureTime = Date.now();
  circuitBreaker.failures.push({ timestamp: Date.now() });
  circuitBreaker.cachedException = error;

  const recentFailures = getRecentFailures();

  if (circuitBreaker.state === CircuitState.HALF_OPEN) {
    circuitBreaker.halfOpenAttempts++;

    if (circuitBreaker.halfOpenAttempts >= CIRCUIT_BREAKER_CONFIG.halfOpenMaxAttempts) {
      console.log('[CircuitBreaker] Half-open test failed, reopening circuit');
      circuitBreaker.state = CircuitState.OPEN;
      circuitBreaker.nextAttempt = Date.now() + CIRCUIT_BREAKER_CONFIG.timeout;
      circuitBreaker.halfOpenAttempts = 0;
      circuitBreaker.successes = 0;
    }
  } else if (circuitBreaker.state === CircuitState.CLOSED) {
    if (recentFailures.length >= CIRCUIT_BREAKER_CONFIG.failureThreshold) {
      console.log(`[CircuitBreaker] ${recentFailures.length} failures in ${CIRCUIT_BREAKER_CONFIG.windowMs}ms window, opening circuit`);
      circuitBreaker.state = CircuitState.OPEN;
      circuitBreaker.nextAttempt = Date.now() + CIRCUIT_BREAKER_CONFIG.timeout;
    }
  }
}

function canAttempt(): boolean {
  if (circuitBreaker.state === CircuitState.CLOSED) {
    return true;
  }

  if (circuitBreaker.state === CircuitState.HALF_OPEN) {
    return circuitBreaker.halfOpenAttempts < CIRCUIT_BREAKER_CONFIG.halfOpenMaxAttempts;
  }

  if (circuitBreaker.state === CircuitState.OPEN) {
    if (Date.now() >= circuitBreaker.nextAttempt) {
      console.log('[CircuitBreaker] Timeout elapsed, entering half-open state');
      circuitBreaker.state = CircuitState.HALF_OPEN;
      circuitBreaker.halfOpenAttempts = 0;
      circuitBreaker.successes = 0;
      return true;
    }
    return false;
  }

  return true;
}

function getCircuitBreakerError(): Error {
  if (circuitBreaker.cachedException) {
    const err = new Error(`Circuit OPEN: ${circuitBreaker.cachedException.message}`) as Error & { code: string; originalError?: Error };
    err.code = 'CIRCUIT_OPEN';
    err.originalError = circuitBreaker.cachedException;
    return err;
  }
  const err = new Error('Circuit breaker is OPEN - request blocked') as Error & { code: string };
  err.code = 'CIRCUIT_OPEN';
  return err;
}

function hashPayload(data: unknown): string {
  const str = data ? JSON.stringify(data) : '';
  let hash = 0;
  for (let i = 0; i < str.length; i++) {
    const char = str.charCodeAt(i);
    hash = (hash << 5) - hash + char;
    hash = hash & hash;
  }
  return Math.abs(hash).toString(36);
}

interface InFlightRequest {
  promise: Promise<unknown>;
  idempotencyKey: string;
}

const inFlightRequests = new Map<string, InFlightRequest>();

function getInFlightKey(method: string, url: string, data: unknown): string {
  return `${method}:${url}:${hashPayload(data)}`;
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
      if (!canAttempt()) {
        throw getCircuitBreakerError();
      }

      if (organizationId) {
        config.headers.set('X-Organization-ID', organizationId);
      }

      if (authToken) {
        config.headers.set('Authorization', `Bearer ${authToken}`);
      }

      if (csrfToken) {
        config.headers.set('X-CSRF-Token', csrfToken);
      }

      // Generate idempotency key for mutations (reuse existing if retry)
      const method = config.method?.toUpperCase();
      if (method && ['POST', 'PUT', 'PATCH', 'DELETE'].includes(method)) {
        const retryConfig = config as InternalAxiosRequestConfig & { __idempotencyKey?: string };
        if (!retryConfig.__idempotencyKey) {
          retryConfig.__idempotencyKey = generateUUIDv4();
        }
        config.headers.set('X-Idempotency-Key', retryConfig.__idempotencyKey);
      }

      console.debug(`[REST] ${method} ${config.url}`);
      return config;
    },
    (error) => {
      console.error('[REST] Request error:', error);
      return Promise.reject(error);
    }
  );

  instance.interceptors.response.use(
    (response: AxiosResponse) => {
      recordSuccess();

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
      if (!axios.isAxiosError(error) || !error.config) {
        const parsedError = parseApiError(error);
        recordFailure(parsedError);
        console.error('[REST] Unexpected error:', error);
        throw parsedError;
      }

      const status = error.response?.status ?? 0;
      const retryCount = (error.config as { __retryCount?: number }).__retryCount ?? 0;

      if (status === 401 && !error.config.headers?.['X-Retry']) {
        const refreshed = await handleTokenRefresh();
        if (refreshed) {
          error.config.headers['X-Retry'] = 'true';
          const response = await instance.request(error.config);
          return response;
        }
      }

      if (shouldRetry(error) && retryCount < RETRY_CONFIG.maxRetries) {
        const delay = calculateBackoff(retryCount);
        console.warn(`[REST] Retry ${retryCount + 1}/${RETRY_CONFIG.maxRetries} after ${Math.round(delay)}ms (status ${status})`);

        (error.config as { __retryCount?: number }).__retryCount = retryCount + 1;

        await new Promise((resolve) => setTimeout(resolve, delay));

        try {
          const response = await instance.request(error.config);
          recordSuccess();
          return response;
        } catch (retryError) {
          const parsed = parseApiError(retryError);
          recordFailure(parsed);
          throw retryError;
        }
      }

      const apiError = parseApiError(error);
      recordFailure(apiError);

      if (!canAttempt()) {
        throw getCircuitBreakerError();
      }

      console.error(`[REST] Error ${status}:`, apiError.message);
      throw apiError;
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
    if (!canAttempt()) {
      throw getCircuitBreakerError();
    }
    const monitor = getConnectivityMonitor();
    if (!monitor.isOnline()) {
      console.warn(`[REST] Offline GET: ${url}`);
    }

    // Use request batching for GET requests (collapses duplicate GETs within 50ms window)
    const batcher = getRESTBatcher();
    const params = config?.params;
    return batcher.execute(
      'GET',
      url,
      params,
      config as Record<string, unknown>,
      async () => {
        const response = await getAxios().get<T>(url, config);
        return response.data;
      }
    );
  },

  async post<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    if (!canAttempt()) {
      throw getCircuitBreakerError();
    }

    // In-flight deduplication
    const inFlightKey = getInFlightKey('POST', url, data);
    const existing = inFlightRequests.get(inFlightKey);
    if (existing) {
      console.warn(`[REST] Duplicate POST detected for ${url}, returning existing promise`);
      return existing.promise as Promise<T>;
    }

    const monitor = getConnectivityMonitor();
    if (!monitor.isOnline()) {
      console.warn(`[REST] Offline POST: ${url} - queueing`);
      return monitor.queueRequest({
        method: 'POST',
        url,
        data,
        params: config?.params as Record<string, unknown>,
        headers: config?.headers as Record<string, string>,
      }) as Promise<T>;
    }

    const requestPromise = getAxios().post<T>(url, data, config);
    const idempotencyKey = (config as InternalAxiosRequestConfig & { __idempotencyKey?: string }).__idempotencyKey || '';
    inFlightRequests.set(inFlightKey, { promise: requestPromise, idempotencyKey });

    try {
      const result = await requestPromise;
      inFlightRequests.delete(inFlightKey);
      return result.data as T;
    } catch (error) {
      inFlightRequests.delete(inFlightKey);
      throw error;
    }
  },

  async put<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    if (!canAttempt()) {
      throw getCircuitBreakerError();
    }

    const inFlightKey = getInFlightKey('PUT', url, data);
    const existing = inFlightRequests.get(inFlightKey);
    if (existing) {
      console.warn(`[REST] Duplicate PUT detected for ${url}, returning existing promise`);
      return existing.promise as Promise<T>;
    }

    const monitor = getConnectivityMonitor();
    if (!monitor.isOnline()) {
      console.warn(`[REST] Offline PUT: ${url} - queueing`);
      return monitor.queueRequest({
        method: 'PUT',
        url,
        data,
        params: config?.params as Record<string, unknown>,
        headers: config?.headers as Record<string, string>,
      }) as Promise<T>;
    }

    const requestPromise = getAxios().put<T>(url, data, config);
    const idempotencyKey = (config as InternalAxiosRequestConfig & { __idempotencyKey?: string }).__idempotencyKey || '';
    inFlightRequests.set(inFlightKey, { promise: requestPromise, idempotencyKey });

    try {
      const result = await requestPromise;
      inFlightRequests.delete(inFlightKey);
      return result.data as T;
    } catch (error) {
      inFlightRequests.delete(inFlightKey);
      throw error;
    }
  },

  async patch<T>(url: string, data?: unknown, config?: AxiosRequestConfig): Promise<T> {
    if (!canAttempt()) {
      throw getCircuitBreakerError();
    }

    const inFlightKey = getInFlightKey('PATCH', url, data);
    const existing = inFlightRequests.get(inFlightKey);
    if (existing) {
      console.warn(`[REST] Duplicate PATCH detected for ${url}, returning existing promise`);
      return existing.promise as Promise<T>;
    }

    const monitor = getConnectivityMonitor();
    if (!monitor.isOnline()) {
      console.warn(`[REST] Offline PATCH: ${url} - queueing`);
      return monitor.queueRequest({
        method: 'PATCH',
        url,
        data,
        params: config?.params as Record<string, unknown>,
        headers: config?.headers as Record<string, string>,
      }) as Promise<T>;
    }

    const requestPromise = getAxios().patch<T>(url, data, config);
    const idempotencyKey = (config as InternalAxiosRequestConfig & { __idempotencyKey?: string }).__idempotencyKey || '';
    inFlightRequests.set(inFlightKey, { promise: requestPromise, idempotencyKey });

    try {
      const result = await requestPromise;
      inFlightRequests.delete(inFlightKey);
      return result.data as T;
    } catch (error) {
      inFlightRequests.delete(inFlightKey);
      throw error;
    }
  },

  async delete<T>(url: string, config?: AxiosRequestConfig): Promise<T> {
    if (!canAttempt()) {
      throw getCircuitBreakerError();
    }

    const inFlightKey = getInFlightKey('DELETE', url, config?.data);
    const existing = inFlightRequests.get(inFlightKey);
    if (existing) {
      console.warn(`[REST] Duplicate DELETE detected for ${url}, returning existing promise`);
      return existing.promise as Promise<T>;
    }

    const monitor = getConnectivityMonitor();
    if (!monitor.isOnline()) {
      console.warn(`[REST] Offline DELETE: ${url} - queueing`);
      return monitor.queueRequest({
        method: 'DELETE',
        url,
        params: config?.params as Record<string, unknown>,
        headers: config?.headers as Record<string, string>,
      }) as Promise<T>;
    }

    const requestPromise = getAxios().delete<T>(url, config);
    const idempotencyKey = (config as InternalAxiosRequestConfig & { __idempotencyKey?: string }).__idempotencyKey || '';
    inFlightRequests.set(inFlightKey, { promise: requestPromise, idempotencyKey });

    try {
      const result = await requestPromise;
      inFlightRequests.delete(inFlightKey);
      return result.data as T;
    } catch (error) {
      inFlightRequests.delete(inFlightKey);
      throw error;
    }
  },
};

export { axios };

export function getCircuitBreakerState(): { state: CircuitState; recentFailures: number; cachedException?: string } {
  return {
    state: circuitBreaker.state,
    recentFailures: getRecentFailures().length,
    cachedException: circuitBreaker.cachedException?.message,
  };
}

export function resetCircuitBreaker(): void {
  circuitBreaker.state = CircuitState.CLOSED;
  circuitBreaker.failures = [];
  circuitBreaker.successes = 0;
  circuitBreaker.halfOpenAttempts = 0;
  circuitBreaker.lastFailureTime = 0;
  circuitBreaker.nextAttempt = 0;
  circuitBreaker.cachedException = null;
  console.log('[CircuitBreaker] Manually reset');
}

export function getInFlightRequestCount(): number {
  return inFlightRequests.size;
}

export function clearInFlightRequests(): void {
  inFlightRequests.clear();
  console.log('[REST] Cleared all in-flight requests');
}
