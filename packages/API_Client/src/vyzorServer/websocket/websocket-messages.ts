export type WSMessageType =
  | "AUTH"
  | "AUTH_ACK"
  | "TELEMETRY"
  | "COMMAND"
  | "COMMAND_ACK"
  | "EVENT"
  | "PING"
  | "PONG"
  | "SUBSCRIBE"
  | "SUBSCRIBE_ACK"
  | "UNSUBSCRIBE"
  | "UNSUBSCRIBE_ACK"
  | "ERROR";

export interface WSMessage<T = unknown> {
  type: WSMessageType;
  payload: T;
  timestamp?: number;
}

export interface WSAuthPayload {
  token: string;
}

export interface WSAuthAckPayload {
  success: boolean;
  error?: string;
}

export interface WSSubscribePayload {
  deviceId: string;
}

export interface WSSubscribeAckPayload {
  success: boolean;
  deviceId: string;
  error?: string;
}

export interface WSUnsubscribePayload {
  deviceId: string;
}

export interface WSUnsubscribeAckPayload {
  success: boolean;
  deviceId: string;
  error?: string;
}

export interface WSCommandPayload {
  dispatchId: string;
  command: string;
  parameters?: Record<string, unknown>;
}

export interface WSCommandAckPayload {
  dispatchId: string;
  success: boolean;
  payload?: Record<string, unknown>;
}

export interface WSPongPayload {
  timestamp: number;
}

export function createWSMessage<T>(type: WSMessageType, payload: T): WSMessage<T> {
  return {
    type,
    payload,
    timestamp: Date.now(),
  };
}

export function parseWSMessage(data: string): WSMessage | null {
  try {
    return JSON.parse(data) as WSMessage;
  } catch {
    return null;
  }
}
