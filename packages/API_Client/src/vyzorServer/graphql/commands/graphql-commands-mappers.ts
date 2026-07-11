import type {
  Command,
  CommandListItem,
  CommandResult,
} from "@/domain/commands";
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
    dispatchId: raw.dispatchId,
    commandId: raw.commandId,
    imei: raw.imei,
    commandType: raw.commandType,
    status: raw.status,
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
    sentAt: parseTimestamp(raw.sentAt),
    acknowledgedAt: parseTimestamp(raw.acknowledgedAt),
    completedAt: parseTimestamp(raw.completedAt),
    result: raw.result ? commandResultFromRaw(raw.result) : undefined,
  };
}

export function commandListItemFromRaw(raw: RawCommandListItem): CommandListItem {
  return {
    dispatchId: raw.dispatchId,
    imei: raw.imei,
    commandType: raw.commandType,
    status: raw.status,
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
  };
}

export function commandResultFromRaw(raw: RawCommandResult): CommandResult {
  return {
    success: raw.success,
    message: raw.message,
    output: raw.output,
    error: raw.error,
  };
}
