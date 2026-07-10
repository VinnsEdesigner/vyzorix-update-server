/**
 * Commands Mappers
 * 
 * Transformations from raw API response to domain types.
 * Raw API uses snake_case, domain uses camelCase.
 */

import type {
  Command,
  CommandListItem,
  PresetCommandType,
  CommandStatus,
  CommandParams,
  CommandResult,
  SendCommandRequest,
} from "./commands-entity";

// ============================================================================
// Raw API Types (snake_case)
// ============================================================================

/**
 * Raw command params from API
 */
export interface RawCommandParams {
  active?: boolean;
  level?: string;
  [key: string]: unknown;
}

/**
 * Raw command result from API
 */
export interface RawCommandResult {
  success?: boolean;
  message?: string;
  output?: string;
  error?: string;
}

/**
 * Raw command from API
 */
export interface RawCommand {
  id?: string;
  dispatch_id?: string;
  type?: string;
  device_imei?: string;
  status?: string;
  params?: RawCommandParams;
  result?: RawCommandResult;
  created_at?: string | number;
  sent_at?: string | number;
  acknowledged_at?: string | number;
  completed_at?: string | number;
  error?: string;
}

/**
 * Raw command list item from API
 */
export interface RawCommandListItem {
  id?: string;
  dispatch_id?: string;
  type?: string;
  device_imei?: string;
  status?: string;
  created_at?: string | number;
}

// ============================================================================
// Transform Helpers
// ============================================================================

/**
 * Parse timestamp from various formats
 */
function parseTimestamp(value?: string | number | null): Date | undefined {
  if (!value) return undefined;
  
  if (typeof value === "number") {
    // Timestamps can be in seconds (1e9) or milliseconds (1e12)
    return new Date(value > 1e12 ? value : value * 1000);
  }
  
  return new Date(value);
}

// ============================================================================
// Transform Functions
// ============================================================================

/**
 * Transform raw command params to domain
 */
export function commandParamsFromRaw(raw?: RawCommandParams | null): CommandParams {
  if (!raw) return {};
  
  return {
    active: raw.active,
    level: raw.level,
    ...raw,
  };
}

/**
 * Transform raw command result to domain
 */
export function commandResultFromRaw(raw?: RawCommandResult | null): CommandResult | undefined {
  if (!raw) return undefined;
  
  return {
    success: raw.success ?? false,
    message: raw.message,
    output: raw.output,
    error: raw.error,
  };
}

/**
 * Transform raw command to domain
 */
export function commandFromRaw(raw: RawCommand): Command {
  return {
    id: raw.id ?? "",
    dispatchId: raw.dispatch_id ?? "",
    type: (raw.type as PresetCommandType) ?? "FORCE_SPEAKER",
    deviceImei: raw.device_imei ?? "",
    status: (raw.status as CommandStatus) ?? "pending",
    params: commandParamsFromRaw(raw.params),
    result: commandResultFromRaw(raw.result),
    createdAt: parseTimestamp(raw.created_at) ?? new Date(),
    sentAt: parseTimestamp(raw.sent_at),
    acknowledgedAt: parseTimestamp(raw.acknowledged_at),
    completedAt: parseTimestamp(raw.completed_at),
    error: raw.error,
  };
}

/**
 * Transform raw command list item to domain
 */
export function commandListItemFromRaw(raw: RawCommandListItem): CommandListItem {
  return {
    id: raw.id ?? "",
    dispatchId: raw.dispatch_id ?? "",
    type: (raw.type as PresetCommandType) ?? "FORCE_SPEAKER",
    deviceImei: raw.device_imei ?? "",
    status: (raw.status as CommandStatus) ?? "pending",
    createdAt: parseTimestamp(raw.created_at) ?? new Date(),
  };
}

// ============================================================================
// Array Transformers
// ============================================================================

/**
 * Transform array of raw commands
 */
export function commandsFromRaw(raw: RawCommand[]): Command[] {
  return raw.map(commandFromRaw);
}

/**
 * Transform array of raw command list items
 */
export function commandListItemsFromRaw(raw: RawCommandListItem[]): CommandListItem[] {
  return raw.map(commandListItemFromRaw);
}

// ============================================================================
// Request Transformers (Domain â API)
// ============================================================================

/**
 * Transform send command request to API format
 */
export function sendCommandRequestToRaw(request: SendCommandRequest): Record<string, unknown> {
  return {
    command_type: request.commandType,
    params: request.params,
  };
}
