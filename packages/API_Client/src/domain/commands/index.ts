/**
 * Commands Domain Index
 * 
 * Re-exports all commands domain types, mappers, and validators.
 */

// Types
export type {
  PresetCommandType,
  PresetCommandInfo,
  CommandStatus,
  CommandStatusInfo,
  CommandParams,
  CommandResult,
  Command,
  CommandListItem,
  SendCommandRequest,
} from "./commands-entity";

export {
  PRESET_COMMANDS,
  COMMAND_STATUSES,
  isCommandTerminal,
  isCommandSuccess,
  requiresConfirmation,
  isDestructive,
  getCommandLabel,
  getStatusLabel,
  getPresetCommandInfo,
  getCommandDuration,
} from "./commands-entity";

// Mappers
export type {
  RawCommand,
  RawCommandListItem,
  RawCommandParams,
  RawCommandResult,
} from "./commands-mappers";

export {
  commandParamsFromRaw,
  commandResultFromRaw,
  commandFromRaw,
  commandListItemFromRaw,
  commandsFromRaw,
  commandListItemsFromRaw,
  sendCommandRequestToRaw,
} from "./commands-mappers";

// Validators
export type { ValidationResult } from "./commands-validators";

export {
  validatePresetCommandType,
  validateIMEI,
  validateLogLevel,
  validateToggleCaptureParams,
  validatePresetCommandParams,
  validateSendCommand,
} from "./commands-validators";
