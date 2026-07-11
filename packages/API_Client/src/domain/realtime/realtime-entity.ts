export type WebSocketState = "connecting" | "connected" | "disconnected" | "reconnecting";

export type WSEventType =
  | "DEVICE_CONNECTED"
  | "DEVICE_DISCONNECTED"
  | "THRESHOLD_BREACH"
  | "COMMAND_DELIVERED"
  | "COMMAND_FAILED"
  | "ERROR";

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
  command: string;
  timestamp: Date;
  params?: Record<string, unknown>;
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
