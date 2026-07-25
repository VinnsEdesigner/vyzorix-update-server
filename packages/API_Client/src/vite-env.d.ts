/// <reference types="vite/client" />

interface ImportMetaEnv {
  readonly VITE_API_URL?: string;
  readonly VITE_WS_URL?: string;
  readonly VITE_WS_RECONNECT_INTERVAL?: string;
  readonly VITE_WS_MAX_RECONNECT_ATTEMPTS?: string;
  readonly VITE_WS_HEARTBEAT_INTERVAL?: string;
  readonly VITE_WS_HEARTBEAT_TIMEOUT?: string;
  readonly VITE_WS_RECONNECT_MAX_DELAY?: string;
  readonly VITE_WS_RECONNECT_MULTIPLIER?: string;
  readonly VITE_REST_TIMEOUT?: string;
  readonly VITE_REST_WITH_CREDENTIALS?: string;
  readonly VITE_METRICS_DEFAULT_LIMIT?: string;
  readonly VITE_METRICS_RETENTION_DAYS?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
