/**
 * Commands Domain Types
 * 
 * Domain types for device commands.
 * Pure TypeScript - no external imports.
 * 
 * Commands are HMAC-SHA256 signed for security.
 * See COMMAND_SECURITY.md for signing specification.
 */

// ============================================================================
// Preset Commands (Device-Specific)
// ============================================================================

/**
 * Available preset command types for Android device control
 */
export type PresetCommandType = 
  | "FORCE_SPEAKER"
  | "RESET_AUDIO_HAL"
  | "TOGGLE_CAPTURE"
  | "REINIT_PROJECTION"
  | "DUMP_FLIGHT_DATA"
  | "UPLOAD_CRASH_ZIP"
  | "SET_LOG_LEVEL"
  | "WAKE_UP_UPDATER";

/**
 * Preset command metadata
 */
export interface PresetCommandInfo {
  type: PresetCommandType;
  label: string;
  description: string;
  params?: Record<string, "boolean" | "string" | "number">;
  requiresConfirmation: boolean;
  destructive: boolean;
}

/**
 * All preset commands with metadata
 */
export const PRESET_COMMANDS: Record<PresetCommandType, PresetCommandInfo> = {
  FORCE_SPEAKER: {
    type: "FORCE_SPEAKER",
    label: "Force Speaker",
    description: "Force speaker on with reassertion loop",
    requiresConfirmation: false,
    destructive: false,
  },
  RESET_AUDIO_HAL: {
    type: "RESET_AUDIO_HAL",
    label: "Reset Audio HAL",
    description: "Soft HAL reset via BT stream cycling",
    requiresConfirmation: false,
    destructive: false,
  },
  TOGGLE_CAPTURE: {
    type: "TOGGLE_CAPTURE",
    label: "Toggle Capture",
    description: "Start/stop AudioRecord read loops",
    params: { active: "boolean" },
    requiresConfirmation: false,
    destructive: false,
  },
  REINIT_PROJECTION: {
    type: "REINIT_PROJECTION",
    label: "Reinit Projection",
    description: "Re-initiate media projection via notification",
    requiresConfirmation: false,
    destructive: false,
  },
  DUMP_FLIGHT_DATA: {
    type: "DUMP_FLIGHT_DATA",
    label: "Dump Flight Data",
    description: "Gather local metrics â JSON postback",
    requiresConfirmation: false,
    destructive: false,
  },
  UPLOAD_CRASH_ZIP: {
    type: "UPLOAD_CRASH_ZIP",
    label: "Upload Crash Zip",
    description: "Zip SQLite logs â POST binary",
    requiresConfirmation: false,
    destructive: false,
  },
  SET_LOG_LEVEL: {
    type: "SET_LOG_LEVEL",
    label: "Set Log Level",
    description: "Dynamically modify Logger minLogLevel",
    params: { level: "string" },
    requiresConfirmation: false,
    destructive: false,
  },
  WAKE_UP_UPDATER: {
    type: "WAKE_UP_UPDATER",
    label: "Wake Up Updater",
    description: "Override WorkManager delays â run UpdateChecker",
    requiresConfirmation: false,
    destructive: false,
  },
} as const;

// ============================================================================
// Command Status
// ============================================================================

/**
 * Command execution status
 */
export type CommandStatus = 
  | "pending"
  | "sent"
  | "acknowledged"
  | "completed"
  | "failed"
  | "timeout"
  | "cancelled";

/**
 * Status metadata
 */
export interface CommandStatusInfo {
  status: CommandStatus;
  label: string;
  terminal: boolean;
  success: boolean;
}

/**
 * All command statuses with metadata
 */
export const COMMAND_STATUSES: Record<CommandStatus, CommandStatusInfo> = {
  pending: {
    status: "pending",
    label: "Pending",
    terminal: false,
    success: false,
  },
  sent: {
    status: "sent",
    label: "Sent",
    terminal: false,
    success: false,
  },
  acknowledged: {
    status: "acknowledged",
    label: "Acknowledged",
    terminal: false,
    success: false,
  },
  completed: {
    status: "completed",
    label: "Completed",
    terminal: true,
    success: true,
  },
  failed: {
    status: "failed",
    label: "Failed",
    terminal: true,
    success: false,
  },
  timeout: {
    status: "timeout",
    label: "Timeout",
    terminal: true,
    success: false,
  },
  cancelled: {
    status: "cancelled",
    label: "Cancelled",
    terminal: true,
    success: false,
  },
} as const;

// ============================================================================
// Command Entity
// ============================================================================

/**
 * Command parameters (varies by preset type)
 */
export interface CommandParams {
  // For TOGGLE_CAPTURE
  active?: boolean;
  // For SET_LOG_LEVEL
  level?: string;
  // Generic extra params
  [key: string]: unknown;
}

/**
 * Command execution result
 */
export interface CommandResult {
  success: boolean;
  message?: string;
  output?: string;
  error?: string;
}

/**
 * Device command
 */
export interface Command {
  id: string;
  dispatchId: string;
  type: PresetCommandType;
  deviceImei: string;
  status: CommandStatus;
  params: CommandParams;
  result?: CommandResult;
  createdAt: Date;
  sentAt?: Date;
  acknowledgedAt?: Date;
  completedAt?: Date;
  error?: string;
}

// ============================================================================
// Command List Item
// ============================================================================

/**
 * Lightweight command for list views
 */
export interface CommandListItem {
  id: string;
  dispatchId: string;
  type: PresetCommandType;
  deviceImei: string;
  status: CommandStatus;
  createdAt: Date;
}

// ============================================================================
// Command Request
// ============================================================================

/**
 * Request to send a command
 */
export interface SendCommandRequest {
  imei: string;
  commandType: PresetCommandType;
  params?: CommandParams;
}

// ============================================================================
// Helper Functions
// ============================================================================

/**
 * Check if command status is terminal
 */
export function isCommandTerminal(command: Command | CommandListItem): boolean {
  return COMMAND_STATUSES[command.status].terminal;
}

/**
 * Check if command was successful
 */
export function isCommandSuccess(command: Command | CommandListItem): boolean {
  return COMMAND_STATUSES[command.status].success;
}

/**
 * Check if command requires user confirmation
 */
export function requiresConfirmation(commandType: PresetCommandType): boolean {
  return PRESET_COMMANDS[commandType].requiresConfirmation;
}

/**
 * Check if command is destructive
 */
export function isDestructive(commandType: PresetCommandType): boolean {
  return PRESET_COMMANDS[commandType].destructive;
}

/**
 * Get command type label
 */
export function getCommandLabel(type: PresetCommandType): string {
  return PRESET_COMMANDS[type].label;
}

/**
 * Get command status label
 */
export function getStatusLabel(status: CommandStatus): string {
  return COMMAND_STATUSES[status].label;
}

/**
 * Get preset command info
 */
export function getPresetCommandInfo(type: PresetCommandType): PresetCommandInfo {
  return PRESET_COMMANDS[type];
}

/**
 * Calculate command duration in seconds
 */
export function getCommandDuration(command: Command): number | null {
  if (!command.completedAt) return null;
  
  const start = command.acknowledgedAt ?? command.sentAt ?? command.createdAt;
  const end = command.completedAt;
  
  return Math.round((end.getTime() - start.getTime()) / 1000);
}
