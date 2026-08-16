export type WebSocketState = "connecting" | "connected" | "disconnected" | "reconnecting";

export type WSEventType =
  | "DEVICE_CONNECTED"
  | "DEVICE_DISCONNECTED"
  | "THRESHOLD_BREACH"
  | "COMMAND_DELIVERED"
  | "COMMAND_FAILED"
  | "ERROR";

/**
 * WSCommandType — the set of commands the dashboard can dispatch to a device
 * over the realtime channel. Mirrors the REST PresetCommandType values (see
 * domain/commands/commands-entity.ts) so the WS dispatch path and the REST
 * dispatch path stay in lockstep.
 */
export type WSCommandType =
  | "FORCE_SPEAKER"
  | "RESET_AUDIO_HAL"
  | "TOGGLE_CAPTURE"
  | "REINIT_PROJECTION"
  | "DUMP_FLIGHT_DATA"
  | "UPLOAD_CRASH_ZIP"
  | "SET_LOG_LEVEL"
  | "WAKE_UP_UPDATER";

export type WSCommandPriority = "high" | "normal" | "low";

export interface WSTelemetry {
  deviceId: string;
  timestamp: Date;
  uptime?: number;
  riskScore?: number;
  audioMode?: number;
  bufferLevel?: number;
  thermalTemp?: number;
  speakerOn?: boolean;
  activeDevice?: string;
}

export interface WSEvent {
  id: string;
  type: WSEventType;
  deviceId: string;
  timestamp: Date;
  data?: Record<string, unknown>;
}

export interface WSCommand {
  dispatchId: string;
  deviceImei: string;
  command: WSCommandType;
  parameters: Record<string, unknown>;
  priority: WSCommandPriority;
  timestamp: Date;
  nonce?: string;
  hmac?: string;
}

export interface WSCommandAck {
  dispatchId: string;
  deviceId: string;
  command: string;
  success: boolean;
  timestamp: Date;
  payload?: {
    error?: string;
    detail?: string;
    [key: string]: unknown;
  };
}

export interface WSAuthResponse {
  success: boolean;
  error?: string;
}

export const WSEVENT_LABELS: Record<WSEventType, string> = {
  DEVICE_CONNECTED: "Device Connected",
  DEVICE_DISCONNECTED: "Device Disconnected",
  THRESHOLD_BREACH: "Threshold Breach",
  COMMAND_DELIVERED: "Command Delivered",
  COMMAND_FAILED: "Command Failed",
  ERROR: "Error",
};

export function getEventLabel(type: WSEventType): string {
  return WSEVENT_LABELS[type] ?? type;
}
