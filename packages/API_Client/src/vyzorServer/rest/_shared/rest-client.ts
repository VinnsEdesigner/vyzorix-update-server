// REST Client using Axios
// Handles all REST API communication with proper error handling

import axios, { type AxiosInstance, type AxiosRequestConfig, type AxiosResponse, type InternalAxiosRequestConfig } from 'axios';

// ============================================================================
// Configuration
// ============================================================================

export interface RESTConfig {
  baseURL: string;
  timeout?: number;
  withCredentials?: boolean;
}

function getRESTConfig(): RESTConfig {
  return {
    baseURL: import.meta.env.VITE_API_URL ?? '/api',
    timeout: 30000,
    withCredentials: true,
  };
}

// ============================================================================
// Axios Instance
// ============================================================================

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
    (config: InternalAxiosRequestConfig) => {
      console.debug(`[REST] ${config.method?.toUpperCase()} ${config.url}`);
      return config;
    },
    (error) => {
      console.error('[REST] Request error:', error);
      return Promise.reject(error);
    }
  );

  instance.interceptors.response.use(
    (response: AxiosResponse) => response,
    (error) => {
      if (axios.isAxiosError(error)) {
        const status = error.response?.status ?? 0;
        const data = error.response?.data;
        console.error(`[REST] Error ${status}:`, data ?? error.message);
        throw error;
      }
      console.error('[REST] Unexpected error:', error);
      throw error;
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

// ============================================================================
// REST Client
// ============================================================================

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
