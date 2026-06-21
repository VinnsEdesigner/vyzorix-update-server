// TypeScript types matching the GraphQL schema
// These mirror the Go domain types

// ============================================================
// SCALAR TYPES
// ============================================================

export type DateTime = string; // ISO 8601

export type JSON = Record<string, unknown> | unknown[] | string | number | boolean | null;

// ============================================================
// ENUMS
// ============================================================

export enum CommandStatus {
  PENDING = 'PENDING',
  DELIVERED = 'DELIVERED',
  FAILED = 'FAILED',
  CANCELLED = 'CANCELLED',
}

// ============================================================
// OBJECT TYPES
// ============================================================

export interface Device {
  id: string;
  name?: string;
  online: boolean;
  lastSeen?: DateTime;
  fcmToken?: string;
  version?: string;
  deviceClass?: string;
  createdAt?: DateTime;
}

export interface Command {
  dispatchId: string;
  commandId: string;
  deviceId: string;
  command: string;
  args?: JSON;
  status: CommandStatus;
  createdAt: DateTime;
  deliveredAt?: DateTime;
}

export interface CommandResult {
  dispatchId: string;
  commandId: string;
  deviceId: string;
  command: string;
  status: CommandStatus;
  createdAt: DateTime;
  deviceOnline: boolean;
}

export interface TelemetryEntry {
  id: string;
  deviceId: string;
  receivedAt: DateTime;
  riskScore?: number;
  bufferLevel?: number;
  thermalTemp?: number;
  audioMode?: number;
  speakerOn?: boolean;
  activeDevice?: string;
  uptime?: number;
  payload?: string;
}

export interface StatValue {
  avg: number;
  min: number;
  max: number;
}

export interface TelemetryStats {
  riskScore: StatValue;
  bufferLevel: StatValue;
  thermalTemp: StatValue;
}

export interface ConnectionStatus {
  deviceId: string;
  connected: boolean;
  connectedAt?: DateTime;
  lastMessageAt?: DateTime;
  uptimeSeconds?: number;
}

export interface Health {
  ok: boolean;
  serverTime: number;
  connectedDevices: number;
  version?: string;
}

// ============================================================
// QUERY RESPONSE TYPES
// ============================================================

export interface GetDeviceResponse {
  device: Device | null;
}

export interface GetDevicesResponse {
  devices: Device[];
}

export interface GetDeviceCountResponse {
  deviceCount: number;
}

export interface GetCommandResponse {
  command: Command | null;
}

export interface GetPendingCommandsResponse {
  pendingCommands: Command[];
}

export interface GetTelemetryHistoryResponse {
  telemetryHistory: TelemetryEntry[];
}

export interface GetLatestTelemetryResponse {
  latestTelemetry: TelemetryEntry | null;
}

export interface GetTelemetryStatsResponse {
  telemetryStats: TelemetryStats | null;
}

export interface GetConnectionStatusResponse {
  connectionStatus: ConnectionStatus | null;
}

export interface GetAllConnectionsResponse {
  allConnections: ConnectionStatus[];
}

export interface GetHealthResponse {
  health: Health;
}

// ============================================================
// MUTATION RESPONSE TYPES
// ============================================================

export interface UpdateFCMTokenResponse {
  updateFCMToken: {
    id: string;
    fcmToken: string;
  };
}

export interface DeleteDeviceResponse {
  deleteDevice: boolean;
}

export interface SendCommandResponse {
  sendCommand: CommandResult;
}

export interface RetryCommandResponse {
  retryCommand: Command;
}

export interface CancelCommandResponse {
  cancelCommand: boolean;
}

export interface DisconnectDeviceResponse {
  disconnectDevice: boolean;
}

// ============================================================
// DASHBOARD TYPES
// ============================================================

export interface DashboardData {
  devices: Device[];
  deviceCount: number;
  allConnections: ConnectionStatus[];
}

export interface DeviceDetail {
  device: Device | null;
  connectionStatus: ConnectionStatus | null;
  telemetryStats: TelemetryStats | null;
  pendingCommands: Command[];
}
