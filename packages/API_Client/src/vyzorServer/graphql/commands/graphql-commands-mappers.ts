import type {
  Command,
  CommandListItem,
  CommandResult,
} from "../../../domain/commands";
import type {
  RawCommand,
  RawCommandListItem,
  RawCommandResult,
} from "./graphql-commands-types";

function parseTimestamp(value?: string | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value);
}

export function commandFromRaw(raw: RawCommand): Command {
  return {
    id: raw.commandId,
    dispatchId: raw.dispatchId,
    deviceId: raw.deviceId,
    command: raw.command,
    status: raw.status as Command["status"],
    failureReason: "",
    args: raw.args ?? {},
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
    updatedAt: parseTimestamp(raw.createdAt) ?? new Date(),
    deliveredAt: parseTimestamp(raw.deliveredAt),
    completedAt: undefined,
  };
}

export function commandListItemFromRaw(raw: RawCommandListItem): CommandListItem {
  return {
    id: raw.commandId,
    dispatchId: raw.dispatchId,
    deviceId: raw.deviceId,
    command: raw.command,
    status: raw.status as CommandListItem["status"],
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
  };
}

export function commandResultFromRaw(raw: RawCommandResult): CommandResult {
  return {
    success: raw.deviceOnline,
    message: raw.status,
    output: undefined,
    error: undefined,
  };
}
