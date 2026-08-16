import { getRESTConfig } from "../config";
import { signHttpRequestBrowser, type SignHeaders } from "../crypto/browser-sign";

/**
 * Developer API client for third-party applications.
 *
 * Uses the `X-API-Key` header for authentication — the server's
 * TenantAPIKeyAuth middleware (apps/api/internal/api/middleware/tenant_api_key.go)
 * validates it on all tenant endpoints (/v1/devices, /v1/command, /v1/telemetry,
 * /v1/updates, …) and enforces per-key rate limiting (429 + Retry-After).
 *
 * The server derives a deterministic signing secret from the full API key
 * (hex(SHA-512(fullKey)) — see api_key_service.go:deriveAPIKeySigningSecret) and
 * uses it to verify HMAC request signatures via SessionSignatureMiddleware.
 * This client derives the same secret and signs every state-changing and
 * tenant-path request with the X-Vyzorix-* headers.
 *
 * This is deliberately separate from the session-based restClient: no cookies,
 * no CSRF token, no refresh-token flow. The developer client is stateless.
 */
export interface DeveloperClientOptions {
  baseUrl?: string;
  /** Organization ID sent as X-Organization-ID on every request (required for org-scoped endpoints). */
  organizationId?: string;
  onUnauthorized?: () => void;
  onRateLimited?: (retryAfter?: number) => void;
}

export interface DeveloperClient {
  getDevices: () => Promise<unknown>;
  getDevice: (imei: string) => Promise<unknown>;
  deleteDevice: (imei: string) => Promise<void>;

  getCommandStatus: (dispatchId: string) => Promise<unknown>;
  retryCommand: (dispatchId: string) => Promise<unknown>;
  cancelCommand: (dispatchId: string) => Promise<void>;

  getTelemetryHistory: (params: Record<string, string | number>) => Promise<unknown>;
  getLatestTelemetry: (deviceId: string) => Promise<unknown>;

  getUpdateStatus: () => Promise<unknown>;
  getVersions: () => Promise<unknown>;

  request: <T>(endpoint: string, init?: RequestInit) => Promise<T>;
}

function resolveBaseUrl(override?: string): string {
  if (override) return override;
  try {
    return getRESTConfig().baseURL;
  } catch {
    return "";
  }
}

/**
 * Derive the HMAC signing secret from a full API key.
 * Mirrors Go's deriveAPIKeySigningSecret (api_key_service.go:412):
 *   signingSecret = hex(SHA-512(fullKey))
 * The server stores this derived secret at key creation time and uses it to
 * verify X-Vyzorix-* signatures on API-key-authenticated requests.
 *
 * Uses the Web Crypto API (crypto.subtle) to stay browser-bundle safe.
 * Results are memoized per key string.
 */
const signingSecretCache = new Map<string, string>();

export async function deriveAPIKeySigningSecret(apiKey: string): Promise<string> {
  const cached = signingSecretCache.get(apiKey);
  if (cached) return cached;
  const data = new TextEncoder().encode(apiKey);
  const buf = await crypto.subtle.digest("SHA-512", data);
  const bytes = new Uint8Array(buf);
  let hex = "";
  for (let i = 0; i < bytes.length; i++) {
    hex += (bytes[i] ?? 0).toString(16).padStart(2, "0");
  }
  signingSecretCache.set(apiKey, hex);
  return hex;
}

/**
 * Determine whether an endpoint requires HMAC signing.
 * Tenant paths (devices, command, telemetry, updates, dashboard, connections)
 * require X-Vyzorix-* headers. Public and auth endpoints do not.
 */
function requiresSigning(endpoint: string): boolean {
  const tenantPrefixes = [
    "/v1/dashboard",
    "/v1/devices",
    "/v1/command",
    "/v1/telemetry",
    "/v1/updates",
    "/v1/connections",
    "/v1/device/diagnostics",
  ];
  return tenantPrefixes.some((p) => endpoint.startsWith(p));
}

export function createDeveloperClient(
  apiKey: string,
  options: DeveloperClientOptions = {},
): DeveloperClient {
  const baseUrl = resolveBaseUrl(options.baseUrl);

  const request = async <T>(endpoint: string, init: RequestInit = {}): Promise<T> => {
    const url = endpoint.startsWith("http") ? endpoint : `${baseUrl}${endpoint}`;
    const method = (init.method || "GET").toUpperCase();
    const bodyStr = init.body ? String(init.body) : "";

    const headers: Record<string, string> = {
      "X-API-Key": apiKey,
      "Content-Type": "application/json",
      ...((init.headers as Record<string, string>) || {}),
    };

    // Inject organization context (required by OrganizationContext middleware
    // for all org-scoped tenant endpoints).
    if (options.organizationId && !headers["X-Organization-ID"]) {
      headers["X-Organization-ID"] = options.organizationId;
    }

    // Sign tenant-path requests with the API-key-derived HMAC secret.
    if (requiresSigning(endpoint)) {
      const signingSecret = await deriveAPIKeySigningSecret(apiKey);
      const signHeaders: SignHeaders = await signHttpRequestBrowser(
        method,
        endpoint,
        bodyStr,
        signingSecret,
      );
      headers["X-Vyzorix-Timestamp"] = signHeaders["X-Vyzorix-Timestamp"];
      headers["X-Vyzorix-Nonce"] = signHeaders["X-Vyzorix-Nonce"];
      headers["X-Vyzorix-Signature"] = signHeaders["X-Vyzorix-Signature"];
    }

    const response = await fetch(url, {
      ...init,
      headers,
    });

    if (response.status === 401) {
      options.onUnauthorized?.();
      throw new Error("Invalid or expired API key");
    }

    if (response.status === 429) {
      const retryAfter = response.headers.get("Retry-After");
      options.onRateLimited?.(retryAfter ? parseInt(retryAfter, 10) : undefined);
      throw new Error("Rate limit exceeded");
    }

    if (!response.ok) {
      const error = (await response.json().catch(() => ({}))) as { message?: string };
      throw new Error(error.message || `Request failed: ${response.status}`);
    }

    if (response.status === 204) {
      return null as T;
    }

    return (await response.json()) as T;
  };

  return {
    getDevices: () => request("/v1/devices"),
    getDevice: (imei) => request(`/v1/devices/${imei}`),
    deleteDevice: (imei) => request(`/v1/devices/${imei}`, { method: "DELETE" }),

    getCommandStatus: (dispatchId) => request(`/v1/command/${dispatchId}/status`),
    retryCommand: (dispatchId) => request(`/v1/command/${dispatchId}/retry`, { method: "POST" }),
    cancelCommand: (dispatchId) => request(`/v1/command/${dispatchId}`, { method: "DELETE" }),

    getTelemetryHistory: (params) => {
      const searchParams = new URLSearchParams(
        Object.entries(params).map(([k, v]) => [k, String(v)]),
      );
      return request(`/v1/telemetry/history?${searchParams}`);
    },
    getLatestTelemetry: (deviceId) => request(`/v1/telemetry/latest/${deviceId}`),

    getUpdateStatus: () => request("/v1/updates/status"),
    getVersions: () => request("/v1/updates/versions"),

    request,
  };
}
