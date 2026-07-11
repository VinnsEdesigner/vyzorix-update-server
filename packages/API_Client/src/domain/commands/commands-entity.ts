export type PresetCommandType =
  | "FORCE_SPEAKER"
  | "RESET_AUDIO_HAL"
  | "TOGGLE_CAPTURE"
  | "REINIT_PROJECTION"
  | "DUMP_FLIGHT_DATA"
  | "UPLOAD_CRASH_ZIP"
  | "SET_LOG_LEVEL"
  | "WAKE_UP_UPDATER";

export type CommandStatus =
  | "pending"
  | "delivered"
  | "completed"
  | "failed"
  | "cancelled";

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

export interface Command {
  id: string;
  dispatchId: string;
  deviceId: string;
  command: string;
  status: CommandStatus;
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
  status: CommandStatus;
  createdAt: Date;
}

export interface SendCommandRequest {
  imei: string;
  commandType: PresetCommandType;
  params?: CommandParams;
}

export const PRESET_COMMANDS: Record<PresetCommandType, { label: string; description: string }> = {
  FORCE_SPEAKER: { label: "Force Speaker", description: "Force speaker on with reassertion loop" },
  RESET_AUDIO_HAL: { label: "Reset Audio HAL", description: "Soft HAL reset via BT stream cycling" },
  TOGGLE_CAPTURE: { label: "Toggle Capture", description: "Start/stop AudioRecord read loops" },
  REINIT_PROJECTION: { label: "Reinit Projection", description: "Re-initiate media projection via notification" },
  DUMP_FLIGHT_DATA: { label: "Dump Flight Data", description: "Gather local metrics → JSON postback" },
  UPLOAD_CRASH_ZIP: { label: "Upload Crash Zip", description: "Zip SQLite logs → POST binary" },
  SET_LOG_LEVEL: { label: "Set Log Level", description: "Dynamically modify Logger minLogLevel" },
  WAKE_UP_UPDATER: { label: "Wake Up Updater", description: "Override WorkManager delays → run UpdateChecker" },
};

export const COMMAND_STATUSES: Record<CommandStatus, { label: string; terminal: boolean }> = {
  pending: { label: "Pending", terminal: false },
  delivered: { label: "Delivered", terminal: false },
  completed: { label: "Completed", terminal: true },
  failed: { label: "Failed", terminal: true },
  cancelled: { label: "Cancelled", terminal: true },
};

export function isCommandTerminal(status: CommandStatus): boolean {
  return COMMAND_STATUSES[status].terminal;
}

export function getCommandLabel(type: PresetCommandType): string {
  return PRESET_COMMANDS[type].label;
}

export function getStatusLabel(status: CommandStatus): string {
  return COMMAND_STATUSES[status].label;
}
