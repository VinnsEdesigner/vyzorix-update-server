import type { WSTelemetry, WSEvent, WSCommand } from "./realtime-entity";

export function validateTelemetry(data: unknown): data is WSTelemetry {
  if (!data || typeof data !== "object") return false;
  const t = data as Record<string, unknown>;
  return typeof t.deviceId === "string";
}

export function validateEvent(data: unknown): data is WSEvent {
  if (!data || typeof data !== "object") return false;
  const e = data as Record<string, unknown>;
  return typeof e.id === "string" && typeof e.type === "string";
}

export function validateCommand(data: unknown): data is WSCommand {
  if (!data || typeof data !== "object") return false;
  const c = data as Record<string, unknown>;
  return (
    typeof c.dispatchId === "string" &&
    typeof c.deviceImei === "string" &&
    typeof c.command === "string"
  );
}
