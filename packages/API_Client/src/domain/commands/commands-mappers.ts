import type { RawPagination } from "../_shared";
import type {
  Command,
  CommandListItem,
  CommandStatus,
  SendCommandRequest,
} from "./commands-entity";

export interface RawCommand {
  id: string;
  dispatchId: string;
  deviceId: string;
  command: string;
  status: string;
  failureReason: string;
  args?: Record<string, unknown>;
  result?: {
    success?: boolean;
    message?: string;
    output?: string;
    error?: string;
  };
  createdAt: number;
  updatedAt: number;
  deliveredAt?: number;
  completedAt?: number;
}

export interface RawCommandListItem {
  id: string;
  dispatchId: string;
  deviceId: string;
  command: string;
  status: string;
  createdAt: number;
}

export interface RawCommandHistoryResult {
  commands: RawCommandListItem[];
  pagination: RawPagination;
}

function parseTimestamp(value?: number | null): Date | undefined {
  if (!value) return undefined;
  return new Date(value > 1e12 ? value : value * 1000);
}

export function commandFromRaw(raw: RawCommand): Command {
  return {
    id: raw.id,
    dispatchId: raw.dispatchId,
    deviceId: raw.deviceId,
    command: raw.command,
    status: (raw.status as CommandStatus) ?? "pending",
    failureReason: raw.failureReason,
    args: raw.args ?? {},
    result: raw.result ? {
      success: raw.result.success ?? false,
      message: raw.result.message,
      output: raw.result.output,
      error: raw.result.error,
    } : undefined,
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
    updatedAt: parseTimestamp(raw.updatedAt) ?? new Date(),
    deliveredAt: parseTimestamp(raw.deliveredAt),
    completedAt: parseTimestamp(raw.completedAt),
  };
}

export function commandListItemFromRaw(raw: RawCommandListItem): CommandListItem {
  return {
    id: raw.id,
    dispatchId: raw.dispatchId,
    deviceId: raw.deviceId,
    command: raw.command,
    status: (raw.status as CommandStatus) ?? "pending",
    createdAt: parseTimestamp(raw.createdAt) ?? new Date(),
  };
}

export function sendCommandRequestToRaw(request: SendCommandRequest): { command: string; params: Record<string, unknown> } {
  return {
    command: request.commandType,
    params: request.params ?? {},
  };
}
