import type {
  WSTelemetry,
  WSEvent,
  WSCommandAck,
  WSEventType,
} from "./realtime-entity";

export interface RawWSTelemetry {
  type?: string;
  deviceId?: string;
  timestamp?: number;
  uptime?: number;
  riskScore?: number;
  audioMode?: number;
  bufferLevel?: number;
  thermalTemp?: number;
  speakerOn?: boolean;
  activeDevice?: string;
}

export interface RawWSEvent {
  id?: string;
  type?: string;
  deviceId?: string;
  timestamp?: number;
  data?: Record<string, unknown>;
}

export interface RawWSCommandAck {
  dispatchId?: string;
  deviceId?: string;
  command?: string;
  success?: boolean;
  timestamp?: number;
  payload?: {
    error?: string;
    detail?: string;
    [key: string]: unknown;
  };
}

export interface RawWSAuthResponse {
  success?: boolean;
  error?: string;
}

function parseTimestamp(value?: number | null): Date {
  if (!value) return new Date();
  return new Date(value > 1e12 ? value : value * 1000);
}

export function telemetryFromRaw(raw: RawWSTelemetry): WSTelemetry {
  return {
    deviceId: raw.deviceId ?? "",
    timestamp: parseTimestamp(raw.timestamp),
    uptime: raw.uptime,
    riskScore: raw.riskScore,
    audioMode: raw.audioMode,
    bufferLevel: raw.bufferLevel,
    thermalTemp: raw.thermalTemp,
    speakerOn: raw.speakerOn,
    activeDevice: raw.activeDevice,
  };
}

export function eventFromRaw(raw: RawWSEvent): WSEvent {
  return {
    id: raw.id ?? "",
    type: (raw.type as WSEventType) ?? "ERROR",
    deviceId: raw.deviceId ?? "",
    timestamp: parseTimestamp(raw.timestamp),
    data: raw.data,
  };
}

export function commandAckFromRaw(raw: RawWSCommandAck): WSCommandAck {
  return {
    dispatchId: raw.dispatchId ?? "",
    deviceId: raw.deviceId ?? "",
    command: raw.command ?? "",
    success: raw.success ?? false,
    timestamp: parseTimestamp(raw.timestamp),
    payload: raw.payload,
  };
}

export function authResponseFromRaw(raw: RawWSAuthResponse): { success: boolean; error?: string } {
  return {
    success: raw.success ?? false,
    error: raw.error,
  };
}
