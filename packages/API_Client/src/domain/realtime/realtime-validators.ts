import type { WSTelemetry, WSEvent, WSCommand, WSEventType, WSCommandType, WSCommandPriority } from "./realtime-entity";

const WS_EVENT_TYPES: readonly WSEventType[] = [
  "DEVICE_CONNECTED",
  "DEVICE_DISCONNECTED",
  "THRESHOLD_BREACH",
  "COMMAND_DELIVERED",
  "COMMAND_FAILED",
  "ERROR",
];

const WS_COMMAND_TYPES: readonly WSCommandType[] = [
  "FORCE_SPEAKER",
  "RESET_AUDIO_HAL",
  "TOGGLE_CAPTURE",
  "REINIT_PROJECTION",
  "DUMP_FLIGHT_DATA",
  "UPLOAD_CRASH_ZIP",
  "SET_LOG_LEVEL",
  "WAKE_UP_UPDATER",
];

const WS_COMMAND_PRIORITIES: readonly WSCommandPriority[] = ["high", "normal", "low"];

function isNumber(value: unknown): value is number {
  return typeof value === "number" && !Number.isNaN(value);
}

function inRange(value: number, min: number, max: number): boolean {
  return value >= min && value <= max;
}

/**
 * Validate a telemetry frame (spec §6.3).
 *
 * `deviceId` is required. When the optional numeric metrics are present they
 * must be within their valid ranges: riskScore and bufferLevel are 0–100
 * (bufferLevel may legitimately be absent). This guards the UI against
 * malformed/attacker-crafted frames without rejecting frames that simply
 * omit optional fields.
 */
export function validateTelemetry(data: unknown): data is WSTelemetry {
  if (!data || typeof data !== "object") return false;
  const t = data as Record<string, unknown>;
  if (typeof t.deviceId !== "string" || t.deviceId.length === 0) return false;
  if (t.riskScore !== undefined && !(isNumber(t.riskScore) && inRange(t.riskScore, 0, 100))) return false;
  if (t.bufferLevel !== undefined && !(isNumber(t.bufferLevel) && inRange(t.bufferLevel, 0, 100))) return false;
  if (t.thermalTemp !== undefined && !isNumber(t.thermalTemp)) return false;
  if (t.uptime !== undefined && !isNumber(t.uptime)) return false;
  if (t.audioMode !== undefined && !isNumber(t.audioMode)) return false;
  return true;
}

export function validateEvent(data: unknown): data is WSEvent {
  if (!data || typeof data !== "object") return false;
  const e = data as Record<string, unknown>;
  if (typeof e.id !== "string" || e.id.length === 0) return false;
  if (typeof e.type !== "string" || !WS_EVENT_TYPES.includes(e.type as WSEventType)) return false;
  if (typeof e.deviceId !== "string") return false;
  return true;
}

export function validateCommand(data: unknown): data is WSCommand {
  if (!data || typeof data !== "object") return false;
  const c = data as Record<string, unknown>;
  if (typeof c.dispatchId !== "string" || c.dispatchId.length === 0) return false;
  if (typeof c.deviceImei !== "string" || c.deviceImei.length === 0) return false;
  if (typeof c.command !== "string" || !WS_COMMAND_TYPES.includes(c.command as WSCommandType)) return false;
  if (c.priority !== undefined && !WS_COMMAND_PRIORITIES.includes(c.priority as WSCommandPriority)) return false;
  return true;
}
