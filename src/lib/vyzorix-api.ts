// Real client for the Go mock server in cmd/mockserver.
// All calls run from the browser. The mock server's WSS upgrader accepts any
// origin; for REST you may need to allow CORS or run the dashboard locally
// alongside the server.

export interface VersionManifest {
  version: string;
  version_code: number;
  apk_filename: string;
  apk_sha256: string;
  apk_size_bytes: number;
  release_notes: string;
}

export interface DeviceStatus {
  deviceId: string;
  online: boolean;
  lastSeen: number;
  appVersion: string;
  deviceClass: string;
}

export interface RegisterPayload {
  deviceId: string;
  firebaseInstallId: string;
  fcmToken: string;
  appVersion: string;
  deviceClass: string;
}

export interface RegisterResponse {
  deviceId: string;
  commandSecret: string;
  registeredAt: number;
  serverTime: number;
}

export interface CommandResponse {
  dispatchId: string;
  delivery: "sent" | "queued";
  serverTime: number;
}

// TelemetryFrame as described in docs/VyzorixUpdate_RepoTree.md and DOC_8.
export interface TelemetryFrame {
  type: "telemetry";
  deviceId?: string;
  uptime?: number;
  riskScore?: number;
  audioMode?: number;
  speakerOn?: boolean;
  activeDevice?: string;
  bufferLevel?: number;
  thermalTemp?: number;
  timestamp?: number | string;
}

function join(base: string, path: string) {
  return base.replace(/\/+$/, "") + path;
}

async function jsonOrThrow<T>(res: Response): Promise<T> {
  const text = await res.text();
  let body: unknown = text;
  try { body = text ? JSON.parse(text) : null; } catch {}
  if (!res.ok) {
    const msg = typeof body === "object" && body && "message" in body
      ? String((body as { message?: unknown }).message)
      : res.statusText || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return body as T;
}

export async function getHealth(serverUrl: string): Promise<{ ok: boolean }> {
  const res = await fetch(join(serverUrl, "/healthz"), { method: "GET" });
  return { ok: res.ok };
}

export async function getVersion(serverUrl: string): Promise<VersionManifest> {
  const res = await fetch(join(serverUrl, "/api/v1/version"), { method: "GET" });
  return jsonOrThrow<VersionManifest>(res);
}

export async function headApk(serverUrl: string, filename: string): Promise<number | null> {
  const res = await fetch(join(serverUrl, `/api/v1/apk/${filename}`), { method: "HEAD" });
  if (!res.ok) return null;
  const v = res.headers.get("content-length");
  return v ? Number(v) : null;
}

export async function getDeviceStatus(serverUrl: string, deviceId: string): Promise<DeviceStatus> {
  const res = await fetch(join(serverUrl, `/v1/device/${encodeURIComponent(deviceId)}/status`));
  return jsonOrThrow<DeviceStatus>(res);
}

export async function registerDevice(serverUrl: string, payload: RegisterPayload): Promise<RegisterResponse> {
  const res = await fetch(join(serverUrl, "/v1/device/register"), {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify(payload),
  });
  return jsonOrThrow<RegisterResponse>(res);
}

// Mock server runs with -strict-hmac=false by default, so an empty signature
// is accepted. Real server signing happens on Android and on the future
// production update server; the dashboard does not hold the per-device secret.
export async function dispatchCommand(
  serverUrl: string,
  deviceId: string,
  command: string,
  args?: Record<string, unknown>,
): Promise<CommandResponse> {
  const nonce = crypto.randomUUID().replace(/-/g, "");
  const timestamp = Date.now();
  const body = JSON.stringify({ command, args: args ?? {}, nonce, timestamp });
  const res = await fetch(join(serverUrl, `/v1/device/${encodeURIComponent(deviceId)}/command`), {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "X-Vyzorix-Nonce": nonce,
      "X-Vyzorix-Timestamp": String(timestamp),
      "X-Vyzorix-Signature": "",
    },
    body,
  });
  return jsonOrThrow<CommandResponse>(res);
}

export const COMMANDS: { id: string; label: string; description: string; danger?: boolean }[] = [
  { id: "FORCE_SPEAKER", label: "Force speaker", description: "Override route to builtin_speaker" },
  { id: "RESET_AUDIO_HAL", label: "Reset audio HAL", description: "Cycle the audio HAL pipeline" },
  { id: "TOGGLE_CAPTURE", label: "Toggle capture", description: "Restart playback capture engine" },
  { id: "REINIT_PROJECTION", label: "Reinit projection", description: "Re-acquire MediaProjection token" },
  { id: "REQUEST_STATUS", label: "Request status", description: "Force a telemetry frame now" },
  { id: "WAKE_UP_UPDATER", label: "Wake updater", description: "Trigger OTA version check" },
  { id: "DUMP_FLIGHT_DATA", label: "Dump flight data", description: "Persist last-known-state to storage", danger: true },
  { id: "ROTATE_KEYS", label: "Rotate keys", description: "Re-register and rotate command_secret", danger: true },
];