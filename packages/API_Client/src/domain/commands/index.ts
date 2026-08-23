// Commands domain — generated types + hand-rolled business rules.
// Preset command validation, log level validation, and per-command parameter
// validation are genuine business rules not expressible in OpenAPI.
import type {
  CommandRequest,
  CommandDispatchResult,
  CommandStatus,
  CommandRetryResult,
  CommandCancelResult,
  CommandPendingResult,
  CommandResponse,
  CommandHistoryResult,
  CommandHistoryEntry,
  CommandConfirmRequest,
  CommandConfirmResult,
} from '../../generated/vyzorixUpdateServerAPI.schemas';

export type {
  CommandRequest,
  CommandDispatchResult,
  CommandStatus,
  CommandRetryResult,
  CommandCancelResult,
  CommandPendingResult,
  CommandResponse,
  CommandHistoryResult,
  CommandHistoryEntry,
  CommandConfirmRequest,
  CommandConfirmResult,
};

// ---- Constants (hand-rolled, not in OpenAPI) ----

export type PresetCommandType =
  | 'FORCE_SPEAKER'
  | 'RESET_AUDIO_HAL'
  | 'TOGGLE_CAPTURE'
  | 'REINIT_PROJECTION'
  | 'DUMP_FLIGHT_DATA'
  | 'UPLOAD_CRASH_ZIP'
  | 'SET_LOG_LEVEL'
  | 'WAKE_UP_UPDATER';

export type CommandStatusType =
  | 'pending'
  | 'delivered'
  | 'completed'
  | 'failed'
  | 'cancelled';

export interface CommandParams {
  active?: boolean;
  level?: string;
  [key: string]: unknown;
}

export interface CommandResult {
  success: boolean;
  message?: string;
  output?: string;
  error?: string;
}

const VALID_PRESET_TYPES: PresetCommandType[] = [
  'FORCE_SPEAKER',
  'RESET_AUDIO_HAL',
  'TOGGLE_CAPTURE',
  'REINIT_PROJECTION',
  'DUMP_FLIGHT_DATA',
  'UPLOAD_CRASH_ZIP',
  'SET_LOG_LEVEL',
  'WAKE_UP_UPDATER',
];

const VALID_LOG_LEVELS = ['debug', 'info', 'warn', 'error', 'verbose', 'assert'];

// ---- Validators (business rules) ----

export function validatePresetCommandType(type: string): string | null {
  if (!type) return 'Command type is required';
  if (!VALID_PRESET_TYPES.includes(type as PresetCommandType)) return `Invalid preset command type: ${type}`;
  return null;
}

export function validateLogLevel(level: string): string | null {
  if (!level) return 'Log level is required for SET_LOG_LEVEL';
  if (!VALID_LOG_LEVELS.includes(level.toLowerCase())) return `Invalid log level: ${level}. Valid levels: ${VALID_LOG_LEVELS.join(', ')}`;
  return null;
}

export function validatePresetCommandParams(type: PresetCommandType, params: CommandParams): string | null {
  switch (type) {
    case 'TOGGLE_CAPTURE':
      if (typeof params.active !== 'boolean') return "'active' parameter must be a boolean for TOGGLE_CAPTURE";
      return null;
    case 'SET_LOG_LEVEL':
      if (!params.level) return 'Log level is required for SET_LOG_LEVEL';
      return validateLogLevel(String(params.level));
    default:
      return null;
  }
}

export function validateSendCommand(imei: string, commandType: string, params?: CommandParams): { isValid: boolean; errors: Record<string, string[]> } {
  const errors: Record<string, string[]> = {};
  if (!/^\d{15}$/.test(imei)) errors.imei = ['IMEI must be 15 digits'];
  const typeError = validatePresetCommandType(commandType);
  if (typeError) errors.type = [typeError];
  if (!typeError) {
    const paramError = validatePresetCommandParams(commandType as PresetCommandType, params ?? {});
    if (paramError) errors.params = [paramError];
  }
  return { isValid: Object.keys(errors).length === 0, errors };
}

export function isCommandTerminal(status: string): boolean {
  return ['completed', 'failed', 'cancelled'].includes(status);
}

// ---- Hook-facing domain types (not in OpenAPI; GraphQL fallback shape) ----

export interface Command {
  id: string;
  dispatchId: string;
  deviceId: string;
  command: string;
  status: CommandStatusType;
  failureReason: string;
  args: CommandParams;
  result?: CommandResult;
  createdAt: Date;
  updatedAt: Date;
  deliveredAt?: Date;
  completedAt?: Date;
}

export interface CommandListItem {
  id: string;
  dispatchId: string;
  deviceId: string;
  command: string;
  status: CommandStatusType;
  createdAt: Date;
}

