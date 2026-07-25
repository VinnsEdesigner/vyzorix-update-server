





export interface WebSocketConfig {
  url: string;
  reconnectInterval: number;
  maxReconnectAttempts: number;
  heartbeatInterval: number;
  heartbeatTimeout: number;
  reconnectMaxDelay: number;
  reconnectMultiplier: number;
}

export interface RESTConfig {
  baseURL: string;
  timeout: number;
  withCredentials: boolean;
}

export interface MetricsConfig {
  defaultLimit: number;
  retentionDays: number;
}

export interface ClientConfig {
  ws: WebSocketConfig;
  rest: RESTConfig;
  metrics: MetricsConfig;
}





function parseIntEnv(key: string, defaultValue: number): number {
  const value = import.meta.env[key];
  if (value === undefined) return defaultValue;
  const parsed = parseInt(value, 10);
  return isNaN(parsed) ? defaultValue : parsed;
}

function parseStringEnv(key: string, defaultValue: string): string {
  return import.meta.env[key] ?? defaultValue;
}

function parseBoolEnv(key: string, defaultValue: boolean): boolean {
  const value = import.meta.env[key];
  if (value === undefined) return defaultValue;
  return value === "true" || value === "1";
}





export function getClientConfig(): ClientConfig {
  return {
    ws: getWebSocketConfig(),
    rest: getRESTConfig(),
    metrics: getMetricsConfig(),
  };
}

export function getWebSocketConfig(): WebSocketConfig {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const host = window.location.host;
  // Default WebSocket URL points to the device streaming endpoint
  // For dashboard clients: wss://host/v1/device/{clientId}/stream
  const wsUrl = parseStringEnv("VITE_WS_URL", `${protocol}//${host}/v1/device`);
  
  return {
    url: wsUrl,
    reconnectInterval: parseIntEnv("VITE_WS_RECONNECT_INTERVAL", 3000),
    maxReconnectAttempts: parseIntEnv("VITE_WS_MAX_RECONNECT_ATTEMPTS", 5),
    heartbeatInterval: parseIntEnv("VITE_WS_HEARTBEAT_INTERVAL", 30000),
    heartbeatTimeout: parseIntEnv("VITE_WS_HEARTBEAT_TIMEOUT", 10000),
    reconnectMaxDelay: parseIntEnv("VITE_WS_RECONNECT_MAX_DELAY", 30000),
    reconnectMultiplier: parseIntEnv("VITE_WS_RECONNECT_MULTIPLIER", 2),
  };
}

/**
 * Build the WebSocket URL for device streaming.
 * @param baseUrl - The base WebSocket URL (e.g., from VITE_WS_URL or getWebSocketConfig().url)
 * @param deviceId - The device ID (IMEI) or client ID for dashboard connections
 * @returns The full WebSocket URL with the device path
 */
export function buildDeviceStreamUrl(baseUrl: string, deviceId: string): string {
  const protocol = baseUrl.startsWith("wss") ? "wss:" : "ws:";
  const separator = baseUrl.includes('?') ? '&' : '?';
  // Remove protocol prefix if present to reconstruct properly
  const urlWithoutProtocol = baseUrl.replace(/^(wss?|https?):\/\//, '');
  return `${protocol}//${urlWithoutProtocol}/${encodeURIComponent(deviceId)}/stream`;
}

export function getRESTConfig(): RESTConfig {
  return {
    baseURL: parseStringEnv("VITE_API_URL", "/api"),
    timeout: parseIntEnv("VITE_REST_TIMEOUT", 30000),
    withCredentials: parseBoolEnv("VITE_REST_WITH_CREDENTIALS", true),
  };
}

export function getMetricsConfig(): MetricsConfig {
  return {
    defaultLimit: parseIntEnv("VITE_METRICS_DEFAULT_LIMIT", 500),
    retentionDays: parseIntEnv("VITE_METRICS_RETENTION_DAYS", 30),
  };
}





export const config = {
  get ws(): WebSocketConfig {
    return getWebSocketConfig();
  },
  get rest(): RESTConfig {
    return getRESTConfig();
  },
  get metrics(): MetricsConfig {
    return getMetricsConfig();
  },
};
