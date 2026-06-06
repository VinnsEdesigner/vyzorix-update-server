// Browser client for the Vyzorix Go update server. It targets the real
// Render-backed Phase 1.5 server while keeping the same Android-facing paths
// as the Phase 1 mock server.
import { logger } from "@/lib/logger";

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
  firebaseInstallId?: string;
  fcmToken?: string;
}

export interface DashboardDevicesResponse {
  devices: DeviceStatus[];
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

function dashboardHeaders(token?: string): Record<string, string> {
  return token ? { Authorization: `Bearer ${token}`, "X-Vyzorix-Token": token } : {};
}

async function jsonOrThrow<T>(res: Response): Promise<T> {
  const text = await res.text();
  let body: unknown = text;
  try {
    body = text ? JSON.parse(text) : null;
  } catch {
    // ignore parse error, use raw text
  }
  if (!res.ok) {
    const msg =
      typeof body === "object" && body && "message" in body
        ? String((body as { message?: unknown }).message)
        : res.statusText || `HTTP ${res.status}`;
    throw new Error(msg);
  }
  return body as T;
}

export async function getHealth(serverUrl: string): Promise<{ ok: boolean }> {
  const t0 = Date.now();
  try {
    const res = await fetch(join(serverUrl, "/healthz"), { method: "GET" });
    logger.debug("api", `GET /healthz · ${res.status} · ${Date.now() - t0}ms`);
    return { ok: res.ok };
  } catch (e) {
    logger.warn("api", `GET /healthz · failed · ${e instanceof Error ? e.message : String(e)}`);
    return { ok: false };
  }
}

export async function getVersion(serverUrl: string): Promise<VersionManifest> {
  const t0 = Date.now();
  try {
    const res = await fetch(join(serverUrl, "/api/v1/version"), { method: "GET" });
    const body = await jsonOrThrow<VersionManifest>(res);
    logger.info(
      "update",
      `manifest v${body.version} (code ${body.version_code}) · ${Date.now() - t0}ms`,
    );
    return body;
  } catch (e) {
    logger.error(
      "update",
      `version.json fetch failed · ${e instanceof Error ? e.message : String(e)}`,
    );
    throw e;
  }
}

export async function headApk(serverUrl: string, filename: string): Promise<number | null> {
  try {
    const res = await fetch(join(serverUrl, `/api/v1/apk/${filename}`), { method: "HEAD" });
    if (!res.ok) {
      logger.warn("update", `HEAD apk ${filename} · ${res.status}`);
      return null;
    }
    const v = res.headers.get("content-length");
    return v ? Number(v) : null;
  } catch (e) {
    logger.warn("update", `HEAD apk failed · ${e instanceof Error ? e.message : String(e)}`);
    return null;
  }
}

export async function getDeviceStatus(serverUrl: string, deviceId: string): Promise<DeviceStatus> {
  const res = await fetch(join(serverUrl, `/v1/device/${encodeURIComponent(deviceId)}/status`));
  return jsonOrThrow<DeviceStatus>(res);
}

export async function getDashboardDevices(
  serverUrl: string,
  dashboardToken?: string,
): Promise<DeviceStatus[]> {
  const res = await fetch(join(serverUrl, "/v1/dashboard/devices"), {
    headers: dashboardHeaders(dashboardToken),
  });
  const body = await jsonOrThrow<DashboardDevicesResponse>(res);
  return body.devices;
}

export async function registerDevice(
  serverUrl: string,
  payload: RegisterPayload,
): Promise<RegisterResponse> {
  logger.info("device", `register → ${payload.deviceId}`);
  try {
    const res = await fetch(join(serverUrl, "/v1/device/register"), {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify(payload),
    });
    const body = await jsonOrThrow<RegisterResponse>(res);
    logger.info("device", `registered · serverTime=${new Date(body.serverTime).toISOString()}`);
    return body;
  } catch (e) {
    logger.error("device", `register failed · ${e instanceof Error ? e.message : String(e)}`);
    throw e;
  }
}

// Development can run with ENFORCE_HMAC=false, so an empty device signature is
// accepted. Production dashboard commands should carry TOKEN_SECRET through
// Authorization/X-Vyzorix-Token; Android-originated requests still use per-device HMAC.
export async function dispatchCommand(
  serverUrl: string,
  deviceId: string,
  command: string,
  args?: Record<string, unknown>,
  dashboardToken?: string,
  strictHmac?: boolean,
): Promise<CommandResponse> {
  const nonce = crypto.randomUUID().replace(/-/g, "");
  const timestamp = Date.now();
  const body = JSON.stringify({ command, args: args ?? {}, nonce, timestamp });
  logger.info("command", `→ ${command}`, { nonce: nonce.slice(0, 8), deviceId });
  try {
    const headers: Record<string, string> = {
      "Content-Type": "application/json",
      "X-Vyzorix-Nonce": nonce,
      "X-Vyzorix-Timestamp": String(timestamp),
      ...dashboardHeaders(dashboardToken),
    };
    // strictHmac is a client-side dev toggle to simulate production HMAC enforcement.
    // When true, the server validates X-Vyzorix-Signature (per-device HMAC).
    // The actual signature is computed client-side using the device's command_secret,
    // which is returned on registration. In production the Android daemon computes
    // this server-side; here we set an empty signature and let the server decide
    // whether to enforce based on its ENFORCE_HMAC setting.
    if (strictHmac) {
      headers["X-Vyzorix-Signature"] = "";
    }
    const res = await fetch(join(serverUrl, `/v1/device/${encodeURIComponent(deviceId)}/command`), {
      method: "POST",
      headers,
      body,
    });
    const out = await jsonOrThrow<CommandResponse>(res);
    logger.info("command", `← ${command} · ${out.delivery} · ${out.dispatchId.slice(0, 8)}`);
    return out;
  } catch (e) {
    logger.error("command", `${command} failed · ${e instanceof Error ? e.message : String(e)}`);
    throw e;
  }
}

export const COMMANDS: { id: string; label: string; description: string; danger?: boolean }[] = [
  { id: "FORCE_SPEAKER", label: "Force speaker", description: "Override route to builtin_speaker" },
  { id: "RESET_AUDIO_HAL", label: "Reset audio HAL", description: "Cycle the audio HAL pipeline" },
  { id: "TOGGLE_CAPTURE", label: "Toggle capture", description: "Restart playback capture engine" },
  {
    id: "REINIT_PROJECTION",
    label: "Reinit projection",
    description: "Re-acquire MediaProjection token",
  },
  { id: "REQUEST_STATUS", label: "Request status", description: "Force a telemetry frame now" },
  { id: "WAKE_UP_UPDATER", label: "Wake updater", description: "Trigger OTA version check" },
  {
    id: "DUMP_FLIGHT_DATA",
    label: "Dump flight data",
    description: "Persist last-known-state to storage",
    danger: true,
  },
  {
    id: "ROTATE_KEYS",
    label: "Rotate keys",
    description: "Re-register and rotate command_secret",
    danger: true,
  },
];
